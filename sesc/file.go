package sesc

import "github.com/kozlov-ma/sesc-backend/pkg/event"

// File represents a file stored in the system.
// If OwnerID is not set, the file is considered common and accessible to all users.
type File struct {
	ID          UUID
	OwnerID     *UUID
	S3ObjectKey string
	FileName    string
	FileSize    int
	DownloadURL string
}

func (f File) EventRecord() *event.Record {
	return event.Group(
		"id", f.ID,
		"owner_id", f.OwnerID,
		"s3_object_key", f.S3ObjectKey,
		"file_name", f.FileName,
		"file_size", f.FileSize,
		"download_url", f.DownloadURL,
	)
}

// FileSearchOptions represents options for searching files.
type FileSearchOptions struct {
	Name    string
	OwnerID *UUID
	Common  bool
	Offset  int
	Limit   int
}

// FileCreateOptions represents options for creating a file.
type FileCreateOptions struct {
	FileName string
	FileSize int
	OwnerID  *UUID
}

func (f FileCreateOptions) Validate() error {
	if f.FileName == "" {
		return ErrInvalidFileName
	}
	if f.FileSize <= 0 {
		return ErrInvalidFileSize
	}
	return nil
}
