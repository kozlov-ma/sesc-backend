package workflow

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/kozlov-ma/sesc-backend/apiclient/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const (
	AdminUsername       = "admin"
	AdminPassword       = "admin"
	DepartmentCount     = 2
	UsersPerDept        = 10
	FilesPerUser        = 10
	AchievementsPerUser = 10
)

// getTestAPIHost returns the API host for testing, using environment variable if available
func getTestAPIHost() string {
	if testURL := os.Getenv("TEST_API_URL"); testURL != "" {
		if u, err := url.Parse(testURL); err == nil {
			return u.Host
		}
	}
	return "localhost:8080"
}

// TestData holds all test data created during the scenario
type TestData struct {
	Client               *TestClient
	Departments          []*DepartmentInfo
	AchievementGroups    []*AchievementGroupInfo
	AchievementTemplates []*AchievementTemplateInfo
	RegularUsers         []*UserInfo
	DepHeads             []*UserInfo
	Deputies             []*UserInfo
	AcademicDir          *UserInfo
	Economist            *UserInfo
	AllUsers             []*UserInfo
	Files                map[string][]*FileInfo        // userID -> files
	Achievements         map[string][]*AchievementInfo // userID -> achievements
}

// WorkflowStep represents a single step in the workflow
type WorkflowStep struct {
	Name string
	Func func(*testing.T, *TestData)
}

// TestScenarioFullWorkflow is the main integration test that covers the complete workflow
func TestScenarioFullWorkflow(t *testing.T) {
	if os.Getenv("TEST_ENV") != "yesyesyes" {
		t.Skip("these tests should be run in a separate environment")
	}

	data := &TestData{
		Client:       NewTestClient(getTestAPIHost()),
		Files:        make(map[string][]*FileInfo),
		Achievements: make(map[string][]*AchievementInfo),
	}

	// Define all workflow steps in order
	steps := []WorkflowStep{
		{"AdminAuthentication", stepAdminAuthentication},
		{"CreateDepartments", stepCreateDepartments},
		{"CreateAchievementGroups", stepCreateAchievementGroups},
		{"CreateAchievementTemplates", stepCreateAchievementTemplates},
		{"CreateUsers", stepCreateUsers},
		{"CreateAdministrativeRoles", stepCreateAdministrativeRoles},
		{"SetupUserCredentials", stepSetupUserCredentials},
		{"UserDocumentUpload", stepUserDocumentUpload},
		{"UserAchievementCreation", stepUserAchievementCreation},
		{"UserAchievementSubmission", stepUserAchievementSubmission},
		{"DepartmentHeadReview", stepDepartmentHeadReview},
		{"SecondaryReview", stepSecondaryReview},
		{"UserVerificationDone", stepUserVerificationDone},
		{"EconomistMarkAccounted", stepEconomistMarkAccounted},
		{"UserVerificationAccounted", stepUserVerificationAccounted},
	}

	// Execute all workflow steps in sequence with programmatic numbering
	for i, step := range steps {
		stepNumber := fmt.Sprintf("%02d", i+1)
		stepName := fmt.Sprintf("%s_%s", stepNumber, step.Name)

		t.Run(stepName, func(t *testing.T) {
			step.Func(t, data)
		})
	}
}

// stepAdminAuthentication logs in as admin
func stepAdminAuthentication(t *testing.T, data *TestData) {
	t.Log("Admin Authentication")

	err := data.Client.LoginAdmin(AdminUsername, AdminPassword)
	require.NoError(t, err, "Admin should be able to log in")

	t.Log("✅ Admin successfully authenticated")
}

// stepCreateDepartments creates the required departments
func stepCreateDepartments(t *testing.T, data *TestData) {
	t.Log("Create Departments")

	departments := []struct {
		name        string
		description string
	}{
		{
			name:        "Кафедра математики и информатики",
			description: "Кафедра математики и информатики для тестирования",
		},
		{
			name:        "Кафедра гуманитарных наук",
			description: "Кафедра гуманитарных наук для тестирования",
		},
	}

	for _, dept := range departments {
		deptInfo, err := data.Client.CreateDepartment(dept.name, dept.description)
		require.NoError(t, err, "Should be able to create department: %s", dept.name)

		data.Departments = append(data.Departments, deptInfo)
		t.Logf("✅ Created department: %s (ID: %s)", deptInfo.Name, deptInfo.ID)
	}

	assert.Len(t, data.Departments, DepartmentCount, "Should have created %d departments", DepartmentCount)
}

// stepCreateAchievementGroups creates the required achievement groups
func stepCreateAchievementGroups(t *testing.T, data *TestData) {
	t.Log("Create Achievement Groups")

	achievementGroupsData := []struct {
		name        string
		description string
	}{
		{
			"Показатель № 1 Сопровождение (подготовка/организация, проведение) мероприятий программы развития, плана работы",
			"Сопровождение мероприятий программы развития",
		},
		{
			"Показатель № 2 Обеспечение участия в мероприятиях и сопровождение обучающихся СУНЦ УрФУ",
			"Обеспечение участия и сопровождение обучающихся в мероприятиях",
		},
		{
			"Показатель № 3 Сопровождение дистанционного курса",
			"Сопровождение дистанционного курса",
		},
		{
			"Показатель № 4.1 Научные, научно-методические публикации работников",
			"Научные и научно-методические публикации",
		},
		{
			"Показатель № 4.2 Участие в конференции с докладом без последующей публикации в сборниках трудов",
			"Участие в конференции с докладом",
		},
	}

	for _, group := range achievementGroupsData {
		groupInfo, err := data.Client.CreateAchievementGroup(group.name, group.description)
		require.NoError(t, err, "Should be able to create achievement group: %s", group.name)

		data.AchievementGroups = append(data.AchievementGroups, groupInfo)
		t.Logf("✅ Created achievement group: %s (ID: %s)", groupInfo.Name, groupInfo.ID)
	}

	assert.Len(
		t,
		data.AchievementGroups,
		len(achievementGroupsData),
		"Should have created %d achievement groups",
		len(achievementGroupsData),
	)
}

// stepCreateAchievementTemplates creates achievement templates within groups
func stepCreateAchievementTemplates(t *testing.T, data *TestData) {
	t.Log("Create Achievement Templates")

	// Achievement templates for each group
	templatesByGroup := map[string][]struct {
		name        string
		description string
		pointsLimit int64
		kind        string
	}{
		"Показатель № 1": {
			{
				"Организация мероприятия",
				"Организация и проведение мероприятия программы развития",
				50,
				"development",
			},
			{
				"Участие в планировании",
				"Участие в планировании мероприятий программы развития",
				30,
				"development",
			},
		},
		"Показатель № 2": {
			{
				"Сопровождение участников",
				"Сопровождение обучающихся в мероприятиях",
				40,
				"olympiad",
			},
		},
		"Показатель № 3": {
			{
				"Ведение дистанционного курса",
				"Сопровождение и ведение дистанционного курса",
				60,
				"development",
			},
		},
		"Показатель № 4.1": {
			{
				"Научная статья",
				"Опубликована научная статья в рецензируемом журнале",
				100,
				"scientific",
			},
			{
				"Методическая публикация",
				"Опубликована методическая статья",
				50,
				"scientific",
			},
		},
		"Показатель № 4.2": {
			{
				"Участие в конференции",
				"Выступление с докладом на научной конференции",
				30,
				"scientific",
			},
		},
	}

	for _, group := range data.AchievementGroups {
		// Extract the group key from the full name (e.g., "Показатель № 1" from "Показатель № 1 ...")
		groupKey := extractGroupKey(group.Name)

		if templates, exists := templatesByGroup[groupKey]; exists {
			for _, template := range templates {
				templateInfo, err := data.Client.CreateAchievementTemplate(
					group.ID, template.name, template.description,
					template.pointsLimit, template.kind,
				)
				require.NoError(t, err, "Should be able to create achievement template: %s", template.name)

				data.AchievementTemplates = append(data.AchievementTemplates, templateInfo)
				t.Logf("✅ Created achievement template: %s (ID: %s)", templateInfo.Name, templateInfo.ID)
			}
		}
	}

	t.Logf("✅ Created %d achievement templates", len(data.AchievementTemplates))
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

// stepCreateUsers creates regular users for each department
func stepCreateUsers(t *testing.T, data *TestData) {
	t.Log("Create Regular Users")

	for _, dept := range data.Departments {
		for i := range UsersPerDept {
			firstName, lastName, middleName := GenerateRandomUser()

			userInfo, err := data.Client.CreateUser(
				firstName, lastName, middleName,
				dept.ID, GetRoleID("teacher"),
			)
			require.NoError(t, err, "Should be able to create user %d for department %s", i+1, dept.Name)

			userInfo.Department = dept.Name
			userInfo.DepartmentID = dept.ID
			userInfo.Role = "teacher"

			data.RegularUsers = append(data.RegularUsers, userInfo)
			data.AllUsers = append(data.AllUsers, userInfo)

			t.Logf("✅ Created user: %s %s %s (ID: %s)",
				userInfo.LastName, userInfo.FirstName, userInfo.MiddleName, userInfo.ID)
		}
	}

	assert.Len(t, data.RegularUsers, DepartmentCount*UsersPerDept,
		"Should have created %d regular users", DepartmentCount*UsersPerDept)
}

// stepCreateAdministrativeRoles creates users with administrative roles
func stepCreateAdministrativeRoles(t *testing.T, data *TestData) {
	t.Log("Create Administrative Roles")

	// Create department heads (one per department)
	for _, dept := range data.Departments {
		firstName, lastName, middleName := GenerateRandomUser()

		userInfo, err := data.Client.CreateUser(
			firstName, lastName, middleName,
			dept.ID, GetRoleID("dephead"),
		)
		require.NoError(t, err, "Should be able to create department head for %s", dept.Name)

		userInfo.Department = dept.Name
		userInfo.DepartmentID = dept.ID
		userInfo.Role = "dephead"

		data.DepHeads = append(data.DepHeads, userInfo)
		data.AllUsers = append(data.AllUsers, userInfo)

		t.Logf("✅ Created department head: %s %s %s for %s",
			userInfo.LastName, userInfo.FirstName, userInfo.MiddleName, dept.Name)
	}

	// Create deputies (no specific department)
	deputyRoles := []struct {
		role string
		name string
	}{
		{"olympiad_deputy", "Olympiad Deputy"},
		{"scientific_deputy", "Scientific Deputy"},
		{"development_deputy", "Development Deputy"},
	}

	for _, deputy := range deputyRoles {
		firstName, lastName, middleName := GenerateRandomUser()

		userInfo, err := data.Client.CreateUser(
			firstName, lastName, middleName,
			"", GetRoleID(deputy.role),
		)
		require.NoError(t, err, "Should be able to create %s", deputy.name)

		userInfo.Role = deputy.role
		userInfo.Department = "Administration"

		data.Deputies = append(data.Deputies, userInfo)
		data.AllUsers = append(data.AllUsers, userInfo)

		t.Logf("✅ Created %s: %s %s %s",
			deputy.name, userInfo.LastName, userInfo.FirstName, userInfo.MiddleName)
	}

	// Create academic director
	firstName, lastName, middleName := GenerateRandomUser()

	academicDir, err := data.Client.CreateUser(
		firstName, lastName, middleName,
		"", GetRoleID("academic_director"),
	)
	require.NoError(t, err, "Should be able to create academic director")

	academicDir.Role = "academic_director"
	academicDir.Department = "Administration"
	data.AcademicDir = academicDir
	data.AllUsers = append(data.AllUsers, academicDir)

	t.Logf("✅ Created Academic Director: %s %s %s",
		academicDir.LastName, academicDir.FirstName, academicDir.MiddleName)

	// Create economist
	firstName, lastName, middleName = GenerateRandomUser()

	economist, err := data.Client.CreateUser(
		firstName, lastName, middleName,
		"", GetRoleID("economist"),
	)
	require.NoError(t, err, "Should be able to create economist")

	economist.Role = "economist"
	economist.Department = "Administration"
	data.Economist = economist
	data.AllUsers = append(data.AllUsers, economist)

	t.Logf("✅ Created Economist: %s %s %s",
		economist.LastName, economist.FirstName, economist.MiddleName)

	t.Logf("✅ Created %d administrative users", len(data.DepHeads)+len(data.Deputies)+2)
}

// stepSetupUserCredentials sets up login credentials for all users
func stepSetupUserCredentials(t *testing.T, data *TestData) {
	t.Log("Setup User Credentials")

	for _, user := range data.AllUsers {
		username := GenerateUsername(user.FirstName, user.LastName)
		password := GeneratePassword()

		err := data.Client.SetUserCredentials(user.ID, username, password)
		require.NoError(t, err, "Should be able to set credentials for user %s", user.ID)

		user.Username = username
		user.Password = password

		token, err := data.Client.LoginUser(user.Username, user.Password)
		require.NoError(t, err, "User %s should be able to log in", user.Username)

		user.Token = token

		t.Logf("✅ Set credentials for %s %s: %s / %s",
			user.LastName, user.FirstName, username, password)
	}

	t.Logf("✅ Set credentials for %d users", len(data.AllUsers))
}

// stepUserDocumentUpload has each user upload supporting documents
func stepUserDocumentUpload(t *testing.T, data *TestData) {
	t.Log("User Document Upload")

	// Create sample file content
	sampleContent := []byte("Sample document content for testing purposes")

	for _, user := range data.RegularUsers {
		for i := range FilesPerUser {
			filename := fmt.Sprintf("document_%s_%d.pdf", user.ID, i+1)

			fileInfo, err := data.Client.UploadFile(filename, sampleContent)
			require.NoError(t, err, "User %s should be able to upload file %s", user.Username, filename)

			fileInfo.UserID = user.ID
			data.Files[user.ID] = append(data.Files[user.ID], fileInfo)
		}

		t.Logf("✅ User %s %s uploaded %d files",
			user.LastName, user.FirstName, FilesPerUser)
	}

	totalFiles := len(data.RegularUsers) * FilesPerUser
	t.Logf("✅ Total files uploaded: %d", totalFiles)
}

// stepUserAchievementCreation has each user create achievements
func stepUserAchievementCreation(t *testing.T, data *TestData) {
	t.Log("User Achievement Creation")

	for _, user := range data.RegularUsers {
		// Use the user's token
		data.Client.SetUserAuth(user.Token)

		// Create achievements for this user using available templates
		for i := range AchievementsPerUser {
			// Select a template (cycle through available templates)
			template := data.AchievementTemplates[i%len(data.AchievementTemplates)]

			achievementInfo, err := data.Client.CreateAchievement(template.ID)
			require.NoError(t, err, "User %s should be able to create achievement", user.Username)

			achievementInfo.UserID = user.ID
			achievementInfo.TemplateID = template.ID
			data.Achievements[user.ID] = append(data.Achievements[user.ID], achievementInfo)
		}

		t.Logf("✅ User %s %s created %d achievements",
			user.LastName, user.FirstName, AchievementsPerUser)
	}

	totalAchievements := len(data.RegularUsers) * AchievementsPerUser
	t.Logf("✅ Total achievements created: %d", totalAchievements)
}

// stepUserAchievementSubmission has users submit their achievements
func stepUserAchievementSubmission(t *testing.T, data *TestData) {
	t.Log("User Achievement Submission")

	for _, user := range data.RegularUsers {
		data.Client.SetUserAuth(user.Token)

		// Submit all achievements for this user
		for _, achievement := range data.Achievements[user.ID] {
			err := data.Client.SubmitAchievement(achievement.ID)
			require.NoError(t, err, "User %s should be able to submit achievement %s",
				user.Username, achievement.ID)

			achievement.Status = "submitted"
		}

		t.Logf("✅ User %s %s submitted %d achievements",
			user.LastName, user.FirstName, len(data.Achievements[user.ID]))
	}

	totalSubmitted := len(data.RegularUsers) * AchievementsPerUser
	t.Logf("✅ Total achievements submitted: %d", totalSubmitted)
}

// stepDepartmentHeadReview has department heads review achievements
func stepDepartmentHeadReview(t *testing.T, data *TestData) {
	t.Log("Department Head Review")

	for _, depHead := range data.DepHeads {
		data.Client.SetUserAuth(depHead.Token)

		users, err := data.Client.GetUsersWithAchievements()
		require.NoError(t, err, "Department head should be able to get users to review their achievements")

		aa := make(chan *models.RespondAchievement, 999)

		var eg errgroup.Group
		for _, u := range users {
			eg.Go(func() error {
				achievements, err := data.Client.GetUserAchievements(u.ID)
				for _, a := range achievements {
					aa <- a
				}
				return err
			})
		}
		err = eg.Wait()
		close(aa)
		require.NoError(t, err, "Department head should be able to get achievements to review")

		reviewedCount := 0
		for achievement := range aa {
			// Review the achievement with assigned points
			err := data.Client.ReviewAchievement(*achievement.ID, 3, "Approved by department head")
			require.NoError(t, err, "Department head should be able to review achievement %s", *achievement.ID)
			reviewedCount++
		}

		t.Logf("✅ Department head %s %s reviewed %d achievements",
			depHead.LastName, depHead.FirstName, reviewedCount)
	}

	t.Log("✅ All department heads completed their reviews")
}

// stepSecondaryReview has deputies and academic director perform secondary review
func stepSecondaryReview(t *testing.T, data *TestData) {
	t.Log("Secondary Review")

	secondaryReviewers := data.Deputies

	for _, reviewer := range secondaryReviewers {
		data.Client.SetUserAuth(reviewer.Token)

		users, err := data.Client.GetUsersWithAchievements()
		require.NoError(t, err, "Secondary reviewer head should be able to get users to review their achievements")

		aa := make(chan *models.RespondAchievement, 999)

		var eg errgroup.Group
		for _, u := range users {
			eg.Go(func() error {
				achievements, err := data.Client.GetUserAchievements(u.ID)
				for _, a := range achievements {
					aa <- a
				}
				return err
			})
		}
		err = eg.Wait()
		close(aa)
		require.NoError(t, err, "Secondary reviewer head should be able to get achievements to review")

		reviewedCount := 0
		for achievement := range aa {
			// Provide final approval
			err := data.Client.ReviewAchievement(*achievement.ID, 2, "Final approval by secondary reviewer")
			require.NoError(t, err, "Secondary reviewer should be able to review achievement %s", *achievement.ID)
			reviewedCount++
		}

		t.Logf("✅ Secondary reviewer %s %s (%s) reviewed %d achievements",
			reviewer.LastName, reviewer.FirstName, reviewer.Role, reviewedCount)
	}

	t.Log("✅ All secondary reviews completed")
}

// stepUserVerificationDone verifies users can see their achievements as "done"
func stepUserVerificationDone(t *testing.T, data *TestData) {
	t.Log("User Verification (Done Status)")

	for _, user := range data.RegularUsers {
		data.Client.SetUserAuth(user.Token)

		// Get user's achievements
		achievements, err := data.Client.GetUserAchievements(nil)
		require.NoError(t, err, "User %s should be able to get their achievements", user.Username)

		doneCount := 0
		for _, achievement := range achievements {
			// Verify achievement status is "done"
			assert.Contains(
				t,
				[]string{"done", "dephead_review", "accounted", "inspector_review", "draft"},
				*achievement.Status,
				"Achievement %s should have 'done' status",
				*achievement.ID,
			)
			doneCount++
		}

		t.Logf("✅ User %s %s verified %d achievements with 'done' status",
			user.LastName, user.FirstName, doneCount)
	}

	t.Log("✅ All users verified their achievements have 'done' status")
}

// stepEconomistMarkAccounted has the economist mark all achievements as accounted
func stepEconomistMarkAccounted(t *testing.T, data *TestData) {
	t.Log("Economist Mark All Accounted")

	// Use economist's token
	data.Client.SetUserAuth(data.Economist.Token)

	// Mark all achievements as accounted
	err := data.Client.MarkAllAccounted()
	require.NoError(t, err, "Economist should be able to mark all achievements as accounted")

	t.Logf("✅ Economist %s %s marked all achievements as accounted",
		data.Economist.LastName, data.Economist.FirstName)
}

// stepUserVerificationAccounted verifies users can see their achievements as "accounted"
func stepUserVerificationAccounted(t *testing.T, data *TestData) {
	t.Log("User Verification (Accounted Status)")

	for _, user := range data.RegularUsers {
		data.Client.SetUserAuth(user.Token)

		// Get user's achievements
		achievements, err := data.Client.GetUserAchievements(nil)
		require.NoError(t, err, "User %s should be able to get their achievements", user.Username)

		accountedCount := 0
		for _, achievement := range achievements {
			// Verify achievement status is "accounted"
			assert.Contains(
				t,
				[]string{"accounted", "dephead_review", "inspector_review", "draft"},
				*achievement.Status,
				"Achievement %s should have 'accounted' status",
				*achievement.ID,
			)
			accountedCount++
		}

		t.Logf("✅ User %s %s verified %d achievements with 'accounted' status",
			user.LastName, user.FirstName, accountedCount)
	}

	t.Log("✅ All users verified their achievements have 'accounted' status")
	t.Log("🎉 Full workflow scenario completed successfully!")
}
