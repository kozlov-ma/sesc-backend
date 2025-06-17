//nolint:sloglint // this is a script.
package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/kozlov-ma/sesc-backend/apiclient/client"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/achievement_groups"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/achievement_templates"
	"github.com/kozlov-ma/sesc-backend/apiclient/models"
)

// processAchievementGroups ensures all required achievement groups exist
func processAchievementGroups(
	apiClient *client.Apiclient,
	authInfo runtime.ClientAuthInfoWriter,
) (map[string]*models.RespondAchievementGroup, error) {
	// Get existing achievement groups
	groupsResp, err := apiClient.AchievementGroups.GetAchievementGroups(
		achievement_groups.NewGetAchievementGroupsParams(),
		authInfo,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement groups: %w", err)
	}

	// Create map of existing groups by name
	existingGroups := make(map[string]*models.RespondAchievementGroup)
	for _, group := range groupsResp.Payload {
		existingGroups[*group.Name] = group
	}

	// Create map to return
	groupMap := make(map[string]*models.RespondAchievementGroup)

	// Create missing groups
	for _, groupData := range achievementGroupsData {
		if group, exists := existingGroups[groupData.Name]; exists {
			groupMap[*group.ID] = group
			slog.Info("Achievement group already exists", "name", *group.Name, "id", *group.ID)
		} else {
			// Create new group
			createParams := achievement_groups.NewPostAchievementGroupsParams()
			createParams.SetRequest(&models.ParamCreateAchievementGroupRequest{
				Name:        &groupData.Name,
				Description: &groupData.Description,
			})

			createResp, err := apiClient.AchievementGroups.PostAchievementGroups(createParams, authInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to create achievement group %s: %w", groupData.Name, err)
			}

			groupMap[*createResp.Payload.ID] = createResp.Payload
			slog.Info("Created achievement group", "name", *createResp.Payload.Name, "id", *createResp.Payload.ID)
		}
	}

	// Process templates for each group
	for groupID, group := range groupMap {
		// Extract the group key from the full name (e.g., "Показатель № 1" from "Показатель № 1 ...")
		groupKey := extractGroupKey(*group.Name)

		// Get templates for this group
		if templates, exists := achievementTemplatesData[groupKey]; exists {
			err := processAchievementTemplates(apiClient, authInfo, groupID, templates)
			if err != nil {
				return nil, fmt.Errorf("failed to process templates for group %s: %w", *group.Name, err)
			}
		}
	}

	return groupMap, nil
}

// processAchievementTemplates ensures all required templates exist for a group
func processAchievementTemplates(
	apiClient *client.Apiclient,
	authInfo runtime.ClientAuthInfoWriter,
	groupID string,
	templates []AchievementTemplateData,
) error {
	// Get existing templates for this group
	templatesResp, err := apiClient.AchievementTemplates.GetAchievementTemplates(
		achievement_templates.NewGetAchievementTemplatesParams(),
		authInfo,
	)
	if err != nil {
		return fmt.Errorf("failed to get achievement templates: %w", err)
	}

	// Create map of existing templates by name
	existingTemplates := make(map[string]*models.RespondAchievementTemplate)
	for _, template := range templatesResp.Payload {
		if *template.GroupID == groupID {
			existingTemplates[*template.Name] = template
		}
	}

	// Create missing templates
	for _, templateData := range templates {
		if _, exists := existingTemplates[templateData.Name]; exists {
			slog.Info("Achievement template already exists", "name", templateData.Name)
		} else {
			// Create new template
			createParams := achievement_templates.NewPostAchievementTemplatesParams()
			createParams.SetRequest(&models.ParamCreateAchievementTemplateRequest{
				Name:        &templateData.Name,
				Description: &templateData.Description,
				GroupID:     &groupID,
				PointsLimit: &templateData.PointsLimit,
			})

			// Set the kind
			createParams.Request.ReviewerRole = &templateData.ReviewerRole

			_, err := apiClient.AchievementTemplates.PostAchievementTemplates(createParams, authInfo)
			if err != nil {
				return fmt.Errorf("failed to create achievement template %s: %w", templateData.Name, err)
			}

			slog.Info("Created achievement template", "name", templateData.Name)
		}
	}

	return nil
}

// extractGroupKey extracts the group key from the full name
func extractGroupKey(fullName string) string {
	// Extract "Показатель № X" from the full name
	parts := strings.Split(fullName, " ")
	if len(parts) >= 3 {
		return strings.Join(parts[0:3], " ")
	}
	return fullName
}
