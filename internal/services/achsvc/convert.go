package achsvc

import (
	"errors"
	"fmt"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// Helper function to convert achievement entity to domain model
func convertAchievementEntityToDomain(
	achievementEntity *ent.Achievement,
	rec *event.Record,
) (achievement.Achievement, error) {
	// Group parameters
	rec.Sub("params").Set("achievement_id", achievementEntity.ID)

	// Validate required relationships
	if achievementEntity.Edges.Owner == nil {
		rec.Add(events.Error, "achievement owner not loaded")
		return achievement.Achievement{}, errors.New("achievement owner not loaded")
	}
	if achievementEntity.Edges.Template == nil {
		rec.Add(events.Error, "achievement template not loaded")
		return achievement.Achievement{}, errors.New("achievement template not loaded")
	}

	// Result to be populated by operations
	var (
		owner     sesc.User
		template  achievement.Template
		documents []achievement.Document
		reviews   []achievement.Review
	)

	// Convert owner
	err := rec.Operation("convert_owner", func(opRec *event.Record) error {
		opRec.Sub("params").Set("owner_id", achievementEntity.Edges.Owner.ID)

		owner = sesc.User{
			ID:         achievementEntity.Edges.Owner.ID,
			FirstName:  achievementEntity.Edges.Owner.FirstName,
			LastName:   achievementEntity.Edges.Owner.LastName,
			MiddleName: achievementEntity.Edges.Owner.MiddleName,
		}

		// Add department if available
		if achievementEntity.Edges.Owner.Edges.Department != nil {
			dept := achievementEntity.Edges.Owner.Edges.Department
			opRec.Sub("department").Set(
				"id", dept.ID,
				"name", dept.Name,
			)

			owner.Department = sesc.Department{
				ID:          dept.ID,
				Name:        dept.Name,
				Description: dept.Description,
			}
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert template
	err = rec.Operation("convert_template", func(opRec *event.Record) error {
		tmpl := achievementEntity.Edges.Template
		opRec.Sub("params").Set(
			"template_id", tmpl.ID,
			"kind", tmpl.Kind,
			"points_limit", tmpl.PointsLimit,
		)

		template = achievement.Template{
			ID:          tmpl.ID,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			PointsLimit: tmpl.PointsLimit,
			GroupID:     tmpl.GroupID,
			Active:      tmpl.Active,
			Kind:        tmpl.Kind,
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert documents
	err = rec.Operation("convert_documents", func(opRec *event.Record) error {
		documents = make([]achievement.Document, 0)
		if achievementEntity.Edges.Documents != nil {
			opRec.Set("total_documents", len(achievementEntity.Edges.Documents))

			validDocs := 0
			for _, doc := range achievementEntity.Edges.Documents {
				if doc.Edges.File == nil {
					opRec.Add(events.Error, fmt.Sprintf("file not loaded for document %s", doc.ID))
					continue
				}
				documents = append(documents, achievement.Document{
					ID:     doc.ID,
					Name:   doc.Name,
					FileID: doc.Edges.File.ID,
				})
				validDocs++
			}
			opRec.Set("valid_documents", validDocs)
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Convert reviews
	err = rec.Operation("convert_reviews", func(opRec *event.Record) error {
		reviews = make([]achievement.Review, 0)
		if achievementEntity.Edges.Reviews != nil {
			opRec.Set("total_reviews", len(achievementEntity.Edges.Reviews))

			validReviews := 0
			for _, rev := range achievementEntity.Edges.Reviews {
				if rev.Edges.Reviewer == nil {
					opRec.Add(events.Error, fmt.Sprintf("reviewer not loaded for review %s", rev.ID))
					continue
				}

				reviewer := sesc.User{
					ID:         rev.Edges.Reviewer.ID,
					FirstName:  rev.Edges.Reviewer.FirstName,
					LastName:   rev.Edges.Reviewer.LastName,
					MiddleName: rev.Edges.Reviewer.MiddleName,
				}

				reviews = append(reviews, achievement.Review{
					ID:             rev.ID,
					From:           reviewer,
					PointsAssigned: rev.PointsAssigned,
					Comment:        rev.Comment,
				})
				validReviews++
			}
			opRec.Set("valid_reviews", validReviews)
		}
		return nil
	})
	if err != nil {
		return achievement.Achievement{}, err
	}

	// Create the achievement domain model
	result := achievement.Achievement{
		ID:        achievementEntity.ID,
		Owner:     owner,
		Template:  template,
		Status:    achievement.Status(achievementEntity.Status),
		Points:    achievementEntity.Points,
		Documents: documents,
		Reviews:   reviews,
	}

	// Record final result stats
	rec.Sub("result").Set(
		"status", result.Status,
		"points", result.Points,
		"documents_count", len(documents),
		"reviews_count", len(reviews),
	)

	return result, nil
}
