package respond

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

type File struct {
	ID          uuid.UUID  `json:"id"`
	OwnerID     *uuid.UUID `json:"ownerId,omitzero"`
	S3ObjectKey *string    `json:"s3ObjectKey,omitzero"`
	FileName    string     `json:"fileName"`
	FileSize    int        `json:"fileSize"`
}

type Files struct {
	Files []*File `json:"files"`
	Total int     `json:"total"`
}

type DeleteAllFilesResponse struct {
	Message   string `json:"message"`
	MarkOnly  bool   `json:"markOnly"`
	Timestamp int64  `json:"timestamp"`
}

func WithFile(f *ent.File) *File {
	return &File{
		ID:          f.ID,
		OwnerID:     f.OwnerID,
		S3ObjectKey: f.S3ObjectKey,
		FileName:    f.Name,
		FileSize:    f.Size,
	}
}

func WithFiles(ff ent.Files, total int) Files {
	fi := make([]*File, len(ff))
	for i, f := range ff {
		fi[i] = WithFile(f)
	}

	return Files{
		Files: fi,
		Total: total,
	}
}
