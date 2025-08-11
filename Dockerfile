# Build stage
FROM golang:1.24-alpine AS builder

# Install git and other dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kwiki ./cmd/kwiki

# Final stage
FROM alpine:latest

# Install ca-certificates and git (needed for cloning repositories)
RUN apk --no-cache add ca-certificates git

# Create app directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/kwiki .

# Copy configuration and web assets
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/web ./web

# Create necessary directories
RUN mkdir -p repos output

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release

# Run the application
CMD ["./kwiki"]
