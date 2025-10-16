package respond

import (
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

// Achievement represents a single achievement response
type Achievement struct {
	ID           uuid.UUID  `json:"id"           example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	OwnerID      uuid.UUID  `json:"ownerId"      example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	TemplateID   uuid.UUID  `json:"templateId"   example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	TemplateName string     `json:"templateName" example:"регионального уровня"                 validate:"required"`
	Status       string     `json:"status"       example:"draft"                                validate:"required"`
	Points       int        `json:"points"       example:"7"                                    validate:"required"`
	MaxPoints    int        `json:"maxPoints"    example:"10"                                   validate:"required"`
	Documents    []Document `json:"documents"                                                   validate:"required"`
	Reviews      []Review   `json:"reviews"                                                     validate:"required"`
}

// Document represents a document response (used in Achievement responses)
type Document struct {
	ID     uuid.UUID  `json:"id"               example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name   string     `json:"name"             example:"Publication proof"                    validate:"required"`
	FileID *uuid.UUID `json:"fileId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status string     `json:"status"           example:"active"                               validate:"required"`
}

// Review represents a review response
type Review struct {
	ID             uuid.UUID `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	ReviewerID     uuid.UUID `json:"reviewerId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	PointsAssigned int       `json:"pointsAssigned" example:"8"                                    validate:"required"`
	Comment        string    `json:"comment"        example:"Good job, but could be better"        validate:"omitempty"`
}

// Achievements represents a paginated list of achievements
type Achievements struct {
	Achievements []Achievement `json:"achievements" validate:"required"`
	Total        int           `json:"total"        validate:"required"`
}

// WithAchievement converts an ent.Achievement to a response (assumes edges are loaded)
func WithAchievement(ach *ent.Achievement) Achievement {
	template := ach.Edges.Template

	response := Achievement{
		ID:           ach.ID,
		OwnerID:      ach.OwnerID,
		TemplateID:   template.ID,
		TemplateName: template.Name,
		Status:       ach.Status,
		Points:       ach.Points,
		MaxPoints:    template.PointsLimit,
		Documents:    make([]Document, 0, len(ach.Edges.Documents)),
		Reviews:      make([]Review, 0, len(ach.Edges.Reviews)),
	}

	// Convert documents
	for _, doc := range ach.Edges.Documents {
		response.Documents = append(response.Documents, Document{
			ID:     doc.ID,
			Name:   doc.Name,
			FileID: doc.FileID,
			Status: doc.Status,
		})
	}

	// Convert reviews
	for _, rev := range ach.Edges.Reviews {
		response.Reviews = append(response.Reviews, Review{
			ID:             rev.ID,
			ReviewerID:     rev.ReviewerID,
			PointsAssigned: rev.PointsAssigned,
			Comment:        rev.Comment,
		})
	}

	return response
}

// WithAchievements converts a slice of ent.Achievement to a paginated response
func WithAchievements(achievements ent.Achievements, totalCount int) Achievements {
	items := make([]Achievement, 0, len(achievements))
	for _, ach := range achievements {
		items = append(items, WithAchievement(ach))
	}

	return Achievements{
		Achievements: items,
		Total:        totalCount,
	}
}
