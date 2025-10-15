package respond

import "github.com/kozlov-ma/sesc-backend/sesc"

// DocumentStats represents the response for document statistics
type DocumentStats struct {
	TotalFiles           int    `json:"totalFiles"`
	DeletedFiles         int    `json:"deletedFiles"`
	ScheduledForDeletion int    `json:"scheduledForDeletion"`
	ReadyForDeletion     int    `json:"readyForDeletion"`
	NotScheduled         int    `json:"notScheduled"`
	DeletionDelay        string `json:"deletionDelay"`
}

// WithDocumentStats creates a response with document statistics
func WithDocumentStats(stats *sesc.FileStats, deletionDelay string) *DocumentStats {
	return &DocumentStats{
		TotalFiles:           stats.TotalFiles,
		DeletedFiles:         stats.DeletedFiles,
		ScheduledForDeletion: stats.ScheduledForDeletion,
		ReadyForDeletion:     stats.ReadyForDeletion,
		NotScheduled:         stats.NotScheduled,
		DeletionDelay:        deletionDelay,
	}
}
