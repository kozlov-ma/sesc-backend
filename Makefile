.PHONY: dev-db dev-backend down lint test

# Spin up the development database
dev-db:
	docker-compose up -d postgres

# Spin up both database and backend
dev-backend:
	docker-compose up -d

# Stop and remove all containers
down:
	docker-compose down

# Stop and remove all containers and volumes
clean:
	docker compose down -v

# Run linting
lint:
	golangci-lint run ./...

# Run tests
test:
	go test -v ./...
