# sesc-backend

## Development instructions
- To start the development database in docker: `make dev-db`
- To start the development API in docker: `make dev-backend`
- To run the API locally (recommended): `make dev-db && go run ./cmd/api/main.go | jq`

## !! READ THIS !! Implementing a new feature
1. Create the domain types in `sesc` package. If your types are purely application-level, like `Credentials`, place them in a separate package, like `iam`.
2. Create the necessary database schemas in `db/entdb/ent/schema`.
3. If your feature extends an existing service, write the related methods in a package of that service. If your feature requires a new service, write one in a new package, like `internal/service-name`.
4. When implementing methods, do not forget about observability. Use tools provided by `pkg/event` to make the current state of the request observable. Find examples of this in existing application logic, like in `internal/sescsvc`, `internal/iamsvc`, `internal/filesvc`, etc.
5. Write unit tests for your service.
6. Write API handlers and swagger docs in `api` package. Then, generate a new swagger doc with `go generate .` in the root of the project. Then, copy `api/docs/swagger.json` to `web/lib/`, in that directory run `npx swagger-typescript-api generate --axios --path ./swagger.json`. This will generate a client on the front end.
7. Write integration tests for your service (directory `tests`).
8. Run `make can-i-push` to verify that your code is compliant with the code style and does not break any tests.
9. Write and test the necessary frontend functionality.
10. Run `pnpm lint` and `pnpm build` to verify that there are no obvious issues in the frontend code.
11. Push to your feature branch, create a PR, set `kozlov-ma` as reviewer.

### EXAMPLE: adding file exchanging features to the application
This is an example of the development approach above, describing the implementation process for file-related application features.

1. Created types `File`, `FileCreationOptions`, `FileSearchOptions` in `sesc` package.
2. Created the `File` schema in `db/entdb/ent/schema/file.go`.
3. Created a file service in `internal/filesvc`. It depends on some s3 client. To test this package, we will need to mock this client, so `filesvc` now depends on `filesvc.ObjectStorage` interface. Then we implement this interface in the package `internal/s3svc`.
4. `filesvc` and `s3svc` were written with observability in mind.
5. Created unit tests in `internal/filesvc/service_test.go`. Also generated a mock `ObjectStorage` with `uber-go/mock`.
6. Wrote the necessary api handlers and middlewares, generated api docs and client.
7. Wrote the integration tests.
8. Using `make can-i-push` verified that the code is acceptable to push.
9. Wrote the frontend functionality.
10. Run `pnpm lint` and `pnpm build` to verify that there are no obvious issues in the frontend code.
11. Pushed the changes.


## Configuration

The application supports two configuration modes:

### Local Development
Uses `config.yml` file with sensible defaults. You can override any value with environment variables.

```bash
# Run with default config.yml
make dev-run

# Or override specific values
export SESC_DATABASE_ADDRESS="postgres://custom:pass@localhost:5432/db"
export SESC_JWT_SECRET="my-secret-key"
make dev-run
```

### Production / Kamal Deployment
Uses **only** environment variables (no `config.yml`). All configuration must be provided via `SESC_` prefixed environment variables.

**📖 See [ENV.md](./ENV.md) for complete environment variables reference**

Quick example:
```bash
SESC_DATABASE_ADDRESS=postgres://user:pass@host:5432/sesc
SESC_JWT_SECRET=your-secret-key
SESC_MINIO_ENDPOINT=minio:9000
SESC_MINIO_ACCESS_KEY=minioadmin
SESC_MINIO_SECRET_KEY=minioadmin
# ... etc
```

## Testing
To run tests:
```bash
make test
make integration-tests
```
