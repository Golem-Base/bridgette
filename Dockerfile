# syntax=docker/dockerfile:1
FROM golang:1.23.8 AS builder

WORKDIR /build
ADD . /build/

# Create output directory
RUN mkdir /out

# Build all binaries with CGO enabled and static linking
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod/ \
    CGO_ENABLED=1 go build -ldflags="-w -s -linkmode external -extldflags '-static'" -o /out/service .
# Create final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl

WORKDIR /app

# Copy all binaries
COPY --from=builder /out/service /app

# Set default entrypoint to server
ENTRYPOINT ["/app/service"] 