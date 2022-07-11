dep:
	go mod download

unit-test: dep
	go clean --testcache
	go test ./...

build-client: dep
	GOOS=linux go build -o build/client cmd/client/main.go

build-server: dep
	GOOS=linux go build -o build/server cmd/server/main.go

start: dep build-client build-server
	docker-compose up --force-recreate --build server client