package main

import (
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// AchievementTemplateData represents a single achievement template for demo data
type AchievementTemplateData struct {
	Name         string
	Description  string
	PointsLimit  int
	ReviewerRole achievement.ReviewerRole
}

// Map of achievement templates by group
var achievementTemplatesData = map[string][]AchievementTemplateData{
	"Показатель № 1": {
		{
			Name:         "1.1 Сопровождение мероприятий международного уровня",
			Description:  "Сопровождение мероприятий международного уровня",
			PointsLimit:  10,
			ReviewerRole: achievement.ReviewerRole(sesc.DevelopmentDeputy),
		},
		{
			Name:         "1.2 Сопровождение мероприятий всероссийского уровня",
			Description:  "Сопровождение мероприятий всероссийского уровня",
			PointsLimit:  7,
			ReviewerRole: achievement.ReviewerRole(sesc.DevelopmentDeputy),
		},
		{
			Name:         "1.3 Сопровождение мероприятий регионального или муниципального уровня",
			Description:  "Сопровождение мероприятий регионального или муниципального уровня",
			PointsLimit:  5,
			ReviewerRole: achievement.ReviewerRole(sesc.DevelopmentDeputy),
		},
		{
			Name:         "1.4 Сопровождение мероприятий локального уровня",
			Description:  "Сопровождение мероприятий локального уровня",
			PointsLimit:  3,
			ReviewerRole: achievement.ReviewerRole(sesc.DevelopmentDeputy),
		},
	},
	"Показатель № 2": {
		{
			Name:         "2.1 Публикация международного уровня",
			Description:  "Публикация международного уровня",
			PointsLimit:  10,
			ReviewerRole: achievement.ReviewerRole(sesc.ScientificDeputy),
		},
		{
			Name:         "2.2 Публикация всероссийского уровня",
			Description:  "Публикация всероссийского уровня",
			PointsLimit:  7,
			ReviewerRole: achievement.ReviewerRole(sesc.ScientificDeputy),
		},
		{
			Name:         "2.3 Публикация регионального или муниципального уровня",
			Description:  "Публикация регионального или муниципального уровня",
			PointsLimit:  5,
			ReviewerRole: achievement.ReviewerRole(sesc.ScientificDeputy),
		},
		{
			Name:         "2.4 Публикация локального уровня",
			Description:  "Публикация локального уровня",
			PointsLimit:  3,
			ReviewerRole: achievement.ReviewerRole(sesc.ScientificDeputy),
		},
	},
	"Показатель № 3": {
		{
			Name:         "3.1 Участие в международных олимпиадах",
			Description:  "Участие в международных олимпиадах",
			PointsLimit:  10,
			ReviewerRole: achievement.ReviewerRole(sesc.OlympiadDeputy),
		},
		{
			Name:         "3.2 Участие во всероссийских олимпиадах",
			Description:  "Участие во всероссийских олимпиадах",
			PointsLimit:  7,
			ReviewerRole: achievement.ReviewerRole(sesc.OlympiadDeputy),
		},
		{
			Name:         "3.3 Участие в региональных олимпиадах",
			Description:  "Участие в региональных олимпиадах",
			PointsLimit:  5,
			ReviewerRole: achievement.ReviewerRole(sesc.OlympiadDeputy),
		},
		{
			Name:         "3.4 Участие в локальных олимпиадах",
			Description:  "Участие в локальных олимпиадах",
			PointsLimit:  3,
			ReviewerRole: achievement.ReviewerRole(sesc.OlympiadDeputy),
		},
	},
	"Показатель № 4": {
		{
			Name:         "4.1 Организация международных конференций",
			Description:  "Организация международных конференций",
			PointsLimit:  10,
			ReviewerRole: achievement.ReviewerRole(sesc.AcademicDirector),
		},
		{
			Name:         "4.2 Организация всероссийских конференций",
			Description:  "Организация всероссийских конференций",
			PointsLimit:  7,
			ReviewerRole: achievement.ReviewerRole(sesc.AcademicDirector),
		},
		{
			Name:         "4.3 Организация региональных конференций",
			Description:  "Организация региональных конференций",
			PointsLimit:  5,
			ReviewerRole: achievement.ReviewerRole(sesc.AcademicDirector),
		},
		{
			Name:         "4.4 Организация локальных конференций",
			Description:  "Организация локальных конференций",
			PointsLimit:  3,
			ReviewerRole: achievement.ReviewerRole(sesc.AcademicDirector),
		},
	},
}
