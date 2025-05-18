package tests

import (
	"os"
	"testing"
)

// SkipIfNoTestAPI skips the integration test if TEST_API_URL is not set
func SkipIfNoTestAPI(t *testing.T) {
	t.Helper()

	if os.Getenv("TEST_API_URL") == "" {
		t.Skip("Skipping integration test, set TEST_API_URL to the test API server address")
	}
}
