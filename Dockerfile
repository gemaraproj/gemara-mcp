FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

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

FROM gcr.io/distroless/static-debian12:nonroot@sha256:aef9602f8710ec12bde19d593fed1f76c708531bb7aba205110f1029786ead7b

COPY --from=builder /build/bin/gemara-mcp /bin/gemara-mcp
WORKDIR /workspace

ENTRYPOINT ["/bin/gemara-mcp"]
