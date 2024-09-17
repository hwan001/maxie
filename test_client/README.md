GOOS=darwin GOARCH=amd64 go build -o client-mac main.go
chmod +x client-mac
xattr -d com.apple.quarantine ./client-mac

GOOS=linux GOARCH=amd64 go build -o client-linux main.go

GOOS=windows GOARCH=amd64 go build -o client-windows.exe main.go


scp ./client-* user@t8p:/home/user/