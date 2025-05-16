# sesc-backend

## Development instructions
- To start the development database in docker: `make dev-db`
- To start the development API in docker: `make dev-backend`
- To run the API for development purposes: `make dev-db && go run cmd/api/main.go | jq`

The application can be configured using either a `config.yml` file or environment variables. A sample configuration file is provided in the repository root.

## Configuration
The application can be configured through:
1. A `config.yml` file in the root directory or in a `./config/` directory
2. Environment variables prefixed with `SESC_` (e.g., `SESC_POSTGRES_ADDRESS`)

Configuration options include:
- `postgres.address`: PostgreSQL connection string
- `http.server_address`: Address and port to bind the server to
- `http.read_header_timeout`, `http.read_timeout`, `http.write_timeout`: HTTP timeouts
- `jwt_secret`: Secret key for JWT token signing
- `admin_credentials`: Initial admin users with their credentials. To set it with env vars:
```bash
SESC_ADMIN_CREDENTIALS_0_ID="f1157f63-65dc-4c3d-bcb2-4d6d55d2e3fd"
SESC_ADMIN_CREDENTIALS_0_USERNAME="admin"
SESC_ADMIN_CREDENTIALS_0_PASSWORD="admin"
SESC_ADMIN_CREDENTIALS_1_ID="a33a8393-5e83-41cd-8532-1390952c00ee"
SESC_ADMIN_CREDENTIALS_1_USERNAME="another_admin"
SESC_ADMIN_CREDENTIALS_1_PASSWORD="secure_password"
```

## Project structure
### Packages and directories
- `cmd/api`: API entry point.
- `sesc`: Domain types and services that model the organization structure (users, roles, permissions).
- `api`: HTTP API adapter, a web entry point for the API. Also has the OpenAPI documentation and DTOs.
- `db`: Database interfaces and implementations.
- `iam`: Authentication and authorization service, providing JWT-based token management.
- `pkg/event`: Implementation of wide event tracking for monitoring and debugging.
- `internal/config`: Configuration management using Viper.
- `internal/slogsink`: Logging sink for structured logging with slog.

### Planned packages
- `achievement`: Domain types and services for managing achievements and achievement lists.

## Testing
To run tests:
```bash
make test
```
