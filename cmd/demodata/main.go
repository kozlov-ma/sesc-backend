//nolint:gosec,sloglint,gocognit,nestif,funlen // this is a demonstration script
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/kozlov-ma/sesc-backend/apiclient/client"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/authentication"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/departments"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/roles"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/users"
	"github.com/kozlov-ma/sesc-backend/apiclient/models"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// Department data to be created if not exists
var departmentsData = []struct {
	Name        string
	Description string
}{
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

// Russian first names for random user generation
var firstNames = []string{
	"Александр", "Алексей", "Анатолий", "Андрей", "Антон", "Аркадий", "Артем", "Борис",
	"Вадим", "Валентин", "Валерий", "Василий", "Виктор", "Виталий", "Владимир", "Владислав",
	"Геннадий", "Георгий", "Глеб", "Григорий", "Даниил", "Денис", "Дмитрий", "Евгений",
	"Егор", "Иван", "Игорь", "Илья", "Кирилл", "Константин", "Лев", "Леонид",
	"Максим", "Марк", "Матвей", "Михаил", "Никита", "Николай", "Олег", "Павел",
	"Петр", "Роман", "Руслан", "Сергей", "Станислав", "Степан", "Тимофей", "Федор",
	"Юрий", "Ярослав",
}

// Russian last names for random user generation
var lastNames = []string{
	"Иванов", "Смирнов", "Кузнецов", "Попов", "Васильев", "Петров", "Соколов", "Михайлов",
	"Новиков", "Федоров", "Морозов", "Волков", "Алексеев", "Лебедев", "Семенов", "Егоров",
	"Павлов", "Козлов", "Степанов", "Николаев", "Орлов", "Андреев", "Макаров", "Никитин",
	"Захаров", "Зайцев", "Соловьев", "Борисов", "Яковлев", "Григорьев", "Романов", "Воробьев",
	"Сергеев", "Кузьмин", "Фролов", "Александров", "Дмитриев", "Королев", "Гусев", "Киселев",
	"Ильин", "Максимов", "Поляков", "Сорокин", "Виноградов", "Ковалев", "Белов", "Медведев",
	"Антонов", "Тарасов", "Жуков", "Баранов", "Филиппов", "Комаров", "Давыдов", "Беляев",
	"Герасимов", "Богданов", "Осипов", "Сидоров", "Матвеев", "Титов", "Марков", "Миронов",
}

// Russian patronymics for random user generation
var patronymics = []string{
	"Александрович", "Алексеевич", "Анатольевич", "Андреевич", "Антонович", "Аркадьевич", "Артемович", "Борисович",
	"Вадимович", "Валентинович", "Валерьевич", "Васильевич", "Викторович", "Витальевич", "Владимирович", "Владиславович",
	"Геннадьевич", "Георгиевич", "Глебович", "Григорьевич", "Даниилович", "Денисович", "Дмитриевич", "Евгеньевич",
	"Егорович", "Иванович", "Игоревич", "Ильич", "Кириллович", "Константинович", "Львович", "Леонидович",
	"Максимович", "Маркович", "Матвеевич", "Михайлович", "Никитич", "Николаевич", "Олегович", "Павлович",
	"Петрович", "Романович", "Русланович", "Сергеевич", "Станиславович", "Степанович", "Тимофеевич", "Федорович",
	"Юрьевич", "Ярославович",
}

// UserInfo stores information about created users
type UserInfo struct {
	ID         string
	FirstName  string
	LastName   string
	MiddleName string
	Role       string
	RoleID     int64
	Department string
	Username   string
	Password   string
}

// AchievementGroupData represents the data for an achievement group
type AchievementGroupData struct {
	Name        string
	Description string
}

// AchievementTemplateData represents the data for an achievement template
type AchievementTemplateData struct {
	Name         string
	Description  string
	PointsLimit  int64
	ReviewerRole int64
}

// Achievement groups data
var achievementGroupsData = []AchievementGroupData{
	{
		Name:        "Показатель № 1 Сопровождение (подготовка/организация, проведение) мероприятий программы развития, плана работы",
		Description: "Сопровождение мероприятий программы развития",
	},
	{
		Name:        "Показатель № 2 Обеспечение участия в мероприятиях и сопровождение обучающихся СУНЦ УрФУ",
		Description: "Обеспечение участия и сопровождение обучающихся в мероприятиях",
	},
	{Name: "Показатель № 3 Сопровождение дистанционного курса", Description: "Сопровождение дистанционного курса"},
	{
		Name:        "Показатель № 4.1 Научные, научно-методические публикации работников",
		Description: "Научные и научно-методические публикации",
	},
	{
		Name:        "Показатель № 4.2 Участие в конференции с докладом без последующей публикации в сборниках трудов",
		Description: "Участие в конференции с докладом",
	},
	{
		Name:        "Показатель № 4.3 Методические пособия, рабочие программы, курсы для обучающихся СУНЦ",
		Description: "Методические пособия и программы",
	},
	{
		Name:        "Показатель № 5 Региональные предметные комиссии по проверке развёрнутых ответов участников государственной итоговой аттестации",
		Description: "Участие в региональных предметных комиссиях",
	},
	{
		Name:        "Показатель № 6 Участие в организационной структуре Всероссийской олимпиады школьников",
		Description: "Участие в организации ВсОШ",
	},
	{Name: "Показатель № 7 Результаты ОГЭ", Description: "Результаты ОГЭ обучающихся"},
	{Name: "Показатель № 8 Результаты ЕГЭ", Description: "Результаты ЕГЭ обучающихся"},
	{Name: "Показатель № 9.1 Олимпиады", Description: "Результаты участия обучающихся в олимпиадах"},
	{
		Name:        "Показатель № 9.2 Научно-практические конференции обучающихся",
		Description: "Результаты участия в научно-практических конференциях",
	},
	{
		Name:        "Показатель № 9.3 Турниры, хакатоны, и иные интеллектуальные соревнования",
		Description: "Результаты участия в интеллектуальных соревнованиях",
	},
	{Name: "Показатель № 9.4 Проектная деятельность обучающихся", Description: "Результаты проектной деятельности"},
	{
		Name:        "Показатель № 10.1 Участие обучающихся СУНЦ УрФУ в спортивных соревнованиях",
		Description: "Результаты участия в спортивных соревнованиях",
	},
	{
		Name:        "Показатель № 10.2 Выполнение обручающимися комплекса ГТО с получением соответствующего значка",
		Description: "Выполнение комплекса ГТО",
	},
	{
		Name:        "Показатель № 11 Конкурсы профессионального мастерства",
		Description: "Участие в конкурсах профессионального мастерства",
	},
	{
		Name:        "Показатель № 12.1 Конкурс педагогического мастерства",
		Description: "Участие в конкурсе педагогического мастерства СУНЦ УрФУ",
	},
	{
		Name:        "Показатель № 12.2 Конкурс работников, реализующих программу воспитания и социализации",
		Description: "Участие в конкурсе работников СУНЦ УрФУ",
	},
	{
		Name:        "Показатель № 13 Участие работника в развитии учебно-методической и нормативно-правовой базы",
		Description: "Участие в развитии нормативной базы",
	},
	{
		Name:        "Показатель № 14 Победа работника в конкурсе на получение гранта",
		Description: "Получение гранта или привлечение пожертвования",
	},
	{
		Name:        "Показатель № 15 Участие работника в спортивных соревнованиях и сдача норм ГТО",
		Description: "Спортивные достижения работника",
	},
}

func main() {
	// Parse command line arguments
	adminUsername := flag.String("username", "", "Admin username")
	adminPassword := flag.String("password", "", "Admin password")
	host := flag.String("host", "localhost:8080", "API host")
	useHTTPS := flag.Bool("https", false, "Use HTTPS instead of HTTP")
	flag.Parse()

	if *adminUsername == "" || *adminPassword == "" {
		slog.Error("Admin username and password are required")
		os.Exit(1)
	}

	// Create API client
	schemes := []string{"http"}
	if *useHTTPS {
		schemes = []string{"https"}
	}
	transport := httptransport.New(*host, "/", schemes)
	apiClient := client.New(transport, strfmt.Default)

	// Login as admin
	authParams := authentication.NewPostAuthAdminLoginParams()
	authParams.SetRequest(&models.APICredentialsRequest{
		Username: adminUsername,
		Password: adminPassword,
	})

	authResp, err := apiClient.Authentication.PostAuthAdminLogin(authParams)
	if err != nil {
		slog.Error("Failed to authenticate", "error", err)
		os.Exit(1)
	}

	// Create auth info with bearer token
	authInfo := httptransport.BearerToken(*authResp.Payload.Token)

	// Process departments
	departmentMap, err := processDepartments(apiClient, authInfo)
	if err != nil {
		slog.Error("Failed to process departments", "error", err)
		os.Exit(1)
	}

	// Process achievement groups and templates
	groupMap, err := processAchievementGroups(apiClient, authInfo)
	if err != nil {
		slog.Error("Failed to process achievement groups", "error", err)
		os.Exit(1)
	}

	// Get roles
	rolesResp, err := apiClient.Roles.GetRoles(roles.NewGetRolesParams())
	if err != nil {
		slog.Error("Failed to get roles", "error", err)
		os.Exit(1)
	}

	roleMap := make(map[int64]*models.RespondRole)
	for _, role := range rolesResp.Payload.Roles {
		roleMap[*role.ID] = (*models.RespondRole)(role)
	}

	// Get all users
	usersResp, err := apiClient.Users.GetUsers(users.NewGetUsersParams(), authInfo)
	if err != nil {
		slog.Error("Failed to get users", "error", err)
		os.Exit(1)
	}

	// Process users
	userInfos, err := processUsers(apiClient, authInfo, departmentMap, roleMap, usersResp.Payload.Users)
	if err != nil {
		slog.Error("Failed to process users", "error", err)
		os.Exit(1)
	}

	// Print results
	slog.Info("=== Departments ===")
	for _, dept := range departmentMap {
		slog.Info("Department", "id", *dept.ID, "name", *dept.Name)
	}

	slog.Info("=== Achievement Groups ===")
	for _, group := range groupMap {
		slog.Info("Achievement Group", "id", *group.ID, "name", *group.Name)
	}

	slog.Info("=== Users ===")
	for _, user := range userInfos {
		slog.Info("User", "lastName", user.LastName, "firstName", user.FirstName, "middleName", user.MiddleName)
		slog.Info("  User ID", "id", user.ID)
		slog.Info("  User Role", "role", user.Role, "roleId", user.RoleID)
		slog.Info("  User Department", "department", user.Department)
		slog.Info("  User Credentials", "username", user.Username, "password", user.Password)
	}
}

// processDepartments ensures all required departments exist
func processDepartments(
	apiClient *client.Apiclient,
	authInfo runtime.ClientAuthInfoWriter,
) (map[string]*models.RespondDepartment, error) {
	// Get existing departments
	deptResp, err := apiClient.Departments.GetDepartments(departments.NewGetDepartmentsParams())
	if err != nil {
		return nil, fmt.Errorf("failed to get departments: %w", err)
	}

	// Create map of existing departments by name
	existingDepts := make(map[string]*models.RespondDepartment)
	for _, dept := range deptResp.Payload.Departments {
		existingDepts[*dept.Name] = dept
	}

	// Create map to return
	departmentMap := make(map[string]*models.RespondDepartment)

	// Create missing departments
	for _, deptData := range departmentsData {
		if dept, exists := existingDepts[deptData.Name]; exists {
			departmentMap[*dept.ID] = dept
			slog.Info("Department already exists", "name", *dept.Name, "id", *dept.ID)
		} else {
			// Create new department
			createParams := departments.NewPostDepartmentsParams()
			createParams.SetRequest(&models.APICreateDepartmentRequest{
				Name:        &deptData.Name,
				Description: &deptData.Description,
			})

			createResp, err := apiClient.Departments.PostDepartments(createParams, authInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to create department %s: %w", deptData.Name, err)
			}

			departmentMap[*createResp.Payload.ID] = createResp.Payload
			slog.Info("Created department", "name", *createResp.Payload.Name, "id", *createResp.Payload.ID)
		}
	}

	return departmentMap, nil
}

// processUsers ensures all required users exist
func processUsers(
	apiClient *client.Apiclient,
	authInfo runtime.ClientAuthInfoWriter,
	departmentMap map[string]*models.RespondDepartment,
	roleMap map[int64]*models.RespondRole,
	existingUsers []*models.RespondUser,
) ([]UserInfo, error) {
	// Map users by department and role
	deptUsers := make(map[string][]*models.RespondUser)
	roleUsers := make(map[int64][]*models.RespondUser)

	for _, user := range existingUsers {
		// Skip suspended users
		if *user.Suspended {
			continue
		}

		// Add to role map
		roleID := *user.Role.ID
		roleUsers[roleID] = append(roleUsers[roleID], user)

		// Add to department map if user has a department
		if user.DepartmentID != "" {
			deptID := user.DepartmentID
			deptUsers[deptID] = append(deptUsers[deptID], user)
		}
	}

	// Store all user infos
	var allUserInfos []UserInfo

	// Process department heads
	for deptID, dept := range departmentMap {
		// Check if department has a head
		hasDephead := false
		for _, user := range deptUsers[deptID] {
			if *user.Role.ID == int64(sesc.Dephead) {
				hasDephead = true

				// Get or create credentials for existing dephead
				userInfo, err := getOrCreateCredentials(apiClient, authInfo, user)
				if err != nil {
					return nil, fmt.Errorf("failed to get/create credentials for dephead %s: %w", *user.ID, err)
				}
				userInfo.Department = *dept.Name
				userInfo.Role = *user.Role.Name
				userInfo.RoleID = *user.Role.ID

				allUserInfos = append(allUserInfos, userInfo)
				break
			}
		}

		// Create department head if needed
		if !hasDephead {
			userInfo, err := createRandomUser(apiClient, authInfo, deptID, int64(sesc.Dephead))
			if err != nil {
				return nil, fmt.Errorf("failed to create dephead for department %s: %w", deptID, err)
			}
			userInfo.Department = *dept.Name
			userInfo.Role = *roleMap[int64(sesc.Dephead)].Name

			allUserInfos = append(allUserInfos, userInfo)

			// Add to department users
			getUserResp, err := apiClient.Users.GetUsersID(users.NewGetUsersIDParams().WithID(userInfo.ID), authInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to get created user: %w", err)
			}
			deptUsers[deptID] = append(deptUsers[deptID], getUserResp.Payload)
		}

		// Ensure department has at least 5 teachers
		teacherCount := 0
		for _, user := range deptUsers[deptID] {
			if *user.Role.ID == int64(sesc.Teacher) {
				teacherCount++

				// Get or create credentials for existing teacher
				userInfo, err := getOrCreateCredentials(apiClient, authInfo, user)
				if err != nil {
					return nil, fmt.Errorf("failed to get/create credentials for teacher %s: %w", *user.ID, err)
				}
				userInfo.Department = *dept.Name
				userInfo.Role = *user.Role.Name
				userInfo.RoleID = *user.Role.ID

				allUserInfos = append(allUserInfos, userInfo)
			}
		}

		// Create additional teachers if needed
		for teacherCount < 5 {
			userInfo, err := createRandomUser(apiClient, authInfo, deptID, int64(sesc.Teacher))
			if err != nil {
				return nil, fmt.Errorf("failed to create teacher for department %s: %w", deptID, err)
			}
			userInfo.Department = *dept.Name
			userInfo.Role = *roleMap[int64(sesc.Teacher)].Name

			allUserInfos = append(allUserInfos, userInfo)
			teacherCount++
		}
	}

	// Ensure all deputy roles are filled
	deputyRoles := []int64{int64(sesc.OlympiadDeputy), int64(sesc.ScientificDeputy), int64(sesc.DevelopmentDeputy)}
	for _, roleID := range deputyRoles {
		if len(roleUsers[roleID]) == 0 {
			// Create deputy without department
			userInfo, err := createRandomUser(apiClient, authInfo, "", roleID)
			if err != nil {
				return nil, fmt.Errorf("failed to create deputy with role %d: %w", roleID, err)
			}
			userInfo.Role = *roleMap[roleID].Name

			allUserInfos = append(allUserInfos, userInfo)
		} else {
			// Get or create credentials for existing deputy
			for _, user := range roleUsers[roleID] {
				userInfo, err := getOrCreateCredentials(apiClient, authInfo, user)
				if err != nil {
					return nil, fmt.Errorf("failed to get/create credentials for deputy %s: %w", *user.ID, err)
				}

				if user.DepartmentID != "" {
					userInfo.Department = user.DepartmentID
				} else {
					userInfo.Department = "Нет"
				}

				userInfo.Role = *user.Role.Name
				userInfo.RoleID = *user.Role.ID

				allUserInfos = append(allUserInfos, userInfo)
			}
		}
	}

	return allUserInfos, nil
}

// createRandomUser creates a random user with the given department and role
func createRandomUser(
	apiClient *client.Apiclient,
	authInfo runtime.ClientAuthInfoWriter,
	departmentID string,
	roleID int64,
) (UserInfo, error) {
	firstName := firstNames[rand.IntN(len(firstNames))]
	lastName := lastNames[rand.IntN(len(lastNames))]
	middleName := patronymics[rand.IntN(len(patronymics))]

	// Create empty picture URL
	pictureURL := ""

	// Create user
	createParams := users.NewPostUsersParams()
	createParams.SetRequest(&models.APICreateUserRequest{
		FirstName:  &firstName,
		LastName:   &lastName,
		MiddleName: middleName,
		Role:       &roleID,
		PictureURL: pictureURL,
	})

	// Add department if specified
	if departmentID != "" {
		// Need to get the current request and update it
		currentRequest := createParams.Request
		currentRequest.DepartmentID = departmentID
		createParams.SetRequest(currentRequest)
	}

	createResp, err := apiClient.Users.PostUsers(createParams, authInfo)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate username and password
	username := generateUsername(firstName, lastName)
	password := generatePassword()

	// Set credentials
	credParams := authentication.NewPutUsersIDCredentialsParams()
	credParams.SetID(*createResp.Payload.ID)
	credParams.SetRequest(&models.APICredentialsRequest{
		Username: &username,
		Password: &password,
	})

	_, err = apiClient.Authentication.PutUsersIDCredentials(credParams, authInfo)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to set credentials: %w", err)
	}

	return UserInfo{
		ID:         *createResp.Payload.ID,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		RoleID:     roleID,
		Username:   username,
		Password:   password,
	}, nil
}

// getOrCreateCredentials gets existing credentials or creates new ones
func getOrCreateCredentials(
	apiClient *client.Apiclient,
	authInfo runtime.ClientAuthInfoWriter,
	user *models.RespondUser,
) (UserInfo, error) {
	userInfo := UserInfo{
		ID:         *user.ID,
		FirstName:  *user.FirstName,
		LastName:   *user.LastName,
		MiddleName: user.MiddleName,
	}

	// Try to get credentials
	credParams := authentication.NewGetAuthCredentialsIDParams()
	credParams.SetID(*user.ID)
	credResp, err := apiClient.Authentication.GetAuthCredentialsID(credParams, authInfo)

	if err == nil && credResp.Payload.Username != nil {
		// Credentials exist
		userInfo.Username = *credResp.Payload.Username
		userInfo.Password = *credResp.Payload.Password
	} else {
		// Create new credentials
		username := generateUsername(*user.FirstName, *user.LastName)
		password := generatePassword()

		credParams := authentication.NewPutUsersIDCredentialsParams()
		credParams.SetID(*user.ID)
		credParams.SetRequest(&models.APICredentialsRequest{
			Username: &username,
			Password: &password,
		})

		_, err = apiClient.Authentication.PutUsersIDCredentials(credParams, authInfo)
		if err != nil {
			return UserInfo{}, fmt.Errorf("failed to set credentials: %w", err)
		}

		userInfo.Username = username
		userInfo.Password = password
	}

	return userInfo, nil
}

// generateUsername generates a username from first and last name
func generateUsername(firstName, lastName string) string {
	// Transliterate to Latin
	firstLatin := transliterate(firstName)
	lastLatin := transliterate(lastName)

	// Take first letter of first name and full last name
	username := strings.ToLower(firstLatin[:1] + lastLatin)

	// Add random number to make it unique
	username += fmt.Sprintf("%03d", rand.IntN(1000))

	return username
}

// generatePassword generates a random password
func generatePassword() string {
	// Simple password for demo purposes
	return fmt.Sprintf("Pass%03d", rand.IntN(1000))
}

// transliterate converts Russian text to Latin
func transliterate(text string) string {
	transMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "",
		'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "E",
		'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M",
		'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sch", 'Ъ': "",
		'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	var result strings.Builder
	for _, r := range text {
		if s, ok := transMap[r]; ok {
			result.WriteString(s)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
