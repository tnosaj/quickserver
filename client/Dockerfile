# Start from the Go base image
FROM golang:1.22 AS builder

WORKDIR /app

# Copy go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

# Final lightweight image
FROM scratch

WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/app .

# App runs on port 8080
EXPOSE 8080

ENTRYPOINT ["/app/app"]
