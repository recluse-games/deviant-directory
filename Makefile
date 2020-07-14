build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./bin/deviant-directory ./cmd/deviant-directory.go
	GOOS="windows" GOARCH="amd64" CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./bin/deviant-directory.exe ./cmd/deviant-directory.go
run:
	go run ./cmd/deviant-directory.go
test:
	go test ./... -cover
docker:
	sudo -E docker-compose up --build 