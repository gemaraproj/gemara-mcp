// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	headerRetryAfter = "Retry-After"
)

// defaultClient is the shared HTTP client configured with retry and timeout.
var defaultClient = &http.Client{
	Timeout:   defaultTimeout,
	Transport: newRetryTransport(),
}

// isRetryable returns whether the request should be retried based on the
// response status code or error type.
func isRetryable(resp *http.Response, err error) (bool, error) {
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true, nil
		}
		return false, err
	}

	if resp.StatusCode == http.StatusRequestTimeout ||
		resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	if resp.StatusCode == 0 || resp.StatusCode >= 500 {
		return true, nil
	}

	return false, nil
}

// retryAfter returns the server-requested wait from a Retry-After header,
// or zero if not present or unparseable.
func retryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	if v := resp.Header.Get(headerRetryAfter); v != "" {
		if seconds, _ := strconv.ParseInt(v, 10, 64); seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 0
}

// retryTransport is an http.RoundTripper that retries transient failures
// with a fixed delay, respecting Retry-After headers on 429 responses.
type retryTransport struct {
	maxRetry int
	wait     time.Duration
}

func newRetryTransport() *retryTransport {
	return &retryTransport{
		maxRetry: 3,
		wait:     1 * time.Second,
	}
}

// RoundTrip executes an HTTP request with retry logic.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	for attempt := 0; ; attempt++ {
		resp, respErr := http.DefaultTransport.RoundTrip(req)

		if attempt >= t.maxRetry {
			return resp, respErr
		}

		shouldRetry, retryErr := isRetryable(resp, respErr)
		if retryErr != nil {
			if respErr == nil {
				_ = resp.Body.Close()
			}
			return nil, retryErr
		}
		if !shouldRetry {
			return resp, respErr
		}

		if req.Body != nil {
			if req.GetBody == nil {
				return resp, respErr
			}
			body, err := req.GetBody()
			if err != nil {
				return resp, respErr
			}
			req.Body = body
		}

		if respErr == nil {
			_ = resp.Body.Close()
		}

		wait := t.wait
		if ra := retryAfter(resp); ra > 0 {
			wait = ra
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
