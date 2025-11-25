package companyservice

import (
	"github.com/kozlov-ma/sesc-backend/company"
)

// Department kaf_* IDs used for demo data
const (
	departmentKafBio     = "kaf_bio"
	departmentKafFil     = "kaf_fil"
	departmentKafForeign = "kaf_foreign"
	departmentKafGum     = "kaf_gum"
	departmentKafInf     = "kaf_inf"
	departmentKafMath    = "kaf_math"
	departmentKafPhys    = "kaf_phys"
	departmentKafSport   = "kaf_sport"
)

func NewDemo() S {
	// Departments used by demo users
	deps := []company.Department{
		{ID: departmentKafBio, Name: "Кафедра Биологии", Description: "Кафедра биологии, основана давно"},
		{ID: departmentKafFil, Name: "Кафедра Филологии", Description: "Кафедра филологии, основана давно"},
		{
			ID:          departmentKafForeign,
			Name:        "Кафедра Иностранных Языков",
			Description: "Кафедра иностранных языков, основана давно",
		},
		{
			ID:          departmentKafGum,
			Name:        "Кафедра Гуманитарных Наук",
			Description: "Кафедра гуманитарных наук, основана давно",
		},
		{ID: departmentKafInf, Name: "Кафедра Информатики", Description: "Кафедра информатики, основана давно"},
		{ID: departmentKafMath, Name: "Кафедра Математики", Description: "Кафедра математики, основана давно"},
		{ID: departmentKafPhys, Name: "Кафедра Физики", Description: "Кафедра физики, основана давно"},
		{
			ID:          departmentKafSport,
			Name:        "Кафедра Физической Культуры, основана давно",
			Description: "Кафедра физической культуры, основана давно",
		},
	}

	// Build users slice and add canonical admin
	users := make([]company.User, 0)

	// Add a canonical admin matching other parts of the repo
	users = append(users, company.User{
		ID:           "kozlovma",
		FullName:     "Козлов Михаил Александрович",
		Roles:        []company.Role{company.Admin, company.DevelopmentDeputy, company.Teacher, company.Dephead},
		DepartmentID: departmentKafInf,
	})

	// Build explicit list per-department: for each department 5 teachers + 1 dephead
	users = append(
		users,
		// kaf_bio
		company.User{
			ID:           "bio_teacher_01",
			FullName:     "Алексей Морозов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafBio,
		},
		company.User{
			ID:           "bio_teacher_02",
			FullName:     "Ольга Павлова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafBio,
		},
		company.User{
			ID:           "bio_teacher_03",
			FullName:     "Ирина Кузнецова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafBio,
		},
		company.User{
			ID:           "bio_teacher_04",
			FullName:     "Максим Соколов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafBio,
		},
		company.User{
			ID:           "bio_teacher_05",
			FullName:     "Наталья Васильева",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafBio,
		},
		company.User{
			ID:           "bio_dephead",
			FullName:     "Дмитрий Ефимов",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafBio,
		},

		// kaf_fil
		company.User{
			ID:           "fil_teacher_01",
			FullName:     "Екатерина Новикова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafFil,
		},
		company.User{
			ID:           "fil_teacher_02",
			FullName:     "Сергей Орлов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafFil,
		},
		company.User{
			ID:           "fil_teacher_03",
			FullName:     "Анна Смирнова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafFil,
		},
		company.User{
			ID:           "fil_teacher_04",
			FullName:     "Роман Лебедев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafFil,
		},
		company.User{
			ID:           "fil_teacher_05",
			FullName:     "Виктория Киселева",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafFil,
		},
		company.User{
			ID:           "fil_dephead",
			FullName:     "Людмила Фролова",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafFil,
		},

		// kaf_foreign
		company.User{
			ID:           "for_teacher_01",
			FullName:     "Мария Белова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafForeign,
		},
		company.User{
			ID:           "for_teacher_02",
			FullName:     "Игорь Михайлов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafForeign,
		},
		company.User{
			ID:           "for_teacher_03",
			FullName:     "Татьяна Волкова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafForeign,
		},
		company.User{
			ID:           "for_teacher_04",
			FullName:     "Павел Егоров",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafForeign,
		},
		company.User{
			ID:           "for_teacher_05",
			FullName:     "Светлана Никитина",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafForeign,
		},
		company.User{
			ID:           "for_dephead",
			FullName:     "Николай Сидоров",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafForeign,
		},

		// kaf_gum
		company.User{
			ID:           "gum_teacher_01",
			FullName:     "Андрей Семенов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafGum,
		},
		company.User{
			ID:           "gum_teacher_02",
			FullName:     "Марина Романова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafGum,
		},
		company.User{
			ID:           "gum_teacher_03",
			FullName:     "Юлия Морозова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafGum,
		},
		company.User{
			ID:           "gum_teacher_04",
			FullName:     "Олег Григорьев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafGum,
		},
		company.User{
			ID:           "gum_teacher_05",
			FullName:     "Антонина Крылова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafGum,
		},
		company.User{
			ID:           "gum_dephead",
			FullName:     "Валерий Захаров",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafGum,
		},

		// kaf_inf
		company.User{
			ID:           "inf_teacher_01",
			FullName:     "Денис Воронов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafInf,
		},
		company.User{
			ID:           "inf_teacher_02",
			FullName:     "Ксения Миронова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafInf,
		},
		company.User{
			ID:           "inf_teacher_03",
			FullName:     "Петр Андреев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafInf,
		},
		company.User{
			ID:           "inf_teacher_04",
			FullName:     "Елена Федорова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafInf,
		},
		company.User{
			ID:           "inf_teacher_05",
			FullName:     "Илья Попов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafInf,
		},
		company.User{
			ID:           "inf_dephead",
			FullName:     "Анатолий Мартынов",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafInf,
		},

		// kaf_math
		company.User{
			ID:           "math_teacher_01",
			FullName:     "Надежда Лукьянова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafMath,
		},
		company.User{
			ID:           "math_teacher_02",
			FullName:     "Григорий Шестаков",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafMath,
		},
		company.User{
			ID:           "math_teacher_03",
			FullName:     "Людмила Козлова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafMath,
		},
		company.User{
			ID:           "math_teacher_04",
			FullName:     "Юрий Павленко",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafMath,
		},
		company.User{
			ID:           "math_teacher_05",
			FullName:     "Оксана Воронова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafMath,
		},
		company.User{
			ID:           "math_dephead",
			FullName:     "Станислав Руднев",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafMath,
		},

		// kaf_phys
		company.User{
			ID:           "phys_teacher_01",
			FullName:     "Роман Дмитриев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafPhys,
		},
		company.User{
			ID:           "phys_teacher_02",
			FullName:     "Алина Николаева",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafPhys,
		},
		company.User{
			ID:           "phys_teacher_03",
			FullName:     "Степан Ковалев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafPhys,
		},
		company.User{
			ID:           "phys_teacher_04",
			FullName:     "Елена Гусева",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafPhys,
		},
		company.User{
			ID:           "phys_teacher_05",
			FullName:     "Михаил Сидоренко",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafPhys,
		},
		company.User{
			ID:           "phys_dephead",
			FullName:     "Игорь Власов",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafPhys,
		},

		// kaf_sport
		company.User{
			ID:           "sport_teacher_01",
			FullName:     "Виктор Антонов",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafSport,
		},
		company.User{
			ID:           "sport_teacher_02",
			FullName:     "Ольга Сергеева",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafSport,
		},
		company.User{
			ID:           "sport_teacher_03",
			FullName:     "Дмитрий Кудрявцев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafSport,
		},
		company.User{
			ID:           "sport_teacher_04",
			FullName:     "Наталья Романова",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafSport,
		},
		company.User{
			ID:           "sport_teacher_05",
			FullName:     "Артем Зайцев",
			Roles:        []company.Role{company.Teacher},
			DepartmentID: departmentKafSport,
		},
		company.User{
			ID:           "sport_dephead",
			FullName:     "Сергей Петров",
			Roles:        []company.Role{company.Dephead},
			DepartmentID: departmentKafSport,
		},
	)

	// Unattached global users: directors, chief economist, global admin
	users = append(
		users,
		company.User{
			ID:       "global_scientific",
			FullName: "Александр Новиков",
			Roles:    []company.Role{company.ScientificDeputy},
		},
		company.User{
			ID:       "global_olympiad",
			FullName: "Евгения Кузнецова",
			Roles:    []company.Role{company.OlympiadDeputy},
		},
		company.User{
			ID:       "global_development",
			FullName: "Михаил Ершов",
			Roles:    []company.Role{company.DevelopmentDeputy},
		},
		company.User{
			ID:       "global_academic",
			FullName: "Ирина Коваленко",
			Roles:    []company.Role{company.AcademicDirector},
		},
		company.User{
			ID:       "global_economist",
			FullName: "Ольга Тимофеева",
			Roles:    []company.Role{company.ChiefEconomist},
		},
		company.User{ID: "global_admin", FullName: "Наталья Чернышева", Roles: []company.Role{company.Admin}},
	)

	// Add several multi-role users (some attached, some not)
	users = append(
		users,
		company.User{
			ID:       "multi_1",
			FullName: "Валерия Клименко",
			Roles:    []company.Role{company.Teacher, company.Dephead},
		},
		company.User{
			ID:           "multi_2",
			FullName:     "Родион Бирюков",
			Roles:        []company.Role{company.Teacher, company.DevelopmentDeputy},
			DepartmentID: departmentKafInf,
		},
		company.User{
			ID:       "multi_3",
			FullName: "Александра Громова",
			Roles:    []company.Role{company.ScientificDeputy, company.AcademicDirector},
		},
		company.User{
			ID:           "multi_4",
			FullName:     "Павел Трошин",
			Roles:        []company.Role{company.Teacher, company.Admin},
			DepartmentID: departmentKafMath,
		},
		company.User{
			ID:       "multi_5",
			FullName: "Евгений Захаров",
			Roles:    []company.Role{company.ChiefEconomist, company.DevelopmentDeputy},
		},
	)

	// Simple deterministic passwords for demo users
	pw := map[string]string{}
	for _, u := range users {
		// password equals id + "-pass" to keep it deterministic and simple
		pw[u.ID] = u.ID + "-pass"
	}
	// Ensure the canonical admin has the expected repo password
	pw["kozlovma"] = "yandexyandex"

	// Use NewLocal which internally creates an optimized storage
	return NewLocal(users, pw, deps)
}
