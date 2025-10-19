# Integration Testing System

This directory contains a comprehensive integration testing system for the SESC Backend API, organized into two main categories:

## Test Categories

### 1. Scenarios (`tests/scenarios/`)
Use case-driven tests that simulate real-world workflows and user interactions. These tests focus on end-to-end business processes.

#### Current Scenarios:
- **Full Workflow** (`tests/scenarios/full_workflow/`): Complete achievement management lifecycle covering 200 achievements across 20 users with multi-role approval processes.

### 2. Regression Tests (`tests/regress/`)
Feature coverage tests that verify all API endpoints work correctly and return appropriate errors when they should.

#### Current Regression Test Categories:
- **API Features** (`tests/regress/api_features/`): Comprehensive testing of all API endpoints, authentication, authorization, data validation, and error handling.

## Architecture

### Test Client Wrapper
Both test categories use a unified test client wrapper that encapsulates the generated `apiclient` package, providing:
- Simplified authentication management
- Helper methods for common operations
- Russian test data generation with transliteration
- Error handling standardization

### Generated API Client Integration
The tests utilize the Go-swagger generated API client (`apiclient` package) ensuring:
- Type safety with generated models
- Automatic request/response serialization
- OpenAPI specification compliance
- Consistent error handling

## Prerequisites

### Go Version Compatibility
**Important**: The project requires Go 1.24.2 as specified in `go.mod`, but this may cause compatibility issues with Go 1.23.x environments.

To resolve Go version compatibility:

1. **Check your Go version:**
   ```bash
   go version
   ```

2. **If using Go 1.23.x, temporarily modify go.mod:**
   ```bash
   # Backup current go.mod
   cp go.mod go.mod.backup
   
   # Edit go.mod to use Go 1.23
   sed -i 's/go 1.24.2/go 1.23/' go.mod
   ```

3. **After testing, restore original go.mod:**
   ```bash
   mv go.mod.backup go.mod
   ```

### API Server Setup
Ensure the SESC Backend API server is running:
```bash
# Start the server (adjust command as needed)
go run cmd/server/main.go

# Or if using a different startup method
make run
```

The tests assume the API server is running on `localhost:8080`.

## Running Tests

### Scenario Tests
Run the complete workflow scenario:
```bash
# Run the full workflow scenario
go test -v ./tests/scenarios/full_workflow/

# Run with detailed output
go test -v -count=1 ./tests/scenarios/full_workflow/
```

### Regression Tests
Run all API feature regression tests:
```bash
# Run all regression tests
go test -v ./tests/regress/api_features/

# Run specific test categories
go test -v ./tests/regress/api_features/ -run TestAuthenticationFeatures
go test -v ./tests/regress/api_features/ -run TestDepartmentFeatures
go test -v ./tests/regress/api_features/ -run TestUserFeatures
```

### Run All Tests
```bash
# Run all integration tests
go test -v ./tests/...

# Run with race detection
go test -v -race ./tests/...

# Generate test coverage
go test -v -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out -o coverage.html
```

## Test Data

### Russian Names and Transliteration
The testing system uses authentic Russian names for realistic test data:
- **First Names**: 38 common Russian first names (both male and female)
- **Last Names**: 76 common Russian surnames with proper gender forms
- **Middle Names**: 76 patronymic names with proper gender forms
- **Transliteration**: Automatic conversion to Latin characters for usernames

### Generated Test Data Examples:
- **User**: Александр Петров Михайлович → Username: APetrov123
- **User**: Екатерина Иванова Сергеевна → Username: EIvanova456
- **Department**: "Кафедра информатики" (Computer Science Department)
- **Achievement**: "Публикация статьи в журнале ВАК" (Publication in VAK journal)

## Test Structure

### Scenario Test Structure
```
tests/scenarios/full_workflow/
├── scenario.md           # Detailed scenario documentation
├── client.go            # Test client with helper methods
└── workflow_test.go     # Main test implementation
```

### Regression Test Structure
```
tests/regress/api_features/
├── client.go            # Test client wrapper
└── regress_test.go      # Comprehensive API feature tests
```

## Test Implementation Details

### Scenario Test: Full Workflow
**Purpose**: Test complete achievement management lifecycle
**Coverage**:
- 2 departments with 10 users each
- Multi-role user creation (teachers, department heads, deputies, academic director, economist)
- Credential setup and authentication
- File uploads (~10 files per user)
- Achievement creation (~10 achievements per user)
- Document attachments
- Multi-level approval workflow
- Report generation and accounting

**Test Steps**:
1. Admin authentication and setup
2. Department and user creation
3. Role assignments and credentials
4. User authentication and file uploads
5. Achievement creation and submission
6. Multi-level review process
7. Report generation and accounting
8. Final verification

### Regression Tests: API Features
**Purpose**: Verify all API endpoints work correctly and handle errors appropriately
**Coverage**:
- Authentication (admin/user login, token validation)
- Department management (CRUD operations, validation)
- User management (creation, updates, role assignments)
- Role and permission verification
- File management (upload, download, validation)
- Achievement lifecycle (creation, submission, review)
- Document attachments
- Achievement templates
- Reports and accounting
- Error handling and data validation
- Security (SQL injection, XSS prevention)

## Adding New Tests

### Adding a New Scenario
1. Create a new directory under `tests/scenarios/`
2. Create `scenario.md` documenting the test case
3. Create `client.go` with test-specific helper methods
4. Create `*_test.go` with the test implementation
5. Update this README with the new scenario

### Adding Regression Tests
1. Add new test methods to `regress_test.go`
2. Follow the naming convention: `test<Feature><Scenario>`
3. Use the existing test client methods or add new ones
4. Ensure both positive and negative test cases
5. Document any new test categories

## Test Configuration

### Environment Variables
```bash
# API server configuration
export API_HOST=localhost:8080
export API_SCHEME=http

# Test configuration
export TEST_TIMEOUT=300s
export TEST_PARALLEL=false
```

### Custom Test Configuration
Modify the test client initialization in each test package to customize:
- API server host/port
- Authentication credentials
- Timeout settings
- Retry policies

## Troubleshooting

### Common Issues

1. **Go Version Mismatch**
   ```
   Error: go.mod requires Go 1.24.2 but using Go 1.23.x
   ```
   Solution: Temporarily modify go.mod as described in Prerequisites

2. **API Server Not Running**
   ```
   Error: connection refused
   ```
   Solution: Ensure the API server is running on the expected port

3. **Authentication Failures**
   ```
   Error: admin login failed
   ```
   Solution: Verify admin credentials in the test configuration

4. **Database State Issues**
   ```
   Error: duplicate department name
   ```
   Solution: Tests assume a clean database state; restart the server or clear test data

### Debug Mode
Enable verbose logging in tests:
```bash
go test -v -args -debug=true ./tests/...
```

## Future Enhancements

### Planned Scenario Tests
- **Multi-Department Collaboration**: Cross-department achievement reviews
- **Bulk Operations**: Mass user creation and achievement processing
- **Data Migration**: Version upgrade and data consistency testing
- **Performance Testing**: Load testing with concurrent users

### Planned Regression Tests
- **API Versioning**: Backward compatibility testing
- **Rate Limiting**: API rate limit enforcement
- **Caching**: Response caching verification
- **Database Transactions**: Concurrency and isolation testing

## Contributing

When contributing new tests:
1. Follow the existing naming conventions
2. Include both positive and negative test cases
3. Document test scenarios in markdown files
4. Use the shared test client utilities
5. Ensure tests can run independently
6. Add appropriate assertions and error messages

## Dependencies

The testing system uses:
- **testify**: Assertion library and test suites
- **go-openapi**: Generated API client and models
- **Go standard library**: HTTP clients and testing framework

Install test dependencies:
```bash
go mod tidy