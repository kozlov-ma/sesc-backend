.PHONY: dev-db dev-backend down lint test integration-tests

# Generate Go code and frontend API client
generate:
	go generate ./... && cd web && rm -rf lib/api && pnpm openapi-ts
	rm -rf apiclient/*
	go tool swagger generate client -f api/docs/swagger.yaml -A apiclient --target apiclient/

# Spin up the development database
dev-db:
	docker compose up -d postgres minio phpldapadmin ldap

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
	go test -coverprofile=c.out ./...

# Can you push and open a PR?
can-i-push:
	@make fmt && make lint && make test && make integration-tests

# Can you push and open a PR for frontend changes?
can-i-push-web:
	cd web && pnpm lint && pnpm build

# Run integration tests in docker
integration-tests:
	docker compose -f docker-compose-test.yml down -v
	docker compose -f docker-compose-test.yml up --build --abort-on-container-exit --exit-code-from integration-tests integration-tests
	docker compose -f docker-compose-test.yml down -v

# Deploy application with Kamal
deploy:
	kamal deploy

# Show deployment status
status:
	kamal app details
	kamal accessory details all
