package sesc

// FileSearchOptions represents options for searching files.
type FileSearchOptions struct {
	Name           string
	OwnerID        *UUID
	Common         bool
	Offset         int
	Limit          int
	IncludeDeleted bool // If true, include deleted files in results
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

type ScheduleDeletionAllOptions struct {
	// No options needed - always schedules files for deletion
}

// ScheduleDeletionOptions represents options for scheduling file deletion.
type ScheduleDeletionOptions struct {
	FileIDs []UUID
}

// FileStats represents statistics about files.
type FileStats struct {
	TotalFiles           int `json:"total_files"`
	DeletedFiles         int `json:"deleted_files"`
	ScheduledForDeletion int `json:"scheduled_for_deletion"`
	ReadyForDeletion     int `json:"ready_for_deletion"`
	NotScheduled         int `json:"not_scheduled"`
}
