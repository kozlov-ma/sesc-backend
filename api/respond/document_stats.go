package respond

import "github.com/kozlov-ma/sesc-backend/internal/services/achsvc"

// DocumentStats represents the response for document statistics
type DocumentStats struct {
	TotalFiles           int    `json:"totalFiles"`
	DeletedFiles         int    `json:"deletedFiles"`
	ScheduledForDeletion int    `json:"scheduledForDeletion"`
	ReadyForDeletion     int    `json:"readyForDeletion"`
	NotScheduled         int    `json:"notScheduled"`
	DeletionDelay        string `json:"deletionDelay"`
}

func WithDocumentStats(stats *achsvc.DocumentStats, deletionDelay string) *DocumentStats {
	return &DocumentStats{
		TotalFiles:           stats.TotalDocuments,
		DeletedFiles:         stats.DeletedDocuments,
		ScheduledForDeletion: stats.ScheduledForDeletion,
		ReadyForDeletion:     stats.ReadyForDeletion,
		NotScheduled:         stats.NotScheduled,
		DeletionDelay:        deletionDelay,
	}
}
