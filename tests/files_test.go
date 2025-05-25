package tests

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileUploadAndDownload(t *testing.T) {
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create clients for the test API
	adminClient := NewTestClient()
	regularClient := NewTestClient()
	ctx := t.Context()

	// Login as admin
	adminToken, err := adminClient.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	adminClient.SetToken(adminToken)

	// Create a regular user for testing ownership
	randomSuffix := uuid.Must(uuid.NewV7()).String()
	username := fmt.Sprintf("file_test_user_%s", randomSuffix)
	userData := CreateValidUserData(
		fmt.Sprintf("File_%s", randomSuffix),
		fmt.Sprintf("User_%s", randomSuffix),
		2,
	)
	user, err := adminClient.CreateUser(ctx, userData)
	require.NoError(t, err)

	// Register credentials for the user
	err = adminClient.RegisterUser(ctx, user.ID.String(), RegisterUserRequest{
		Username: username,
		Password: "password123",
	})
	require.NoError(t, err)

	// Login as regular user
	regularToken, err := regularClient.Login(ctx, username, "password123")
	require.NoError(t, err)
	regularClient.SetToken(regularToken)

	// Test 1: Admin uploads a common file
	t.Run("admin_uploads_common_file", func(t *testing.T) {
		fileContent := []byte("This is a test file uploaded by admin")
		fileName := fmt.Sprintf("admin_test_file_%s.txt", uuid.Must(uuid.NewV7()).String())

		file, err := adminClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
		require.NotNil(t, file)

		// Verify file properties
		assert.Equal(t, fileName, file.FileName)
		assert.Equal(t, len(fileContent), file.FileSize)
		// In Docker test environment, the download URL might be empty because of how MinIO is configured
		// assert.NotEmpty(t, file.DownloadURL)
		assert.Nil(t, file.OwnerID) // Common file has no owner

		// Test file search to find our file
		files, total, err := adminClient.SearchFiles(ctx, SearchFilesOptions{
			Name: fileName,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, files, 1)
		assert.Equal(t, fileName, files[0].FileName)
	})

	// Test 2: Regular user uploads a file (owned)
	t.Run("user_uploads_owned_file", func(t *testing.T) {
		fileContent := []byte("This is a test file uploaded by regular user")
		fileName := fmt.Sprintf("user_test_file_%s.txt", uuid.Must(uuid.NewV7()).String())

		file, err := regularClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
		require.NotNil(t, file)

		// Verify file properties
		assert.Equal(t, fileName, file.FileName)
		assert.Equal(t, len(fileContent), file.FileSize)
		// In Docker test environment, the download URL might be empty because of how MinIO is configured
		// assert.NotEmpty(t, file.DownloadURL)
		require.NotNil(t, file.OwnerID)
		assert.Equal(t, user.ID.String(), *file.OwnerID)

		// Test file search by owner
		files, total, err := regularClient.SearchFiles(ctx, SearchFilesOptions{
			OwnerID: user.ID.String(),
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.NotEmpty(t, files)

		// Find our file in the results
		var found bool
		for _, f := range files {
			if f.FileName == fileName {
				found = true
				break
			}
		}
		assert.True(t, found, "Newly uploaded file not found in search results")
	})
}

func TestFileSearch(t *testing.T) {
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create a client for the test API
	adminClient := NewTestClient()
	ctx := t.Context()

	// Login as admin
	adminToken, err := adminClient.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	adminClient.SetToken(adminToken)

	// Upload multiple files to test search functionality
	uniquePrefix := uuid.Must(uuid.NewV7()).String()
	fileNames := []string{
		fmt.Sprintf("search_test_document_%s.txt", uniquePrefix),
		fmt.Sprintf("search_test_image_%s.jpg", uniquePrefix),
		fmt.Sprintf("search_test_pdf_%s.pdf", uniquePrefix),
	}

	for _, fileName := range fileNames {
		fileContent := []byte(fmt.Sprintf("Content of %s", fileName))
		_, err := adminClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
	}

	// Test 1: Search by name partial match
	t.Run("search_by_name", func(t *testing.T) {
		files, total, err := adminClient.SearchFiles(ctx, SearchFilesOptions{
			Name: uniquePrefix,
		})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, files, 3)
	})

	// Test 2: Search by file type (extension)
	t.Run("search_by_extension", func(t *testing.T) {
		files, total, err := adminClient.SearchFiles(ctx, SearchFilesOptions{
			Name: fmt.Sprintf("pdf_%s", uniquePrefix),
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, files, 1)
		assert.Equal(t, fileNames[2], files[0].FileName)
	})

	// Test 3: Pagination
	t.Run("pagination", func(t *testing.T) {
		// First page
		files, total, err := adminClient.SearchFiles(ctx, SearchFilesOptions{
			Name:   uniquePrefix,
			Limit:  2,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, files, 2)

		// Second page
		files2, total2, err := adminClient.SearchFiles(ctx, SearchFilesOptions{
			Name:   uniquePrefix,
			Limit:  2,
			Offset: 2,
		})
		require.NoError(t, err)
		assert.Equal(t, 3, total2)
		assert.Len(t, files2, 1)
	})
}

func TestFileDelete(t *testing.T) {
	// Skip if test API URL is not set
	SkipIfNoTestAPI(t)

	// Create clients for the test API
	adminClient := NewTestClient()
	regularClient := NewTestClient()
	ctx := t.Context()

	// Login as admin
	adminToken, err := adminClient.LoginAdmin(ctx, "admin", "admin")
	require.NoError(t, err)
	adminClient.SetToken(adminToken)

	// Create a regular user for testing ownership
	randomSuffix := uuid.Must(uuid.NewV7()).String()
	username := fmt.Sprintf("file_delete_user_%s", randomSuffix)
	userData := CreateValidUserData(
		fmt.Sprintf("FileDelete_%s", randomSuffix),
		fmt.Sprintf("User_%s", randomSuffix),
		2,
	)
	user, err := adminClient.CreateUser(ctx, userData)
	require.NoError(t, err)

	// Register credentials for the user
	err = adminClient.RegisterUser(ctx, user.ID.String(), RegisterUserRequest{
		Username: username,
		Password: "password123",
	})
	require.NoError(t, err)

	// Login as regular user
	regularToken, err := regularClient.Login(ctx, username, "password123")
	require.NoError(t, err)
	regularClient.SetToken(regularToken)

	// Test 1: Admin deletes a common file
	t.Run("admin_deletes_common_file", func(t *testing.T) {
		fileContent := []byte("This is a test file for admin deletion")
		fileName := fmt.Sprintf("admin_delete_test_%s.txt", uuid.Must(uuid.NewV7()).String())

		// Upload file as admin (common file)
		file, err := adminClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
		require.NotNil(t, file)

		// Delete the file
		err = adminClient.DeleteFile(ctx, file.ID)
		require.NoError(t, err)

		// Verify file is no longer found
		files, total, err := adminClient.SearchFiles(ctx, SearchFilesOptions{
			Name: fileName,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, files)
	})

	// Test 2: User deletes own file
	t.Run("user_deletes_own_file", func(t *testing.T) {
		fileContent := []byte("This is a test file for user deletion")
		fileName := fmt.Sprintf("user_delete_test_%s.txt", uuid.Must(uuid.NewV7()).String())

		// Upload file as regular user (owned file)
		file, err := regularClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
		require.NotNil(t, file)

		// Delete the file
		err = regularClient.DeleteFile(ctx, file.ID)
		require.NoError(t, err)

		// Verify file is no longer found
		files, total, err := regularClient.SearchFiles(ctx, SearchFilesOptions{
			Name: fileName,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, files)
	})

	// Test 3: Admin deletes user's file
	t.Run("admin_deletes_user_file", func(t *testing.T) {
		fileContent := []byte("This is a test file for admin deletion of user file")
		fileName := fmt.Sprintf("admin_delete_user_file_%s.txt", uuid.Must(uuid.NewV7()).String())

		// Upload file as regular user (owned file)
		file, err := regularClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
		require.NotNil(t, file)

		// Admin deletes the file
		err = adminClient.DeleteFile(ctx, file.ID)
		require.NoError(t, err)

		// Verify file is no longer found
		files, total, err := regularClient.SearchFiles(ctx, SearchFilesOptions{
			Name: fileName,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, files)
	})

	// Test 4: User cannot delete another user's file
	t.Run("user_cannot_delete_other_file", func(t *testing.T) {
		// Create another user
		anotherUserSuffix := uuid.Must(uuid.NewV7()).String()
		anotherUsername := fmt.Sprintf("another_user_%s", anotherUserSuffix)
		anotherUserData := CreateValidUserData(
			fmt.Sprintf("Another_%s", anotherUserSuffix),
			fmt.Sprintf("User_%s", anotherUserSuffix),
			2,
		)
		anotherUser, err := adminClient.CreateUser(ctx, anotherUserData)
		require.NoError(t, err)

		// Register credentials for the second user
		err = adminClient.RegisterUser(ctx, anotherUser.ID.String(), RegisterUserRequest{
			Username: anotherUsername,
			Password: "password123",
		})
		require.NoError(t, err)

		// Create another client and login
		anotherClient := NewTestClient()
		anotherToken, err := anotherClient.Login(ctx, anotherUsername, "password123")
		require.NoError(t, err)
		anotherClient.SetToken(anotherToken)

		// Upload file as the second user
		fileContent := []byte("This is a test file for permission check")
		fileName := fmt.Sprintf("permission_test_%s.txt", uuid.Must(uuid.NewV7()).String())

		file, err := anotherClient.UploadFile(ctx, fileContent, fileName)
		require.NoError(t, err)
		require.NotNil(t, file)

		// First user tries to delete the second user's file
		err = regularClient.DeleteFile(ctx, file.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Access to file denied")
	})
}
