# build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o server .

# run
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 50000

CMD ["./server"]
