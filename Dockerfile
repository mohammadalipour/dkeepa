# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/api

# Run Stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/bin/server .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]
