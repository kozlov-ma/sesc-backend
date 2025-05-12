# sesc-backend

## Development instructions
- To start the development database and apply migrations to it: `docker compose up`
- To run the API for development purposes: `POSTGRES_ADDRESS="postgres://postgres:password@localhost:5432/postgres?sslmode=disable" go run cmd/api/main.go | jq`

## Project structure
### Existing packages and directories
- `cmd/api`: API entry point.
- `cmd/pgsetup`: A tool to set up postgres for development, used in docker-compose to spin up a dev database.
- `sesc`: Domain types and services that model the organization structure (users, roles, permissions).
- `api`: HTTP API adapter, a web entry point for the API. Also has the openapi documentation and dtos.
- `db`: Package that defines database errors (probably should be refactored to use domain errors)
- `db/*` : Implementations of the DB interfaces that other services use.
- `migrations.old`: DB schema thoughts that were initially discussed, for reference only.
### Planned packages
- `achievement`: Domain types and services for managing achievements and achievement lists.
- `iam`: Auth process & access control to differentiate users and admins, make routes private, etc.

## Existing issues
- Poor testing. Need to write more comprehensive tests for all packages.
- For observability, currently only logs are available. We should consider adding metrics and tracing, also a profiler endpoint.
- No good way to configure the application.
- Code written in the `db/pgdb` package is mostly AI-generated, because of the lack of time. Its quality is rather poor, should be rewritten.
- Inconsistent log field naming.
