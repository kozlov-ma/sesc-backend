# Integration Tests

This directory contains integration tests for the SESC Management API. These tests start a real application instance with a test database and make HTTP requests to it to verify the functionality.

## Structure

- `client.go` - Contains the API client used for testing
- `models.go` - Contains the data models used by the client
- `*_test.go` - Test files for different parts of the API

## Running the Tests

To run the integration tests, you need a PostgreSQL database. By default, the tests will use a local database with the credentials:

```
postgres://postgres:postgres@localhost:5432/test_*
```

Where `test_*` is a randomly generated database name.

You can specify a different database connection string using the `TEST_POSTGRES_ADDR` environment variable:

```bash
export TEST_POSTGRES_ADDR="postgres://user:password@host:port/db?sslmode=disable"
go test -v ./tests/...
```

## What is Tested

1. **Authentication**
   - Admin Login
   - User Login Flow (create user, register credentials, login)
   - Token Validation

2. **Departments**
   - CRUD operations (Create, Read, Update, Delete)
   - Listing departments

3. **Users**
   - CRUD operations
   - User patching
   - Credential management

## Adding More Tests

To add more tests:

1. Create a new `_test.go` file or add tests to an existing one
2. Use the `Client` from `client.go` to interact with the API
3. Use the test utilities from `internal/testutil` to create and start the application 