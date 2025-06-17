package migrations

import (
	"context"
	"fmt"
	"log"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementtemplate"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// MigrateKindToReviewerRole converts achievement template kind field to reviewer_role
func MigrateKindToReviewerRole(ctx context.Context, client *ent.Client) error {
	log.Println("Starting migration from achievement template kind to reviewer_role...")

	// Get all achievement templates
	templates, err := client.AchievementTemplate.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query templates: %w", err)
	}

	log.Printf("Found %d templates to migrate", len(templates))

	// For each template, update the reviewer_role based on the kind
	for _, template := range templates {
		var reviewerRole sesc.Role

		// Get the kind value - in case it's not accessible directly
		kindStr := ""
		// Try to get the kind value from the template
		kindValue, err := template.Value("kind")
		if err != nil {
			// If we can't get it directly, try a query
			t, err := client.AchievementTemplate.Query().
				Where(achievementtemplate.ID(template.ID)).
				Only(ctx)
			if err != nil {
				return fmt.Errorf("failed to query template %s (%s): %w", template.Name, template.ID, err)
			}

			// Extract kind as string since we may not have the Kind type available
			kindValue, err = t.Value("kind")
			if err != nil {
				log.Printf("Warning: Could not get kind for template %s (%s), defaulting to AcademicDirector",
					template.Name, template.ID)
				reviewerRole = sesc.AcademicDirector
			} else {
				kindStr, _ = kindValue.(string)
			}
		} else {
			kindStr, _ = kindValue.(string)
		}

		// Map the old kind string to the appropriate reviewer role
		switch kindStr {
		case "olympiad":
			reviewerRole = sesc.OlympiadDeputy
		case "development":
			reviewerRole = sesc.DevelopmentDeputy
		case "scientific":
			reviewerRole = sesc.ScientificDeputy
		default:
			log.Printf("Warning: Unknown kind %s for template %s (%s), defaulting to AcademicDirector",
				kindStr, template.Name, template.ID)
			reviewerRole = sesc.AcademicDirector
		}

		// Update the template with the new reviewer_role
		_, err = client.AchievementTemplate.UpdateOne(template).
			SetReviewerRole(achievement.ReviewerRole(reviewerRole)).
			Save(ctx)

		if err != nil {
			// Log the error but continue with other templates
			log.Printf("Error updating template %s (%s): %v", template.Name, template.ID, err)
			continue
		}

		log.Printf("Migrated template %s (%s): %s -> %s",
			template.Name, template.ID, kindStr, sesc.Role(reviewerRole).String())
	}

	log.Println("Migration completed successfully")
	return nil
}
