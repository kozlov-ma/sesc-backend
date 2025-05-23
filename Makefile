.PHONY: dev-db dev-backend down lint test integration-tests

# Generate Go code and frontend API client
generate:
	go generate ./... && cd web && pnpm openapi-ts

# Spin up the development database
dev-db:
	docker compose up -d postgres minio

# Spin up both database and backend
dev-backend: generate
	docker compose up -d --build postgres backend

# Stop and remove all containers
down:
	docker compose down

# Format the go files
fmt:
	go tool golangci-lint fmt ./...

# Stop and remove all containers and volumes
clean:
	docker compose down -v

# Run linting
lint:
	go tool golangci-lint run ./...

# Run tests
test:
	go test -v -coverprofile=c.out ./...

# Can you push and open a PR?
can-i-push:
	@make lint && make test && make integration-tests

# Can you push and open a PR for frontend changes?
can-i-push-web:
	cd web && pnpm lint && pnpm build

# Run integration tests in docker
integration-tests:
	docker compose up --build --abort-on-container-exit --exit-code-from integration-tests integration-tests
