//nolint:cyclop // it should be high.
package main

import (
	"fmt"
	//nolint:depguard // this is a main file.
	"log"
	"os"
	"strings"
	"sync"
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

type LoginResponse struct {
	Token string `json:"token"`
}

type UserWithCreds struct {
	User        User
	Credentials Credentials
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

// duplicateChannel creates two channels that receive the same values from the input channel
func duplicateChannel[T any](in <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)
		for v := range in {
			out1 <- v
			out2 <- v
		}
	}()

	return out1, out2
}

// createDepartments creates departments and sends them through the output channel
func createDepartments(client *resty.Client, departments []Department, out chan<- Department) {
	defer close(out)
	for _, d := range departments {
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
		out <- d
	}
}

// createDepartmentHeads creates department heads for each department
func createDepartmentHeads(client *resty.Client, departments <-chan Department, out chan<- UserWithCreds) {
	defer close(out)
	for d := range departments {
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
			continue
		}

		out <- UserWithCreds{user, creds}
	}
}

// createTeachers creates teachers for each department
func createTeachers(client *resty.Client, departments <-chan Department, out chan<- UserWithCreds) {
	defer close(out)
	for d := range departments {
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
				continue
			}

			out <- UserWithCreds{user, creds}
		}
	}
}

// createDeputies creates deputies with different roles
func createDeputies(client *resty.Client, out chan<- UserWithCreds) {
	defer close(out)
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
			log.Printf("Failed to create deputy with role %d: %v (%s)", roleID, err, r.String())
			continue
		}
		user.ID = resp.ID

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
			continue
		}

		out <- UserWithCreds{user, creds}
	}
}

// createCommonFiles creates common files
func createCommonFiles(client *resty.Client, jobs <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for range jobs {
		content := gofakeit.LoremIpsumParagraph(5, 25, 400, "\n")
		name := gofakeit.MovieName() + ".txt"
		r, err := client.R().
			SetFileReader("file", name, strings.NewReader(content)).
			SetFormData(map[string]string{
				"name": name,
			}).
			Post("/files")
		if err != nil || !r.IsSuccess() {
			log.Printf("Failed to create common file %s: %v (%s)", name, err, r.String())
		}
	}
}

// createUserFiles creates files for each user
func createUserFiles(client *resty.Client, users <-chan UserWithCreds, wg *sync.WaitGroup) {
	defer wg.Done()
	const filesPerUser = 8
	for user := range users {
		for range filesPerUser {
			var resp LoginResponse
			r, err := client.R().SetBody(user.Credentials).SetResult(&resp).Post("/auth/login")
			if err != nil || !r.IsSuccess() {
				log.Printf("couldn't login as user %v: %v (%s)", user.User, err, r.String())
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
				log.Printf("Failed to create user file %s: %v (%s)", name, err, r.String())
			}
		}
	}
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

	const numWorkers = 8

	// Create channels
	departmentsChan := make(chan Department)
	departmentsForHeads, departmentsForTeachers := duplicateChannel(departmentsChan)
	departmentHeadsChan := make(chan UserWithCreds)
	teachersChan := make(chan UserWithCreds)
	deputiesChan := make(chan UserWithCreds)
	commonFilesJobs := make(chan struct{}, 150)

	// Start department creation
	go createDepartments(client, fakeDepartments, departmentsChan)

	// Start department heads creation
	go createDepartmentHeads(client, departmentsForHeads, departmentHeadsChan)

	// Start teachers creation
	go createTeachers(client, departmentsForTeachers, teachersChan)

	// Start deputies creation
	go createDeputies(client, deputiesChan)

	// Create common files with worker pool
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go createCommonFiles(client, commonFilesJobs, &wg)
	}

	// Send jobs for common files
	for range 150 {
		commonFilesJobs <- struct{}{}
	}
	close(commonFilesJobs)
	wg.Wait()

	// Create user files with worker pools
	wg = sync.WaitGroup{}

	// Department heads files
	for range numWorkers {
		wg.Add(1)
		go createUserFiles(client, departmentHeadsChan, &wg)
	}

	// Teachers files
	for range numWorkers {
		wg.Add(1)
		go createUserFiles(client, teachersChan, &wg)
	}

	// Deputies files
	for range numWorkers {
		wg.Add(1)
		go createUserFiles(client, deputiesChan, &wg)
	}

	wg.Wait()

	log.Println("Fake data generation completed")
}
