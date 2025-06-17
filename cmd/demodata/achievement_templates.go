package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementgroup"
)

// createAchievementTemplates creates achievement templates for demo data
func createAchievementTemplates(ctx context.Context, client *ent.Client) error {
	log.Println("Creating achievement templates...")

	// Create achievement groups
	groupIDs := make(map[string]uuid.UUID)
	for groupName := range achievementTemplatesData {
		// Check if the group already exists
		group, err := client.AchievementGroup.Query().
			Where(achievementgroup.Name(groupName)).
			Only(ctx)

		if err != nil {
			if !errors.Is(err, &ent.NotFoundError{}) {
				return fmt.Errorf("error querying group %s: %w", groupName, err)
			}

			// Create the group if it doesn't exist
			group, err = client.AchievementGroup.Create().
				SetName(groupName).
				SetDescription(fmt.Sprintf("Группа достижений: %s", groupName)).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("error creating group %s: %w", groupName, err)
			}
			log.Printf("Created achievement group: %s", groupName)
		} else {
			log.Printf("Found existing achievement group: %s", groupName)
		}

		groupIDs[groupName] = group.ID
	}

	// Create achievement templates
	templatesCreated := 0
	for groupName, templates := range achievementTemplatesData {
		groupID, ok := groupIDs[groupName]
		if !ok {
			return fmt.Errorf("group ID not found for %s", groupName)
		}

		for _, template := range templates {
			// Create the template
			_, err := client.AchievementTemplate.Create().
				SetName(template.Name).
				SetDescription(template.Description).
				SetPointsLimit(template.PointsLimit).
				SetGroupID(groupID).
				SetReviewerRole(template.ReviewerRole).
				Save(ctx)

			if err != nil {
				// Ignore duplicate entries
				if errors.Is(err, &ent.ConstraintError{}) {
					log.Printf("Template already exists: %s", template.Name)
					continue
				}
				return fmt.Errorf("error creating template %s: %w", template.Name, err)
			}

			templatesCreated++
		}
	}

	log.Printf("Created %d achievement templates", templatesCreated)
	return nil
}
