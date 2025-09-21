package workflow

import (
	"testing"
)

// TestAccountingPeriodFileDeletionIntegration tests file deletion when accounting periods finish
func TestAccountingPeriodFileDeletionIntegration(t *testing.T) {
	t.Log("🧪 Testing Accounting Period File Deletion")

	// Setup
	client := setupTestClient(t)
	client.cleanupExistingPeriods(t)

	// Create test periods with files
	periods := createTestPeriodsWithFiles(t, client)

	// Verify file deletion behavior
	verifyFileDeletion(t, client, periods)

	t.Log("🎉 Accounting Period File Deletion test completed successfully!")
}
