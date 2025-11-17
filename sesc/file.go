package sesc

// FileSearchOptions represents options for searching files.
type FileSearchOptions struct {
	Name    string
	OwnerID *string
	Common  bool
	Offset  int
	Limit   int
}

// FileCreateOptions represents options for creating a file.
type FileCreateOptions struct {
	FileName string
	FileSize int
	Common   bool
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
