# Multi-stage build
FROM golang:alpine AS builder

LABEL org.opencontainers.image.source=https://github.com/cyb3rko/matrix-guardian
LABEL org.opencontainers.image.licenses=MPL-2.0
LABEL org.opencontainers.image.title="Matrix Guardian"
LABEL org.opencontainers.image.description="A friendly Matrix bot for protecting the people"

# Move to working directory /build
WORKDIR /build

# Copy the go.mod and go.sum files to the /build directory
COPY go.mod go.sum ./
# Install dependencies
RUN go mod download
COPY . .

# Install gcc components (required for sqlite)
RUN apk add gcc musl-dev
# Build the application (additional flags for external libraries on scratch; required for sqlite)
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o app

FROM scratch
# Copy TLS certificates to allow TLS traffic
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Copy binary
COPY --from=builder /build/app /app
# Copy empty database directory
COPY --from=builder /build/data /data
ENTRYPOINT ["/app"]
