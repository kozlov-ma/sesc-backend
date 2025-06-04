package respond

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
)

// Achievement represents a single achievement response
type Achievement struct {
	ID           uuid.UUID          `json:"id"           example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	OwnerID      uuid.UUID          `json:"ownerId"      example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	OwnerName    string             `json:"ownerName"    example:"Иванов Иван Иванович"                 validate:"required"`
	TemplateID   uuid.UUID          `json:"templateId"   example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	TemplateName string             `json:"templateName" example:"регионального уровня"                 validate:"required"`
	Status       achievement.Status `json:"status"       example:"draft"                                validate:"required"`
	Points       int                `json:"points"       example:"10"                                   validate:"required"`
	Documents    []Document         `json:"documents"                                                   validate:"required"`
	Reviews      []Review           `json:"reviews"                                                     validate:"required"`
}

// Document represents a document response (used in Achievement responses)
type Document struct {
	ID     uuid.UUID `json:"id"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Name   string    `json:"name"   example:"Publication proof"                    validate:"required"`
	FileID uuid.UUID `json:"fileId" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
}

// Review represents a review response
type Review struct {
	ID             uuid.UUID `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	ReviewerID     uuid.UUID `json:"reviewerId"     example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	ReviewerName   string    `json:"reviewerName"   example:"Петров Петр Петрович"                 validate:"required"`
	PointsAssigned int       `json:"pointsAssigned" example:"8"                                    validate:"required"`
	Comment        string    `json:"comment"        example:"Good job, but could be better"        validate:"omitempty"`
}

// Achievements represents a paginated list of achievements
type Achievements struct {
	Items      []Achievement `json:"items"      validate:"required"`
	TotalCount int           `json:"totalCount" validate:"required"`
	Offset     int           `json:"offset"     validate:"required"`
	Limit      int           `json:"limit"      validate:"required"`
}

// WithAchievement converts an ent.Achievement to a response (assumes edges are loaded)
func WithAchievement(ach *ent.Achievement) Achievement {
	owner := ach.Edges.Owner
	template := ach.Edges.Template

	response := Achievement{
		ID:           ach.ID,
		OwnerID:      owner.ID,
		OwnerName:    fmt.Sprintf("%s %s %s", owner.LastName, owner.FirstName, owner.MiddleName),
		TemplateID:   template.ID,
		TemplateName: template.Name,
		Status:       ach.Status,
		Points:       ach.Points,
		Documents:    make([]Document, 0, len(ach.Edges.Documents)),
		Reviews:      make([]Review, 0, len(ach.Edges.Reviews)),
	}

	// Convert documents
	for _, doc := range ach.Edges.Documents {
		response.Documents = append(response.Documents, Document{
			ID:     doc.ID,
			Name:   doc.Name,
			FileID: doc.FileID,
		})
	}

	// Convert reviews
	for _, rev := range ach.Edges.Reviews {
		reviewer := rev.Edges.Reviewer
		response.Reviews = append(response.Reviews, Review{
			ID:             rev.ID,
			ReviewerID:     reviewer.ID,
			ReviewerName:   fmt.Sprintf("%s %s %s", reviewer.LastName, reviewer.FirstName, reviewer.MiddleName),
			PointsAssigned: rev.PointsAssigned,
			Comment:        rev.Comment,
		})
	}

	return response
}

// WithAchievements converts a slice of ent.Achievement to a paginated response
func WithAchievements(achievements ent.Achievements, totalCount, offset, limit int) Achievements {
	items := make([]Achievement, 0, len(achievements))
	for _, ach := range achievements {
		items = append(items, WithAchievement(ach))
	}

	return Achievements{
		Items:      items,
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}
}

// UsersWithAchievements represents a list of users with achievement counts
type UsersWithAchievements struct {
	Items      []UserWithAchievements `json:"items"      validate:"required"`
	TotalCount int                    `json:"totalCount" validate:"required"`
	Offset     int                    `json:"offset"     validate:"required"`
	Limit      int                    `json:"limit"      validate:"required"`
}

// UserWithAchievements represents a user with achievement summary
type UserWithAchievements struct {
	ID         string `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	FirstName  string `json:"firstName"  example:"Иван"                                 validate:"required"`
	LastName   string `json:"lastName"   example:"Иванов"                               validate:"required"`
	MiddleName string `json:"middleName" example:"Иванович"`
	Role       string `json:"role"       example:"teacher"                              validate:"required"`
}

// WithUsersWithAchievements converts a slice of ent.User to a paginated response
func WithUsersWithAchievements(users ent.Users, totalCount, offset, limit int) UsersWithAchievements {
	items := make([]UserWithAchievements, 0, len(users))
	for _, u := range users {
		items = append(items, UserWithAchievements{
			ID:         u.ID.String(),
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			MiddleName: u.MiddleName,
			Role:       u.Role.String(),
		})
	}

	return UsersWithAchievements{
		Items:      items,
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}
}
