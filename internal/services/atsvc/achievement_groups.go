package atsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementgroup"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// AchievementGroups gets all achievement groups with optional filtering.
func (s *ATS) AchievementGroups(
	ctx context.Context,
	options AchievementGroupSearchOptions,
) ([]AchievementGroup, error) {
	rec := event.Get(ctx).Sub("sesc/achievement_groups")

	query := s.client.AchievementGroup.Query()

	// Apply filters
	if !options.ShowInactive {
		query = query.Where(achievementgroup.Active(true))
	}

	if options.Search != "" {
		searchTerm := strings.ToLower(options.Search)
		query = query.Where(achievementgroup.Or(
			achievementgroup.NameContainsFold(searchTerm),
			achievementgroup.DescriptionContainsFold(searchTerm),
		))
	}

	groups, err := query.All(ctx)
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to query achievement groups: %w", err))
		return nil, err
	}

	result := make([]AchievementGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, AchievementGroup{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Active:      g.Active,
		})
	}

	rec.Add("groups_count", len(result))
	return result, nil
}
