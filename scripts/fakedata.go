//nolint:cyclop // it should be high.
package main

import (
	"fmt"
	//nolint:depguard // this is a main file.
	"log"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"resty.dev/v3"
)

type Department struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type User struct {
	ID           string `json:"id,omitzero"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	MiddleName   string `json:"middleName,omitzero"`
	DepartmentID string `json:"departmentId,omitzero"`
	RoleID       int32  `json:"roleId"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type File struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Common      bool   `json:"common"`
}

var fakeDepartments = []Department{
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

type LoginResponse struct {
	Token string `json:"token"`
}

//nolint:funlen,gocognit // it should be long.
func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: fakedata <base_url> <admin_token>")
	}

	baseURL := os.Args[1]
	adminToken := os.Args[2]

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", "Bearer "+adminToken).
		SetTimeout(3 * time.Second)

	// Create departments
	departments := make([]Department, 0, len(fakeDepartments))
	for _, d := range fakeDepartments {
		var resp struct {
			ID string `json:"id"`
		}
		r, err := client.R().
			SetBody(d).
			SetResult(&resp).
			Post("/departments")
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create department %s: %v (%s)", d.Name, err, r.String())
			continue
		}
		d.ID = resp.ID
		departments = append(departments, d)
	}

	// Create users
	users := make([]User, 0)
	userCredentials := make(map[string]Credentials)

	// Create department heads
	for _, d := range departments {
		user := User{
			FirstName:    gofakeit.FirstName(),
			LastName:     gofakeit.LastName(),
			MiddleName:   gofakeit.MiddleName(),
			DepartmentID: d.ID,
			RoleID:       2, // Dephead role
		}
		var resp struct {
			ID string `json:"id"`
		}
		r, err := client.R().
			SetBody(user).
			SetResult(&resp).
			Post("/users")
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create department head for %s: %v (%s)", d.Name, err, r.String())
			continue
		}
		user.ID = resp.ID
		users = append(users, user)
		creds := Credentials{
			Username: gofakeit.Username(),
			Password: "password",
		}

		// Create credentials
		r, err = client.R().
			SetBody(creds).
			Put(fmt.Sprintf("/users/%s/credentials", user.ID))
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create credentials for user %s: %v (%s)", user.ID, err, r.String())
		}

		userCredentials[user.ID] = creds
	}

	// Create teachers
	for _, d := range departments {
		numTeachers := gofakeit.Number(7, 27)
		for range numTeachers {
			user := User{
				FirstName:    gofakeit.FirstName(),
				LastName:     gofakeit.LastName(),
				MiddleName:   gofakeit.MiddleName(),
				DepartmentID: d.ID,
				RoleID:       1, // Teacher role
			}
			var resp struct {
				ID string `json:"id"`
			}
			r, err := client.R().
				SetBody(user).
				SetResult(&resp).
				Post("/users")
			if err != nil || !r.IsSuccess() {
				log.Printf("Failed to create teacher for %s: %v (%s)", d.Name, err, r.String())
				continue
			}
			user.ID = resp.ID
			users = append(users, user)

			creds := Credentials{
				Username: gofakeit.Username(),
				Password: "password",
			}
			// Create credentials
			r, err = client.R().
				SetBody(creds).
				Put(fmt.Sprintf("/users/%s/credentials", user.ID))
			if err != nil || !r.IsSuccess() {
				log.Printf("Failed to create credentials for user %s: %v (%s)", user.ID, err, r)
			}

			userCredentials[user.ID] = creds
		}
	}

	// Create deputies
	deputyRoles := []int32{3, 4, 5} // ContestDeputy, ScientificDeputy, DevelopmentDeputy
	for _, roleID := range deputyRoles {
		user := User{
			FirstName:  gofakeit.FirstName(),
			LastName:   gofakeit.LastName(),
			MiddleName: gofakeit.MiddleName(),
			RoleID:     roleID,
		}
		var resp struct {
			ID string `json:"id"`
		}
		r, err := client.R().
			SetBody(user).
			SetResult(&resp).
			Post("/users")
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create deputy with role %d: %v (%s)", roleID, err, r)
			continue
		}
		user.ID = resp.ID
		users = append(users, user)

		creds := Credentials{
			Username: gofakeit.Username(),
			Password: "password",
		}
		// Create credentials
		r, err = client.R().
			SetBody(creds).
			Put(fmt.Sprintf("/users/%s/credentials", user.ID))
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create credentials for user %s: %v (%s)", user.ID, err, r)
		}

		userCredentials[user.ID] = creds
	}

	const commonFiles = 150
	// Create common files
	for range commonFiles {
		// Create a simple text file
		content := gofakeit.LoremIpsumParagraph(5, 25, 400, "\n")
		name := gofakeit.MovieName() + ".txt"
		r, err := client.R().
			SetFileReader("file", name, strings.NewReader(content)).
			SetFormData(map[string]string{
				"name": name,
			}).
			Post("/files")
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create common file %s: %v (%s)", name, err, r)
		}
	}

	const filesPerUser = 8
	// Create user files
	for _, user := range users {
		for range filesPerUser {
			var resp LoginResponse
			r, err := client.R().SetBody(userCredentials[user.ID]).SetResult(&resp).Post("/auth/login")
			if err != nil || !r.IsSuccess() {
				log.Printf("couldn't login as user %v: %v (%s)", user, err, r)
				continue
			}

			content := gofakeit.LoremIpsumParagraph(5, 25, 400, "\n")
			name := gofakeit.MovieName() + ".txt"
			r, err = client.R().
				SetHeader("Authorization", fmt.Sprintf("Bearer %s", resp.Token)).
				SetFileReader("file", name, strings.NewReader(content)).
				SetFormData(map[string]string{
					"name": name,
				}).
				Post("/files")
			if err != nil || !r.IsSuccess() {
				log.Printf("Failed to create common file %s: %v (%s)", name, err, r)
			}
		}
	}

	log.Println("Fake data generation completed")
}
