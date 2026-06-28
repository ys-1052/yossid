.PHONY: fmt vet test build run-local db-up db-down migrate

# Default target
all: fmt vet test build

# Format Go code
fmt:
	go fmt ./...

# Run go vet (standard linter)
vet:
	go vet ./...

# Run tests
test:
	go test -v -race ./...

# Build lambda and server binaries
build:
	GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/lambda/main.go
	go build -o server ./cmd/server/main.go

# Run server locally
run-local:
	go run cmd/server/main.go

# Docker Compose DB commands
db-up:
	docker-compose up -d

db-down:
	docker-compose down

# Run DB Migrations
migrate:
	go run cmd/migrate/main.go
