//nolint:gosec // this is a test package.
package workflow

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/kozlov-ma/sesc-backend/apiclient/client"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/achievement_groups"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/achievement_templates"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/achievements"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/authentication"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/departments"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/files"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/reports"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/roles"
	"github.com/kozlov-ma/sesc-backend/apiclient/client/users"
	"github.com/kozlov-ma/sesc-backend/apiclient/models"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// TestClient wraps the generated API client for easier testing
type TestClient struct {
	apiClient *client.Apiclient
	authInfo  runtime.ClientAuthInfoWriter
	userToken string
}

// UserInfo stores information about created users
type UserInfo struct {
	ID           string
	FirstName    string
	LastName     string
	MiddleName   string
	Role         string
	RoleID       int64
	DepartmentID string
	Department   string
	Username     string
	Password     string
	Token        string
}

// DepartmentInfo stores information about departments
type DepartmentInfo struct {
	ID          string
	Name        string
	Description string
}

// AchievementGroupInfo stores information about achievement groups
type AchievementGroupInfo struct {
	ID          string
	Name        string
	Description string
}

// AchievementTemplateInfo stores information about achievement templates
type AchievementTemplateInfo struct {
	ID          string
	GroupID     string
	Name        string
	Description string
	PointsLimit int64
	Kind        string
}

// AchievementInfo stores information about achievements
type AchievementInfo struct {
	ID         string
	UserID     string
	TemplateID string
	Title      string
	Status     string
	Points     int64
	Documents  []string
}

// FileInfo stores information about uploaded files
type FileInfo struct {
	ID       string
	UserID   string
	Filename string
}

// NewTestClient creates a new test client
func NewTestClient(host string) *TestClient {
	// If host is empty, try to get from environment
	if host == "" {
		if testAPIURL := os.Getenv("TEST_API_URL"); testAPIURL != "" {
			// Parse URL to extract host
			if u, err := url.Parse(testAPIURL); err == nil {
				host = u.Host
			}
		}
		// Fallback to localhost
		if host == "" {
			host = "localhost:8080"
		}
	}

	transport := httptransport.New(host, "/", []string{"http"})
	apiClient := client.New(transport, strfmt.Default)

	return &TestClient{
		apiClient: apiClient,
	}
}

// LoginAdmin authenticates as admin and stores auth info
func (c *TestClient) LoginAdmin(username, password string) error {
	authParams := authentication.NewPostAuthAdminLoginParams()
	authParams.SetRequest(&models.APICredentialsRequest{
		Username: &username,
		Password: &password,
	})

	authResp, err := c.apiClient.Authentication.PostAuthAdminLogin(authParams)
	if err != nil {
		return fmt.Errorf("admin login failed: %w", err)
	}

	c.authInfo = httptransport.BearerToken(*authResp.Payload.Token)
	return nil
}

// LoginUser authenticates as a regular user
func (c *TestClient) LoginUser(username, password string) (string, error) {
	authParams := authentication.NewPostAuthLoginParams()
	authParams.SetRequest(&models.APICredentialsRequest{
		Username: &username,
		Password: &password,
	})

	authResp, err := c.apiClient.Authentication.PostAuthLogin(authParams)
	if err != nil {
		return "", fmt.Errorf("user login failed: %w", err)
	}

	return *authResp.Payload.Token, nil
}

// SetUserAuth sets authentication for a specific user token
func (c *TestClient) SetUserAuth(token string) {
	c.userToken = token
	c.authInfo = httptransport.BearerToken(token)
}

// CreateDepartment creates a new department
func (c *TestClient) CreateDepartment(name, description string) (*DepartmentInfo, error) {
	createParams := departments.NewPostDepartmentsParams()
	createParams.SetRequest(&models.APICreateDepartmentRequest{
		Name:        &name,
		Description: &description,
	})

	createResp, err := c.apiClient.Departments.PostDepartments(createParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return &DepartmentInfo{
		ID:          *createResp.Payload.ID,
		Name:        *createResp.Payload.Name,
		Description: *createResp.Payload.Description,
	}, nil
}

// GetRoles retrieves all available roles
func (c *TestClient) GetRoles() (map[int64]*models.RespondRole, error) {
	rolesResp, err := c.apiClient.Roles.GetRoles(roles.NewGetRolesParams())
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	roleMap := make(map[int64]*models.RespondRole)
	for _, role := range rolesResp.Payload.Roles {
		roleMap[*role.ID] = (*models.RespondRole)(role)
	}

	return roleMap, nil
}

// CreateUser creates a new user
func (c *TestClient) CreateUser(
	firstName, lastName, middleName, departmentID string,
	roleID int64,
) (*UserInfo, error) {
	pictureURL := ""

	createParams := users.NewPostUsersParams()
	createParams.SetRequest(&models.APICreateUserRequest{
		FirstName:    &firstName,
		LastName:     &lastName,
		MiddleName:   middleName,
		Role:         &roleID,
		PictureURL:   pictureURL,
		DepartmentID: departmentID,
	})

	createResp, err := c.apiClient.Users.PostUsers(createParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserInfo{
		ID:           *createResp.Payload.ID,
		FirstName:    firstName,
		LastName:     lastName,
		MiddleName:   middleName,
		RoleID:       roleID,
		DepartmentID: departmentID,
	}, nil
}

// SetUserCredentials sets login credentials for a user
func (c *TestClient) SetUserCredentials(userID, username, password string) error {
	credParams := authentication.NewPutUsersIDCredentialsParams()
	credParams.SetID(userID)
	credParams.SetRequest(&models.APICredentialsRequest{
		Username: &username,
		Password: &password,
	})

	_, err := c.apiClient.Authentication.PutUsersIDCredentials(credParams, c.authInfo)
	if err != nil {
		return fmt.Errorf("failed to set credentials: %w", err)
	}

	return nil
}

type file struct {
	name string
	*bytes.Buffer
}

func (file) Close() error { return nil }

func (f file) Name() string {
	return f.name
}

// UploadFile uploads a file for the current user
func (c *TestClient) UploadFile(filename string, fileContent []byte) (*FileInfo, error) {
	// Note: This is a simplified version. The actual file upload might need multipart/form-data
	uploadParams := files.NewPostFilesParams()
	uploadParams.SetFile(file{name: filename, Buffer: bytes.NewBuffer(fileContent)})

	uploadResp, err := c.apiClient.Files.PostFiles(uploadParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &FileInfo{
		ID:       uploadResp.Payload.ID,
		Filename: filename,
		UserID:   uploadResp.Payload.OwnerID,
	}, nil
}

// CreateAchievement creates a new achievement
func (c *TestClient) CreateAchievement(templateID string) (*AchievementInfo, error) {
	createParams := achievements.NewPostAchievementsParams()
	createParams.SetRequest(&models.ParamCreateAchievementRequest{
		TemplateID: &templateID,
	})

	createResp, err := c.apiClient.Achievements.PostAchievements(createParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create achievement: %w", err)
	}

	return &AchievementInfo{
		ID:         *createResp.Payload.ID,
		TemplateID: templateID,
		Status:     "created",
	}, nil
}

// GetUserAchievements retrieves achievements for the current user
func (c *TestClient) GetUserAchievements(id *string) ([]*models.RespondAchievement, error) {
	getParams := achievements.NewGetAchievementsParams().WithID(id)

	getResp, err := c.apiClient.Achievements.GetAchievements(getParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	return getResp.Payload.Achievements, nil
}

func (c *TestClient) GetUsersWithAchievements() ([]*models.RespondUser, error) {
	gp := achievements.NewGetAchievementsUsersParams()

	res, err := c.apiClient.Achievements.GetAchievementsUsers(gp, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with achievements: %w", err)
	}

	return res.Payload.Users, nil
}

// SubmitAchievement submits an achievement for review
func (c *TestClient) SubmitAchievement(achievementID string) error {
	submitParams := achievements.NewPostAchievementsIDSubmitParams()
	submitParams.SetID(achievementID)

	_, err := c.apiClient.Achievements.PostAchievementsIDSubmit(submitParams, c.authInfo)
	if err != nil {
		return fmt.Errorf("failed to submit achievement: %w", err)
	}

	return nil
}

// ReviewAchievement reviews an achievement with the specified action
func (c *TestClient) ReviewAchievement(achievementID, action, comment string) error {
	reviewParams := achievements.NewPostAchievementsIDReviewParams()
	reviewParams.SetID(achievementID)
	reviewParams.SetRequest(&models.ParamReviewAchievementRequest{
		Action:  &action,
		Comment: comment,
	})

	_, err := c.apiClient.Achievements.PostAchievementsIDReview(reviewParams, c.authInfo)
	if err != nil {
		return fmt.Errorf("failed to review achievement: %w", err)
	}

	return nil
}

// SubmitAchievementWithNewPoints allows teacher to update points and resubmit
func (c *TestClient) SubmitAchievementWithNewPoints(
	achievementID string,
	points int64,
	comment string,
) error {
	submitParams := achievements.NewPostAchievementsIDSubmitWithNewPointsParams()
	submitParams.SetID(achievementID)
	submitParams.SetRequest(&models.ParamUpdateAchievementPointsRequest{
		Points:  &points,
		Comment: comment,
	})

	_, err := c.apiClient.Achievements.PostAchievementsIDSubmitWithNewPoints(
		submitParams,
		c.authInfo,
	)
	if err != nil {
		return fmt.Errorf("failed to submit achievement with new points: %w", err)
	}

	return nil
}

// MarkAllAccounted marks all achievements as accounted
func (c *TestClient) MarkAllAccounted() error {
	markParams := reports.NewPostReportsMarkAllAccountedParams()

	_, err := c.apiClient.Reports.PostReportsMarkAllAccounted(markParams, c.authInfo)
	if err != nil {
		return fmt.Errorf("failed to mark all as accounted: %w", err)
	}

	return nil
}

// GetDocumentStats retrieves document statistics
func (c *TestClient) GetDocumentStats() (*models.RespondDocumentStats, error) {
	getParams := files.NewGetDocumentsStatsParams()

	getResp, err := c.apiClient.Files.GetDocumentsStats(getParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get document stats: %w", err)
	}

	return getResp.Payload, nil
}

// DeleteFile schedules a file for deletion
func (c *TestClient) DeleteFile(fileID string) error {
	deleteParams := files.NewDeleteFilesIDParams()
	deleteParams.SetID(fileID)

	_, err := c.apiClient.Files.DeleteFilesID(deleteParams, c.authInfo)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ScheduleDeletionAll schedules deletion for all files
func (c *TestClient) ScheduleDeletionAll() error {
	deleteParams := files.NewPostDocumentsScheduleDeletionAllParams()

	_, err := c.apiClient.Files.PostDocumentsScheduleDeletionAll(deleteParams, c.authInfo)
	if err != nil {
		return fmt.Errorf("failed to schedule deletion all: %w", err)
	}

	return nil
}

// CreateAchievementGroup creates a new achievement group
func (c *TestClient) CreateAchievementGroup(
	name, description string,
) (*AchievementGroupInfo, error) {
	createParams := achievement_groups.NewPostAchievementGroupsParams()
	createParams.SetRequest(&models.ParamCreateAchievementGroupRequest{
		Name:        &name,
		Description: &description,
	})

	createResp, err := c.apiClient.AchievementGroups.PostAchievementGroups(createParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create achievement group: %w", err)
	}

	return &AchievementGroupInfo{
		ID:          *createResp.Payload.ID,
		Name:        *createResp.Payload.Name,
		Description: *createResp.Payload.Description,
	}, nil
}

// GetAchievementGroups retrieves all achievement groups
func (c *TestClient) GetAchievementGroups() ([]*models.RespondAchievementGroup, error) {
	getParams := achievement_groups.NewGetAchievementGroupsParams()

	getResp, err := c.apiClient.AchievementGroups.GetAchievementGroups(getParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement groups: %w", err)
	}

	return getResp.Payload, nil
}

// CreateAchievementTemplate creates a new achievement template
func (c *TestClient) CreateAchievementTemplate(
	groupID, name, description string,
	pointsLimit int64,
	kind string,
) (*AchievementTemplateInfo, error) {
	createParams := achievement_templates.NewPostAchievementTemplatesParams()

	var rr int64
	switch kind {
	case "scientific":
		rr = 3
	case "development":
		rr = 4
	case "olympiad":
		rr = 5
	case "academic":
		rr = 6
	default:
		panic("invalid kind")
	}
	createParams.SetRequest(&models.ParamCreateAchievementTemplateRequest{
		GroupID:      &groupID,
		Name:         &name,
		Description:  &description,
		PointsLimit:  &pointsLimit,
		ReviewerRole: &rr,
	})

	createResp, err := c.apiClient.AchievementTemplates.PostAchievementTemplates(
		createParams,
		c.authInfo,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create achievement template: %w", err)
	}

	return &AchievementTemplateInfo{
		ID:          *createResp.Payload.ID,
		GroupID:     groupID,
		Name:        name,
		Description: description,
		PointsLimit: pointsLimit,
		Kind:        kind,
	}, nil
}

// GetAchievementTemplates retrieves all achievement templates
func (c *TestClient) GetAchievementTemplates() ([]*models.RespondAchievementTemplate, error) {
	getParams := achievement_templates.NewGetAchievementTemplatesParams()

	getResp, err := c.apiClient.AchievementTemplates.GetAchievementTemplates(getParams, c.authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement templates: %w", err)
	}

	return getResp.Payload, nil
}

// Helper functions for generating test data

// GenerateRandomUser generates a random user with Russian names
func GenerateRandomUser() (firstName, lastName, middleName string) {
	firstNames := []string{
		"Александр", "Алексей", "Анатолий", "Андрей", "Антон", "Борис",
		"Вадим", "Валентин", "Виктор", "Владимир", "Геннадий", "Дмитрий",
		"Евгений", "Игорь", "Иван", "Кирилл", "Максим", "Михаил",
		"Николай", "Олег", "Павел", "Петр", "Роман", "Сергей",
	}

	lastNames := []string{
		"Иванов", "Смирнов", "Кузнецов", "Попов", "Васильев", "Петров",
		"Соколов", "Михайлов", "Новиков", "Федоров", "Морозов", "Волков",
		"Алексеев", "Лебедев", "Семенов", "Егоров", "Павлов", "Козлов",
		"Степанов", "Николаев", "Орлов", "Андреев", "Макаров", "Никитин",
	}

	patronymics := []string{
		"Александрович", "Алексеевич", "Анатольевич", "Андреевич", "Антонович",
		"Борисович", "Вадимович", "Валентинович", "Викторович", "Владимирович",
		"Геннадьевич", "Дмитриевич", "Евгеньевич", "Игоревич", "Иванович",
		"Кириллович", "Максимович", "Михайлович", "Николаевич", "Олегович",
		"Павлович", "Петрович", "Романович", "Сергеевич",
	}

	firstName = firstNames[rand.IntN(len(firstNames))]
	lastName = lastNames[rand.IntN(len(lastNames))]
	middleName = patronymics[rand.IntN(len(patronymics))]

	return firstName, lastName, middleName
}

// GenerateUsername generates a username from names
func GenerateUsername(firstName, lastName string) string {
	firstLatin := transliterate(firstName)
	lastLatin := transliterate(lastName)
	username := strings.ToLower(firstLatin[:1] + lastLatin)
	username += fmt.Sprintf("%03d", rand.IntN(1000))
	return username
}

// GeneratePassword generates a simple password for testing
func GeneratePassword() string {
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

// GetRoleID returns the role ID for common roles
func GetRoleID(role string) int64 {
	switch role {
	case "teacher":
		return int64(sesc.Teacher)
	case "dephead":
		return int64(sesc.Dephead)
	case "olympiad_deputy":
		return int64(sesc.OlympiadDeputy)
	case "scientific_deputy":
		return int64(sesc.ScientificDeputy)
	case "development_deputy":
		return int64(sesc.DevelopmentDeputy)
	case "academic_director":
		return int64(sesc.AcademicDirector)
	case "economist":
		return int64(sesc.ChiefEconomist)
	default:
		return int64(sesc.Teacher) // Default to teacher
	}
}
