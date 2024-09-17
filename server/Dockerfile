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

# Build the Go server
RUN go build -o server main.go

# Final image
FROM alpine:latest

WORKDIR /root/

# Copy the compiled binary from the builder
COPY --from=builder /app/server .

# Copy static files (React build) to serve with the backend
COPY static/ ./static/

# Execute the binary
CMD ["./server"]
