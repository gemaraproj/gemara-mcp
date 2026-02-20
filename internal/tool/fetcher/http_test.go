// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		resp      *http.Response
		err       error
		wantRetry bool
		wantErr   bool
	}{
		{
			name:      "retry on 500",
			resp:      &http.Response{StatusCode: http.StatusInternalServerError},
			wantRetry: true,
		},
		{
			name:      "retry on 502",
			resp:      &http.Response{StatusCode: http.StatusBadGateway},
			wantRetry: true,
		},
		{
			name:      "retry on 503",
			resp:      &http.Response{StatusCode: http.StatusServiceUnavailable},
			wantRetry: true,
		},
		{
			name:      "retry on 429",
			resp:      &http.Response{StatusCode: http.StatusTooManyRequests},
			wantRetry: true,
		},
		{
			name:      "retry on 408",
			resp:      &http.Response{StatusCode: http.StatusRequestTimeout},
			wantRetry: true,
		},
		{
			name:      "no retry on 404",
			resp:      &http.Response{StatusCode: http.StatusNotFound},
			wantRetry: false,
		},
		{
			name:      "no retry on 200",
			resp:      &http.Response{StatusCode: http.StatusOK},
			wantRetry: false,
		},
		{
			name:      "no retry on 401",
			resp:      &http.Response{StatusCode: http.StatusUnauthorized},
			wantRetry: false,
		},
		{
			name:      "retry on network timeout",
			err:       &net.DNSError{IsTimeout: true},
			wantRetry: true,
		},
		{
			name:    "no retry on non-timeout error, propagates error",
			err:     io.ErrUnexpectedEOF,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retry, err := isRetryable(tt.resp, tt.err)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRetry, retry)
		})
	}
}

func TestRetryAfter(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		want time.Duration
	}{
		{
			name: "parses valid header",
			resp: &http.Response{
				Header: http.Header{"Retry-After": []string{"5"}},
			},
			want: 5 * time.Second,
		},
		{
			name: "returns zero for missing header",
			resp: &http.Response{Header: http.Header{}},
			want: 0,
		},
		{
			name: "returns zero for nil response",
			resp: nil,
			want: 0,
		},
		{
			name: "returns zero for non-numeric value",
			resp: &http.Response{
				Header: http.Header{"Retry-After": []string{"not-a-number"}},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, retryAfter(tt.resp))
		})
	}
}

func TestRetryTransport_NoRetryNeeded(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	transport := &retryTransport{maxRetry: 5, wait: time.Millisecond}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := transport.RoundTrip(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), calls.Load(), "should only make one request")
	_ = resp.Body.Close()
}

func TestRetryTransport_RetriesThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "success")
	}))
	defer srv.Close()

	transport := &retryTransport{maxRetry: 5, wait: time.Millisecond}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := transport.RoundTrip(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), calls.Load(), "should retry twice then succeed")
	_ = resp.Body.Close()
}

func TestRetryTransport_ExhaustsRetries(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	transport := &retryTransport{maxRetry: 2, wait: time.Millisecond}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := transport.RoundTrip(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode,
		"should return last response after exhausting retries")
	assert.Equal(t, int32(3), calls.Load(), "initial + 2 retries")
	_ = resp.Body.Close()
}

func TestRetryTransport_RespectsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	transport := &retryTransport{maxRetry: 5, wait: 10 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)

	done := make(chan struct{})
	var retErr error
	go func() {
		_, retErr = transport.RoundTrip(req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("RoundTrip did not return after context cancellation")
	}

	assert.ErrorIs(t, retErr, context.Canceled)
}
