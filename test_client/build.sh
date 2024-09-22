# cd ../test_client
GOOS=darwin GOARCH=amd64 go build -o client-mac .
GOOS=linux GOARCH=amd64 go build -o client-linux .
GOOS=windows GOARCH=amd64 go build -o client-windows .
# mkdir ../server/client
# mv client-* /home/user/client
# cd - 

# 배포 서버에서 실행
mv client-* /home/user/client