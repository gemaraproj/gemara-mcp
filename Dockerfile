FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder

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

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

COPY --from=builder /build/bin/gemara-mcp /bin/gemara-mcp
WORKDIR /workspace

ENTRYPOINT ["/bin/gemara-mcp"]
