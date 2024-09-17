# Base image
FROM golang:1.18-alpine AS builder

# Set work directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all files to the container
COPY . .

# Build the Go client
RUN go build -o client main.go

# Final image
FROM alpine:latest

WORKDIR /root/

# Copy the compiled binary from the builder
COPY --from=builder /app/client .

# Execute the binary
CMD ["./client"]
