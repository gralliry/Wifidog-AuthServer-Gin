env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auth-server
#env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o auth-server
#env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o auth-server.exe