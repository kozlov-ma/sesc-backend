package api

import (
	"net/http"
	"slices"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// FakeData godoc
// @Summary Create a lot of fake data (for testing and development purposes)
// @Description Creates departments, users, credentials, ...
// @Tags dev
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200
// @Failure 500 {object} ServerError "Internal server error"
// @Router /dev/fakedata [post]
func (a *API) FakeData(_ http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var fakeDepartments = []sesc.Department{
		{
			Name:        "Кафедра филологии",
			Description: "Кафедра была организована в 1999 году. Первым заведующим кафедрой был В. С. Рабинович.",
		},
		{
			Name:        "Кафедра гуманитарного образования",
			Description: "Кафедра была организована в 1989 году. Первым заведующим кафедрой был В. И. Михайленко.",
		},
		{
			Name:        "Кафедра иностранных языков",
			Description: "Кафедра была организована в 1990 году. Первым заведующим кафедрой была Н. А. Столярова.",
		},
		{
			Name:        "Кафедра математики",
			Description: "Кафедра была организована в 1995 году. Первым заведующим кафедрой был В. В. Расин.",
		},
		{
			Name:        "Кафедра информатики",
			Description: "Кафедра была организована в 1995 году. Первым заведующим кафедрой был Д. Я. Шараев.",
		},
		{
			Name:        "Кафедра физики и астрономии",
			Description: "Кафедра была организована в 1993 году. Первым заведующим кафедрой был З. И. Урицкий.",
		},
		{
			Name:        "Кафедра химии и биологии",
			Description: "Кафедра была организована в 1993 году. Первым заведующим кафедрой был А. В. Гурьев.",
		},
		{
			Name:        "Кафедра психофизической культуры",
			Description: "Кафедра была организована в 1990 году. Первым заведующим кафедрой был В. Р. Малкин.",
		},
	}

	depts := make([]sesc.Department, 0, len(fakeDepartments))
	for _, d := range fakeDepartments {
		de, err := a.sesc.CreateDepartment(ctx, d.Name, d.Description)
		if err != nil {
			a.log.WarnContext(ctx, "couldn't create department", "name", d.Name, "error", err)
		}

		depts = append(depts, de)
	}

	var depheads []sesc.UserUpdateOptions
	for _, d := range depts {
		depheads = append(depheads, sesc.UserUpdateOptions{
			FirstName:    gofakeit.FirstName(),
			LastName:     gofakeit.LastName(),
			MiddleName:   gofakeit.MiddleName(),
			DepartmentID: d.ID,
			NewRoleID:    sesc.Dephead.ID,
		})
	}

	var teachers []sesc.UserUpdateOptions
	for _, d := range depts {
		const minTeachersPerDept = 7
		const maxTeachersPerDept = 27
		for range gofakeit.Number(minTeachersPerDept, maxTeachersPerDept) {
			teachers = append(teachers, sesc.UserUpdateOptions{
				FirstName:    gofakeit.FirstName(),
				LastName:     gofakeit.LastName(),
				MiddleName:   gofakeit.MiddleName(),
				DepartmentID: d.ID,
				NewRoleID:    sesc.Teacher.ID,
			})
		}
	}

	var deputies = []sesc.UserUpdateOptions{
		{
			FirstName:  gofakeit.FirstName(),
			LastName:   gofakeit.LastName(),
			MiddleName: gofakeit.MiddleName(),
			NewRoleID:  sesc.ContestDeputy.ID,
		},
		{
			FirstName:  gofakeit.FirstName(),
			LastName:   gofakeit.LastName(),
			MiddleName: gofakeit.MiddleName(),
			NewRoleID:  sesc.ScientificDeputy.ID,
		},
		{
			FirstName:  gofakeit.FirstName(),
			LastName:   gofakeit.LastName(),
			MiddleName: gofakeit.MiddleName(),
			NewRoleID:  sesc.DevelopmentDeputy.ID,
		},
	}

	allUsers := slices.Concat(teachers, depheads, deputies)

	for _, u := range allUsers {
		us, err := a.sesc.CreateUser(ctx, u)
		if err != nil {
			a.log.ErrorContext(ctx, "couldn't create user", "error", err)
			continue
		}

		_, err = a.iam.RegisterCredentials(ctx, us.ID, iam.Credentials{
			Username: gofakeit.Username(),
			Password: "password",
		})
		if err != nil {
			a.log.ErrorContext(ctx, "couldn't register user credentials", "error", err)
			continue
		}
	}

	a.log.InfoContext(ctx, "created fake data")
}
