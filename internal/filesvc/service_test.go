//go:generate go tool mockgen -destination=./mocks/mock_storage.go -package=mocks . ObjectStorage
package filesvc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/internal/filesvc/mocks"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// requireFileMatches is a helper to verify that a file matches expected values
func requireFileMatches(t *testing.T, expected, actual *ent.File) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID, "File ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "File name mismatch")
	require.Equal(t, expected.Size, actual.Size, "File size mismatch")

	// Only check ownerID if expected has one
	if expected.OwnerID != nil {
		require.NotNil(t, actual.OwnerID, "Expected file to have an owner")
		require.Equal(t, *expected.OwnerID, *actual.OwnerID, "File owner ID mismatch")
	} else {
		require.Nil(t, actual.OwnerID, "Expected file to have no owner")
	}

	require.Equal(t, expected.S3ObjectKey, actual.S3ObjectKey, "S3 object key mismatch")
}

// setupFileService creates a new FileService for testing
func setupFileService(t *testing.T) (*FileService, *mocks.MockObjectStorage, *ent.Client) {
	t.Helper()

	connStr := fmt.Sprintf("file:ent%d?mode=memory&cache=shared&_fk=1", rand.Int64())
	client := enttest.Open(t, "sqlite3", connStr)
	t.Cleanup(func() {
		_ = client.Close()
	})

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	mockStorage := mocks.NewMockObjectStorage(ctrl)

	return New(client, mockStorage, "test-bucket"), mockStorage, client
}

// createTestUser creates a test user in the database
func createTestUser(ctx context.Context, t *testing.T, client *ent.Client) uuid.UUID {
	t.Helper()
	userID := uuid.Must(uuid.NewV7())

	// Create a department with a unique name
	deptID := uuid.Must(uuid.NewV7())
	deptName := "Test Department " + deptID.String()
	_, err := client.Department.Create().
		SetID(deptID).
		SetName(deptName).
		SetDescription("For Testing").
		Save(ctx)
	require.NoError(t, err)

	// Then create the user with that department
	user, err := client.User.Create().
		SetID(userID).
		SetFirstName("Test").
		SetLastName("User").
		SetRole(1).
		SetDepartmentID(deptID).
		SetSubdivision("Test Subdivision").
		SetJobTitle("Test Position").
		SetEmploymentRate(1.0).
		SetPersonnelCategory(1).
		SetEmploymentType(1).
		SetDateOfEmployment(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	return user.ID
}

func TestCreate(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *FileService, storage *mocks.MockObjectStorage, client *ent.Client) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, client = setupFileService(t)
		return ctx, svc, storage, client
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, storage, _ := setup(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "test.txt",
			FileSize: len(content),
		}

		// Setup expectations
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := svc.Create(ctx, reader, opts)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, file.ID)
		require.Equal(t, opts.FileName, file.Name)
		require.Equal(t, opts.FileSize, file.Size)
		require.Nil(t, file.OwnerID)
		require.NotEmpty(t, file.S3ObjectKey)

		// Verify the file exists in the database
		savedFile, err := svc.ByID(ctx, file.ID)
		require.NoError(t, err)
		requireFileMatches(t, file, savedFile)
	})

	t.Run("with_owner", func(t *testing.T) {
		ctx, svc, storage, client := setup(t)

		// Create a test user first
		ownerID := createTestUser(ctx, t, client)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "test.txt",
			FileSize: len(content),
			OwnerID:  &ownerID,
		}

		// Setup expectations
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := svc.Create(ctx, reader, opts)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, file.ID)
		require.Equal(t, opts.FileName, file.Name)
		require.Equal(t, opts.FileSize, file.Size)
		require.NotNil(t, file.OwnerID)
		require.Equal(t, ownerID, *file.OwnerID)
		require.NotEmpty(t, file.S3ObjectKey)

		// Verify the file exists in the database
		savedFile, err := svc.ByID(ctx, file.ID)
		require.NoError(t, err)
		requireFileMatches(t, file, savedFile)
	})

	t.Run("invalid_file_name", func(t *testing.T) {
		ctx, svc, _, _ := setup(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "", // Empty file name
			FileSize: len(content),
		}

		_, err := svc.Create(ctx, reader, opts)
		require.ErrorIs(t, err, sesc.ErrInvalidFileName)
	})

	t.Run("invalid_file_size", func(t *testing.T) {
		ctx, svc, _, _ := setup(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "test.txt",
			FileSize: 0, // Zero file size
		}

		_, err := svc.Create(ctx, reader, opts)
		require.ErrorIs(t, err, sesc.ErrInvalidFileSize)
	})

	t.Run("storage_error", func(t *testing.T) {
		ctx, svc, storage, _ := setup(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "test.txt",
			FileSize: len(content),
		}

		storageErr := errors.New("storage error")
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(storageErr)

		// Add mock for RemoveObject since it's called during error cleanup
		storage.EXPECT().RemoveObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := svc.Create(ctx, reader, opts)

		require.Error(t, err)
		require.Equal(t, storageErr, err)

		// Verify no file was created in the database
		files, _, err := svc.Search(ctx, sesc.FileSearchOptions{})
		require.NoError(t, err)
		require.Empty(t, files)
	})
}

func TestDelete(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *FileService, fileID uuid.UUID, storage *mocks.MockObjectStorage) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, _ = setupFileService(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "test.txt",
			FileSize: len(content),
		}

		// Setup expectations for file creation
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := svc.Create(ctx, reader, opts)
		require.NoError(t, err)

		return ctx, svc, file.ID, storage
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, fileID, storage := setup(t)

		// Setup expectation for deletion
		storage.EXPECT().RemoveObject(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.Delete(ctx, fileID)
		require.NoError(t, err)

		// Verify the file no longer exists
		_, err = svc.ByID(ctx, fileID)
		require.ErrorIs(t, err, sesc.ErrFileNotFound)
	})

	t.Run("non_existent_file", func(t *testing.T) {
		ctx, svc, _, _ := setup(t)

		nonExistentID := uuid.Must(uuid.NewV7())
		err := svc.Delete(ctx, nonExistentID)
		require.ErrorIs(t, err, sesc.ErrFileNotFound)
	})

	t.Run("storage_error", func(t *testing.T) {
		ctx, svc, fileID, storage := setup(t)

		storageErr := errors.New("storage error")
		storage.EXPECT().RemoveObject(gomock.Any(), gomock.Any()).
			Return(storageErr)

		err := svc.Delete(ctx, fileID)
		require.Error(t, err)
		require.Equal(t, storageErr, err)

		// Verify the file still exists in the database
		_, err = svc.ByID(ctx, fileID)
		require.NoError(t, err)
	})
}

func TestSearch(t *testing.T) {
	setupBase := func(t *testing.T) (ctx context.Context, svc *FileService, client *ent.Client) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, _, client = setupFileService(t)
		return ctx, svc, client
	}

	t.Run("by_owner", func(t *testing.T) {
		ctx, svc, client := setupBase(t)

		// Create a test user
		ownerID := createTestUser(ctx, t, client)

		// Get mock storage to set expectations
		mockStorage := svc.storage.(*mocks.MockObjectStorage)
		mockStorage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(4))).Return(nil)

		_, err := svc.Create(ctx, bytes.NewReader([]byte("test")), FileOpts{
			FileName: "new-owner-file.txt",
			FileSize: 4,
			OwnerID:  &ownerID,
		})
		require.NoError(t, err)

		files, total, err := svc.Search(ctx, sesc.FileSearchOptions{
			OwnerID: &ownerID,
		})
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, files, 1)
		require.Equal(t, "new-owner-file.txt", files[0].Name)
	})

	t.Run("common_files", func(t *testing.T) {
		ctx, svc, client := setupBase(t)

		// Create some files
		fileContents := []byte("test content")

		// Get mock storage to set expectations
		mockStorage := svc.storage.(*mocks.MockObjectStorage)
		mockStorage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(fileContents)))).
			Return(nil).
			Times(2)

		// Create a common file
		_, err := svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "common-file.txt",
			FileSize: len(fileContents),
		})
		require.NoError(t, err)

		// Create one with an owner
		ownerID := createTestUser(ctx, t, client)
		_, err = svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "owned-file.txt",
			FileSize: len(fileContents),
			OwnerID:  &ownerID,
		})
		require.NoError(t, err)

		files, total, err := svc.Search(ctx, sesc.FileSearchOptions{
			Common: true,
		})
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, files, 1)
		require.Equal(t, "common-file.txt", files[0].Name)
	})

	t.Run("by_name", func(t *testing.T) {
		ctx, svc, client := setupBase(t)

		fileContents := []byte("test content")

		// Get mock storage to set expectations
		mockStorage := svc.storage.(*mocks.MockObjectStorage)
		mockStorage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(fileContents)))).
			Return(nil).
			Times(2)

		// Create files with specific names
		_, err := svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "document.txt",
			FileSize: len(fileContents),
		})
		require.NoError(t, err)

		// Create one with "image" in the name
		ownerID := createTestUser(ctx, t, client)
		_, err = svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "test-image.jpg",
			FileSize: len(fileContents),
			OwnerID:  &ownerID,
		})
		require.NoError(t, err)

		files, total, err := svc.Search(ctx, sesc.FileSearchOptions{
			Name: "image",
		})
		require.NoError(t, err)
		require.Equal(t, 1, total)
		require.Len(t, files, 1)
		require.Equal(t, "test-image.jpg", files[0].Name)
	})

	t.Run("pagination", func(t *testing.T) {
		ctx, svc, _ := setupBase(t)

		fileContents := []byte("test content")

		// Get mock storage to set expectations
		mockStorage := svc.storage.(*mocks.MockObjectStorage)
		mockStorage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(fileContents)))).
			Return(nil).
			Times(4)

		// Create 4 files to test pagination
		for i := range 4 {
			fileName := fmt.Sprintf("file-%d.txt", i+1)
			_, err := svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
				FileName: fileName,
				FileSize: len(fileContents),
			})
			require.NoError(t, err)
		}

		// First page
		files, total, err := svc.Search(ctx, sesc.FileSearchOptions{
			Limit:  2,
			Offset: 0,
			Common: true,
		})
		require.NoError(t, err)
		require.Equal(t, 4, total)
		require.Len(t, files, 2)

		// Second page
		files2, total2, err := svc.Search(ctx, sesc.FileSearchOptions{
			Limit:  2,
			Offset: 2,
			Common: true,
		})
		require.NoError(t, err)
		require.Equal(t, 4, total2)
		require.Len(t, files2, 2)

		// Ensure no overlap between pages
		for _, f1 := range files {
			for _, f2 := range files2 {
				require.NotEqual(t, f1.ID, f2.ID)
			}
		}
	})

	t.Run("combined_filters", func(t *testing.T) {
		ctx, svc, client := setupBase(t)

		// Create a test user
		ownerID := createTestUser(ctx, t, client)

		fileContents := []byte("test")

		// Get mock storage to set expectations
		mockStorage := svc.storage.(*mocks.MockObjectStorage)
		mockStorage.EXPECT().
			PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(fileContents)))).
			Return(nil).
			Times(3)

		// Create files with PDF in the name
		_, err := svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "test-pdf-doc.pdf",
			FileSize: len(fileContents),
			OwnerID:  &ownerID,
		})
		require.NoError(t, err)

		_, err = svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "another-pdf-file.pdf",
			FileSize: len(fileContents),
			OwnerID:  &ownerID,
		})
		require.NoError(t, err)

		// Create one without PDF in the name
		_, err = svc.Create(ctx, bytes.NewReader(fileContents), FileOpts{
			FileName: "document.txt",
			FileSize: len(fileContents),
			OwnerID:  &ownerID,
		})
		require.NoError(t, err)

		// Test combined filtering
		files, total, err := svc.Search(ctx, sesc.FileSearchOptions{
			Name:    "pdf",
			OwnerID: &ownerID,
		})
		require.NoError(t, err)
		require.Equal(t, 2, total)
		require.Len(t, files, 2)
	})
}

func TestByID(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *FileService, fileID uuid.UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, _ := setupFileService(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		opts := FileOpts{
			FileName: "test.txt",
			FileSize: len(content),
		}

		// Setup expectations for file creation
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := svc.Create(ctx, reader, opts)
		require.NoError(t, err)

		return ctx, svc, file.ID
	}

	t.Run("existing_file", func(t *testing.T) {
		ctx, svc, fileID := setup(t)

		file, err := svc.ByID(ctx, fileID)
		require.NoError(t, err)
		require.Equal(t, fileID, file.ID)
		require.Equal(t, "test.txt", file.Name)
	})

	t.Run("non_existent_file", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		nonExistentID := uuid.Must(uuid.NewV7())
		_, err := svc.ByID(ctx, nonExistentID)
		require.ErrorIs(t, err, sesc.ErrFileNotFound)
	})
}

func TestDownloadURL(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *FileService, storage *mocks.MockObjectStorage, fileID uuid.UUID, fileName string) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc, storage, _ = setupFileService(t)

		content := []byte("test file content")
		reader := bytes.NewReader(content)
		fileName = "test-download.txt"

		opts := FileOpts{
			FileName: fileName,
			FileSize: len(content),
		}

		// Setup expectations for file creation
		storage.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(len(content)))).Return(nil)

		file, err := svc.Create(ctx, reader, opts)
		require.NoError(t, err)

		return ctx, svc, storage, file.ID, fileName
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, storage, fileID, fileName := setup(t)

		expectedURL := "https://example.com/presigned-url"

		// Setup expectation for GetObjectURL
		storage.EXPECT().
			GetObjectURL(gomock.Any(), gomock.Any(), gomock.Eq(fileName), gomock.Eq(time.Hour)).
			Return(expectedURL, nil)

		url, err := svc.DownloadURL(ctx, fileID)
		require.NoError(t, err)
		require.Equal(t, expectedURL, url)
	})

	t.Run("non_existent_file", func(t *testing.T) {
		ctx, svc, _, _, _ := setup(t)

		nonExistentID := uuid.Must(uuid.NewV7())
		_, err := svc.DownloadURL(ctx, nonExistentID)
		require.ErrorIs(t, err, sesc.ErrFileNotFound)
	})

	t.Run("storage_error", func(t *testing.T) {
		ctx, svc, storage, fileID, fileName := setup(t)

		storageError := errors.New("storage service unavailable")

		// Setup expectation for GetObjectURL to return error
		storage.EXPECT().
			GetObjectURL(gomock.Any(), gomock.Any(), gomock.Eq(fileName), gomock.Eq(time.Hour)).
			Return("", storageError)

		_, err := svc.DownloadURL(ctx, fileID)
		require.Error(t, err)
		require.ErrorIs(t, err, storageError)
	})
}
