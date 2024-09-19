GOOS=darwin GOARCH=amd64 go build -o client-mac main.go
chmod +x client-mac
xattr -d com.apple.quarantine ./client-mac

GOOS=linux GOARCH=amd64 go build -o client-linux main.go

GOOS=windows GOARCH=amd64 go build -o client-windows.exe main.go


scp ./client-* user@t8p:/home/user/


# 코드 사이닝

codesign --deep --force --verify --verbose --sign "Developer ID Application: Your Name" ./client-binary

xcrun altool --notarize-app --primary-bundle-id "com.yourcompany.client" --username "your-apple-id" --password "app-specific-password" --file ./client-binary
