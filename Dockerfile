# Multi-stage build
FROM golang:alpine AS builder

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
ENTRYPOINT ["/app"]
