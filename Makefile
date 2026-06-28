.PHONY: up down restart migrate logs status test vet build-lambda

# Start all services (Database, Backend, Frontend)
up:
	docker compose up --build -d

# Stop all services
down:
	docker compose down

# Show running containers status
status:
	docker compose ps

# Tail logs of all containers
logs:
	docker compose logs -f

# Run database migrations inside backend container
migrate:
	docker compose exec backend go run cmd/migrate/main.go

# Restart Go backend to re-compile changed source files
restart:
	docker compose restart backend

# Run backend unit tests inside container
test:
	docker compose exec backend go test -v -race ./...

# Run Go vet linter inside container
vet:
	docker compose exec backend go vet ./...

# Build AWS Lambda deployment bundle locally (using Docker container compilation)
build-lambda:
	docker compose exec backend go build -o bootstrap ./cmd/lambda/main.go
