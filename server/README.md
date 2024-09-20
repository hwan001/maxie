```bash
GOOS=linux GOARCH=amd64 go build -o server .

scp server user@t8p:/home/user/server

go run .

curl -i -X POST http://localhost:8080/api/auth/google \
-H "Content-Type: application/x-www-form-urlencoded" \
-d "code=코드" -> 이건 리액트 로그로 확인
```