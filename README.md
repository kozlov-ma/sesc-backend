# sesc-backend

## Development instructions
- To start the development database and apply migrations to it: `docker compose up`
- To run the API for development purposes: `POSTGRES_ADDRESS="postgres://postgres:password@localhost:5432/postgres" go run cmd/api/main.go | jq -R`
