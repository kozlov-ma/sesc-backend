.PHONY: dev-db dev-backend down lint test integration-tests

# Generate Go code and frontend API client
generate:
	go generate ./... && cd web && rm -r lib/api && pnpm openapi-ts
	rm -r apiclient/*
	go tool swagger generate client -f api/docs/swagger.yaml -A apiclient --target apiclient/

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

# SSL Management Commands
ssl-init:
	@echo "🔒 Initializing SSL certificates..."
	./scripts/init-ssl.sh

ssl-renew:
	@echo "🔄 Renewing SSL certificates..."
	docker run --rm -v $(PWD)/certbot/conf:/etc/letsencrypt -v $(PWD)/certbot/www:/var/www/certbot certbot/certbot renew
	docker-compose -f docker-compose.simple.yml exec nginx nginx -s reload

ssl-status:
	@echo "📋 SSL Certificate Status:"
	docker run --rm -v $(PWD)/certbot/conf:/etc/letsencrypt certbot/certbot certificates

ssl-test:
	@echo "🧪 Testing SSL configuration..."
	@echo "Testing HTTPS redirect for sesc.online..."
	curl -I http://sesc.online || echo "❌ Failed to connect to sesc.online"
	@echo "Testing API endpoint..."
	curl -I https://api.sesc.online/health || echo "❌ Failed to connect to api.sesc.online"

# Production deployment
deploy-prod:
	@echo "🚀 Deploying to production..."
	docker-compose -f docker-compose.simple.yml pull
	docker-compose -f docker-compose.simple.yml up -d --remove-orphans
	docker image prune -f
	@echo "✅ Production deployment complete!"

# Production logs
logs-prod:
	docker-compose -f docker-compose.simple.yml logs -f

# Production status
status-prod:
	docker-compose -f docker-compose.simple.yml ps
