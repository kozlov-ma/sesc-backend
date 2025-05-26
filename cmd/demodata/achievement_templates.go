package main

import (
	"github.com/kozlov-ma/sesc-backend/apiclient/models"
)

// Map of achievement templates by group
var achievementTemplatesData = map[string][]AchievementTemplateData{
	"Показатель № 1": {
		{
			Name:        "1.1 Сопровождение мероприятий международного уровня",
			Description: "Сопровождение мероприятий международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "1.2 Сопровождение мероприятий всероссийского уровня",
			Description: "Сопровождение мероприятий всероссийского уровня",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "1.3 Сопровождение мероприятий регионального или муниципального уровня",
			Description: "Сопровождение мероприятий регионального или муниципального уровня",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "1.4 Сопровождение мероприятий локального уровня",
			Description: "Сопровождение мероприятий локального уровня",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "1.5 Другое",
			Description: "Другие виды сопровождения мероприятий",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 2": {
		{
			Name:        "2.1 Обеспечение участия в мероприятиях за пределами РФ",
			Description: "Обеспечение участия в мероприятиях, проводимых за пределами Российской Федерации",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "2.2 Обеспечение участия в мероприятиях всероссийского уровня",
			Description: "Обеспечение участия в мероприятиях всероссийского уровня",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "2.3 Обеспечение участия в мероприятиях за пределами Свердловской области",
			Description: "Обеспечение участия в мероприятиях, проводимых на территории РФ, но за пределами Свердловской области",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "2.4 Обеспечение участия в мероприятиях за пределами г. Екатеринбурга",
			Description: "Обеспечение участия в мероприятиях, проводимых на территории Свердловской области, но за пределами г. Екатеринбурга",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "2.5 Обеспечение участия в мероприятиях в г. Екатеринбурге",
			Description: "Обеспечение участия в мероприятиях, проводимых на территории г. Екатеринбурга",
			PointsLimit: 2,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "2.6 Другое",
			Description: "Другие виды обеспечения участия в мероприятиях",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 3": {
		{
			Name:        "3.1 Сопровождение дистанционного курса учебной дисциплины",
			Description: "Сопровождение дистанционного курса учебной дисциплины",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "3.2 Сопровождение дистанционного курса внеурочной деятельности",
			Description: "Сопровождение дистанционного курса внеурочной деятельности",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "3.3 Другое",
			Description: "Другие виды сопровождения дистанционных курсов",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 4.1": {
		{
			Name:        "4.1.1 Публикация монографии или коллективной монографии",
			Description: "Публикация монографии или коллективной монографии",
			PointsLimit: 10,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.2 Публикация в высокорейтинговом издании",
			Description: "Публикация в высокорейтинговом издании",
			PointsLimit: 10,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.3 Публикация учебного пособия РИНЦ",
			Description: "Публикация учебного пособия РИНЦ",
			PointsLimit: 9,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.4 Публикация РИНЦ в международных сборниках",
			Description: "Публикация РИНЦ в международных сборниках, в сборниках конференций с международным участием",
			PointsLimit: 8,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.5 Публикация РИНЦ в сборниках всероссийских конференций",
			Description: "Публикация РИНЦ в сборниках всероссийских конференций, в сборниках всероссийских с международным участием конференций, в сборниках межрегиональных конференций",
			PointsLimit: 7,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.6 Публикация в сборниках региональных конференций",
			Description: "Публикация в сборниках региональных конференций",
			PointsLimit: 6,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.7 Публикация в сборниках муниципальных конференций",
			Description: "Публикация в сборниках муниципальных конференций",
			PointsLimit: 5,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.1.8 Другое",
			Description: "Другие виды научных публикаций",
			PointsLimit: 40,
			Kind:        models.AchievementKindScientific,
		},
	},
	"Показатель № 4.2": {
		{
			Name:        "4.2.1 Участие в международной конференции с докладом",
			Description: "Участие в международной конференции с докладом без последующей публикации в сборниках трудов",
			PointsLimit: 8,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.2.2 Участие во всероссийской конференции с докладом",
			Description: "Участие во всероссийской конференции с докладом без последующей публикации в сборниках трудов",
			PointsLimit: 7,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.2.3 Участие в региональной конференции с докладом",
			Description: "Участие в региональной конференции с докладом без последующей публикации в сборниках трудов",
			PointsLimit: 4,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.2.4 Другое",
			Description: "Другие виды участия в конференциях с докладом",
			PointsLimit: 40,
			Kind:        models.AchievementKindScientific,
		},
	},
	"Показатель № 4.3": {
		{
			Name:        "4.3.1 Разработка методического пособия с официальным изданием",
			Description: "Разработка методического пособия с официальным изданием",
			PointsLimit: 10,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.3.2 Разработка методических рекомендаций, рабочих программ к новому курсу",
			Description: "Разработка методических рекомендаций, рабочих программ к новому (авторскому) курсу",
			PointsLimit: 8,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.3.3 Разработка курса для обучающихся СУНЦ",
			Description: "Разработка курса для обучающихся СУНЦ",
			PointsLimit: 8,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "4.3.4 Другое",
			Description: "Другие виды методических разработок",
			PointsLimit: 40,
			Kind:        models.AchievementKindScientific,
		},
	},
	"Показатель № 5": {
		{
			Name:        "5.1 Наличие статуса ведущего эксперта",
			Description: "Наличие статуса ведущего эксперта в региональных предметных комиссиях",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "5.2 Наличие статуса старшего эксперта",
			Description: "Наличие статуса старшего эксперта в региональных предметных комиссиях",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "5.3 Наличие статуса основного эксперта",
			Description: "Наличие статуса основного эксперта в региональных предметных комиссиях",
			PointsLimit: 2,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "5.4 Другое",
			Description: "Другие статусы в региональных предметных комиссиях",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 6": {
		{
			Name:        "6.1 Наличие статуса члена жюри заключительного этапа или ЦПМК",
			Description: "Наличие статуса члена жюри заключительного этапа или ЦПМК",
			PointsLimit: 10,
			Kind:        models.AchievementKindOlympiad,
		},
		{
			Name:        "6.2 Наличие статуса члена жюри регионального этапа или РПМК",
			Description: "Наличие статуса члена жюри регионального этапа или РПМК",
			PointsLimit: 7,
			Kind:        models.AchievementKindOlympiad,
		},
		{
			Name:        "6.3 Наличие статуса члена жюри муниципального этапа",
			Description: "Наличие статуса члена жюри муниципального этапа",
			PointsLimit: 5,
			Kind:        models.AchievementKindOlympiad,
		},
		{
			Name:        "6.4 Другое",
			Description: "Другие статусы в организационной структуре ВсОШ",
			PointsLimit: 40,
			Kind:        models.AchievementKindOlympiad,
		},
	},
	"Показатель № 7": {
		{
			Name:        "7.1 Средний балл у обучающихся в профильных классах не менее 4,65",
			Description: "Средний балл у обучающихся в профильных классах не менее 4,65",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "7.2 Средний балл у обучающихся в непрофильных классах не менее 4",
			Description: "Средний балл у обучающихся в непрофильных классах не менее 4",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "7.3 Другое",
			Description: "Другие результаты ОГЭ",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 8": {
		{
			Name:        "8.1 Обучающийся набравший 100 баллов",
			Description: "Обучающийся набравший 100 баллов (независимо от профиля класса)",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "8.2 Средний балл у обучающихся в профильных классах не менее 80",
			Description: "Средний балл у обучающихся в профильных классах не менее 80",
			PointsLimit: 8,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "8.3 Другое",
			Description: "Другие результаты ЕГЭ",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 9.1": {
		{
			Name:        "9.1.1 Олимпиада международного уровня",
			Description: "Результаты участия обучающихся в олимпиадах международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindOlympiad,
		},
		{
			Name:        "9.1.2 Олимпиада всероссийского уровня",
			Description: "Результаты участия обучающихся в олимпиадах всероссийского уровня",
			PointsLimit: 8,
			Kind:        models.AchievementKindOlympiad,
		},
		{
			Name:        "9.1.3 Олимпиада регионального уровня",
			Description: "Результаты участия обучающихся в олимпиадах регионального уровня",
			PointsLimit: 5,
			Kind:        models.AchievementKindOlympiad,
		},
		{
			Name:        "9.1.4 Другое",
			Description: "Другие результаты участия в олимпиадах",
			PointsLimit: 40,
			Kind:        models.AchievementKindOlympiad,
		},
	},
	"Показатель № 9.2": {
		{
			Name:        "9.2.1 Научно-практическая конференция международного уровня",
			Description: "Результаты участия обучающихся в научно-практических конференциях международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "9.2.2 Научно-практическая конференция всероссийского уровня",
			Description: "Результаты участия обучающихся в научно-практических конференциях всероссийского уровня",
			PointsLimit: 8,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "9.2.3 Научно-практическая конференция регионального уровня",
			Description: "Результаты участия обучающихся в научно-практических конференциях регионального уровня",
			PointsLimit: 5,
			Kind:        models.AchievementKindScientific,
		},
		{
			Name:        "9.2.4 Другое",
			Description: "Другие результаты участия в научно-практических конференциях",
			PointsLimit: 40,
			Kind:        models.AchievementKindScientific,
		},
	},
	"Показатель № 9.3": {
		{
			Name:        "9.3.1 Интеллектуальное соревнование международного уровня",
			Description: "Результаты участия обучающихся в интеллектуальных соревнованиях международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "9.3.2 Интеллектуальное соревнование всероссийского уровня",
			Description: "Результаты участия обучающихся в интеллектуальных соревнованиях всероссийского уровня",
			PointsLimit: 8,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "9.3.3 Интеллектуальное соревнование регионального уровня",
			Description: "Результаты участия обучающихся в интеллектуальных соревнованиях регионального уровня",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "9.3.4 Другое",
			Description: "Другие результаты участия в интеллектуальных соревнованиях",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 9.4": {
		{
			Name:        "9.4.1 Проектная деятельность международного уровня",
			Description: "Результаты проектной деятельности обучающихся международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "9.4.2 Проектная деятельность всероссийского уровня",
			Description: "Результаты проектной деятельности обучающихся всероссийского уровня",
			PointsLimit: 8,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "9.4.3 Проектная деятельность регионального уровня",
			Description: "Результаты проектной деятельности обучающихся регионального уровня",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "9.4.4 Другое",
			Description: "Другие результаты проектной деятельности",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 10.1": {
		{
			Name:        "10.1.1 Участие в спортивных соревнованиях международного уровня",
			Description: "Результаты участия обучающихся в спортивных соревнованиях международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "10.1.2 Участие в спортивных соревнованиях всероссийского уровня",
			Description: "Результаты участия обучающихся в спортивных соревнованиях всероссийского уровня",
			PointsLimit: 8,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "10.1.3 Участие в спортивных соревнованиях регионального уровня",
			Description: "Результаты участия обучающихся в спортивных соревнованиях регионального уровня",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "10.1.4 Другое",
			Description: "Другие результаты участия в спортивных соревнованиях",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 10.2": {
		{
			Name:        "10.2.1 Получение золотого значка ГТО",
			Description: "Выполнение обручающимися комплекса ГТО с получением золотого значка",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "10.2.2 Получение серебряного значка ГТО",
			Description: "Выполнение обручающимися комплекса ГТО с получением серебряного значка",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "10.2.3 Получение бронзового значка ГТО",
			Description: "Выполнение обручающимися комплекса ГТО с получением бронзового значка",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "10.2.4 Другое",
			Description: "Другие результаты выполнения комплекса ГТО",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 11": {
		{
			Name:        "11.1 Победа в конкурсе профессионального мастерства международного уровня",
			Description: "Победа в конкурсе профессионального мастерства международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "11.2 Призовое место в конкурсе профессионального мастерства международного уровня",
			Description: "Призовое место в конкурсе профессионального мастерства международного уровня",
			PointsLimit: 9,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "11.3 Победа в конкурсе профессионального мастерства всероссийского уровня",
			Description: "Победа в конкурсе профессионального мастерства всероссийского уровня",
			PointsLimit: 9,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "11.4 Призовое место в конкурсе профессионального мастерства всероссийского уровня",
			Description: "Призовое место в конкурсе профессионального мастерства всероссийского уровня",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "11.5 Победа в конкурсе профессионального мастерства регионального уровня",
			Description: "Победа в конкурсе профессионального мастерства регионального уровня",
			PointsLimit: 6,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "11.6 Призовое место в конкурсе профессионального мастерства регионального уровня",
			Description: "Призовое место в конкурсе профессионального мастерства регионального уровня",
			PointsLimit: 6,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "11.7 Другое",
			Description: "Другие результаты участия в конкурсах профессионального мастерства",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 12.1": {
		{
			Name:        "12.1.1 Победа в конкурсе педагогического мастерства",
			Description: "Победа в конкурсе педагогического мастерства СУНЦ УрФУ",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "12.1.2 Призовое место в конкурсе педагогического мастерства",
			Description: "Призовое место в конкурсе педагогического мастерства СУНЦ УрФУ",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "12.1.3 Финалист конкурса педагогического мастерства",
			Description: "Финалист конкурса педагогического мастерства СУНЦ УрФУ",
			PointsLimit: 1,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "12.1.4 Другое",
			Description: "Другие результаты участия в конкурсе педагогического мастерства",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 12.2": {
		{
			Name:        "12.2.1 Победа в конкурсе работников, реализующих программу воспитания",
			Description: "Победа в конкурсе работников, реализующих программу воспитания и социализации",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "12.2.2 Призовое место в конкурсе работников, реализующих программу воспитания",
			Description: "Призовое место в конкурсе работников, реализующих программу воспитания и социализации",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "12.2.3 Финалист конкурса работников, реализующих программу воспитания",
			Description: "Финалист конкурса работников, реализующих программу воспитания и социализации",
			PointsLimit: 1,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "12.2.4 Другое",
			Description: "Другие результаты участия в конкурсе работников",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 13": {
		{
			Name:        "13.1 Участие в развитии учебно-методической и нормативно-правовой базы",
			Description: "Участие работника в пределах трудовой функции и компетенции в развитии учебно-методической и нормативно-правовой базы",
			PointsLimit: 2,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "13.2 Другое",
			Description: "Другие виды участия в развитии нормативной базы",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 14": {
		{
			Name:        "14.1 Победа в конкурсе на получение гранта или привлечение пожертвования",
			Description: "Победа работника в конкурсе на получение гранта, привлечение пожертвования для реализации социально-значимых проектов",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "14.2 Другое",
			Description: "Другие результаты в получении грантов и привлечении пожертвований",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
	"Показатель № 15": {
		{
			Name:        "15.1 Победа в спортивных соревнованиях международного уровня",
			Description: "Победа работника в спортивных соревнованиях международного уровня",
			PointsLimit: 10,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.2 Призовое место в спортивных соревнованиях международного уровня",
			Description: "Призовое место работника в спортивных соревнованиях международного уровня",
			PointsLimit: 8,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.3 Победа в спортивных соревнованиях всероссийского уровня",
			Description: "Победа работника в спортивных соревнованиях всероссийского уровня",
			PointsLimit: 7,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.4 Призовое место в спортивных соревнованиях всероссийского уровня",
			Description: "Призовое место работника в спортивных соревнованиях всероссийского уровня",
			PointsLimit: 6,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.5 Получение золотого значка ГТО",
			Description: "Получение работником золотого значка ГТО",
			PointsLimit: 5,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.6 Получение серебряного значка ГТО",
			Description: "Получение работником серебряного значка ГТО",
			PointsLimit: 4,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.7 Получение бронзового значка ГТО",
			Description: "Получение работником бронзового значка ГТО",
			PointsLimit: 3,
			Kind:        models.AchievementKindDevelopment,
		},
		{
			Name:        "15.8 Другое",
			Description: "Другие спортивные достижения работника",
			PointsLimit: 40,
			Kind:        models.AchievementKindDevelopment,
		},
	},
}
