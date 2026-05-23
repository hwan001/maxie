# build
FROM golang:alpine AS builder

WORKDIR /app
COPY . .

# secret.go가 없을 때(gitignore 환경) build arg로 생성
ARG GOOGLE_CLIENT_ID=""
ARG GOOGLE_CLIENT_SECRET=""
ARG REDIRECT_URI="http://localhost:8080"
ARG JWT_SECRET="dev-jwt-secret-change-me"

RUN if [ ! -f secret.go ]; then \
    printf 'package main\n\nvar (\n\tclientID     = "%s"\n\tclientSecret = "%s"\n\tredirectURI  = "%s"\n\tjwtSecret    = []byte("%s")\n)\n' \
    "$GOOGLE_CLIENT_ID" "$GOOGLE_CLIENT_SECRET" "$REDIRECT_URI" "$JWT_SECRET" > secret.go; \
    fi

RUN go mod tidy
RUN go build -o server .

# run
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 50000

CMD ["./server"]
