# cd ../test_client
GOOS=darwin GOARCH=amd64 go build -o client-mac .
GOOS=linux GOARCH=amd64 go build -o client-linux .
GOOS=windows GOARCH=amd64 go build -o client-windows .
# mkdir ../server/client
# mv client-* ../server/client
# cd - 
