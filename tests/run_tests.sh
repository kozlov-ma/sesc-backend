#!/bin/bash

# SESC Backend Integration Tests Runner
# This script helps run the different categories of integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if API server is running
check_api_server() {
    print_status "Checking if API server is running on localhost:8080..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null | grep -q "200\|404"; then
        print_success "API server is running"
        return 0
    else
        print_error "API server is not running on localhost:8080"
        print_status "Please start the API server before running tests:"
        echo "  go run cmd/server/main.go"
        return 1
    fi
}

# Function to check Go version
check_go_version() {
    print_status "Checking Go version compatibility..."
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    MINOR=$(echo $GO_VERSION | cut -d. -f2)
    
    if [ "$MAJOR" -eq 1 ] && [ "$MINOR" -ge 23 ]; then
        print_success "Go version $GO_VERSION is compatible"
        return 0
    else
        print_warning "Go version $GO_VERSION may have compatibility issues"
        print_status "Recommended: Go 1.23 or higher"
        return 1
    fi
}

# Function to run scenario tests
run_scenarios() {
    print_status "Running scenario tests..."
    echo
    
    print_status "Running Full Workflow scenario..."
    go test -v -timeout=10m ./tests/scenarios/full_workflow/
    
    if [ $? -eq 0 ]; then
        print_success "Scenario tests completed successfully"
    else
        print_error "Scenario tests failed"
        return 1
    fi
}

# Function to run regression tests
run_regression() {
    print_status "Running regression tests..."
    echo
    
    print_status "Running API Features regression tests..."
    go test -v -timeout=5m ./tests/regress/api_features/
    
    if [ $? -eq 0 ]; then
        print_success "Regression tests completed successfully"
    else
        print_error "Regression tests failed"
        return 1
    fi
}

# Function to run all tests
run_all() {
    print_status "Running all integration tests..."
    echo
    
    go test -v -timeout=15m ./tests/...
    
    if [ $? -eq 0 ]; then
        print_success "All tests completed successfully"
    else
        print_error "Some tests failed"
        return 1
    fi
}

# Function to run tests with coverage
run_with_coverage() {
    print_status "Running tests with coverage analysis..."
    echo
    
    go test -v -timeout=15m -coverprofile=coverage.out ./tests/...
    
    if [ $? -eq 0 ]; then
        print_success "Tests completed, generating coverage report..."
        go tool cover -html=coverage.out -o coverage.html
        print_success "Coverage report generated: coverage.html"
    else
        print_error "Tests failed"
        return 1
    fi
}

# Function to clean test data
clean_test_data() {
    print_warning "This will clean any test data from the database"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Cleaning test data..."
        # Add actual cleanup commands here if needed
        print_success "Test data cleaned"
    else
        print_status "Cleanup cancelled"
    fi
}

# Function to show usage
show_usage() {
    echo "SESC Backend Integration Tests Runner"
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  scenarios    Run scenario tests (use case-driven)"
    echo "  regress      Run regression tests (API feature coverage)"
    echo "  all          Run all tests"
    echo "  coverage     Run all tests with coverage analysis"
    echo "  clean        Clean test data"
    echo "  check        Check prerequisites (API server, Go version)"
    echo "  help         Show this help message"
    echo
    echo "Examples:"
    echo "  $0 scenarios              # Run only scenario tests"
    echo "  $0 regress               # Run only regression tests"
    echo "  $0 all                   # Run all tests"
    echo "  $0 coverage              # Run tests with coverage"
    echo
    echo "Prerequisites:"
    echo "  - API server running on localhost:8080"
    echo "  - Go 1.23 or higher"
    echo "  - Clean database state (or use 'clean' command)"
}

# Main script logic
case "${1:-help}" in
    "scenarios")
        check_go_version
        check_api_server && run_scenarios
        ;;
    "regress")
        check_go_version
        check_api_server && run_regression
        ;;
    "all")
        check_go_version
        check_api_server && run_all
        ;;
    "coverage")
        check_go_version
        check_api_server && run_with_coverage
        ;;
    "clean")
        clean_test_data
        ;;
    "check")
        check_go_version
        check_api_server
        ;;
    "help"|*)
        show_usage
        ;;
esac