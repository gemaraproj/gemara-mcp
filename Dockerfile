FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder

# Install build dependencies
RUN apk add --no-cache git make

ARG VERSION="dev"
ARG BUILD="dev"

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build VERSION=${VERSION} BUILD=${BUILD}

FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

COPY --from=builder /build/bin/gemara-mcp /bin/gemara-mcp
WORKDIR /workspace

ENTRYPOINT ["/bin/gemara-mcp"]
