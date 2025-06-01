//nolint:gosec // this is a test utility.
package testutil

import (
	"context"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	// Import SQLite driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := strings.Builder{}
	b.Grow(length)

	for range length {
		b.WriteByte(charset[rand.IntN(len(charset))])
	}

	return b.String()
}

// RandomUUID generates a random UUID for testing
func RandomUUID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// SetupDatabase creates a test database client
func SetupDatabase(t *testing.T) *ent.Client {
	t.Helper()
	// Use a unique database name for each test to avoid locking issues
	dbName := "file:sesc_test_" + strconv.Itoa(rand.Int()) + "?mode=memory&_fk=1"
	client, err := ent.Open(
		"sqlite3",
		dbName,
	)
	require.NoError(t, err)

	err = client.Schema.Create(t.Context())
	require.NoError(t, err)

	// Ensure database is closed at the end of the test
	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// CreateTestUser creates a test user directly in the database
func CreateTestUser(
	ctx context.Context,
	t *testing.T,
	client *ent.Client,
	firstName, lastName string,
	role sesc.Role,
) sesc.User {
	t.Helper()

	// Create a department first
	deptName := "Test Department " + strconv.Itoa(rand.Int())
	dept, err := client.Department.Create().
		SetName(deptName).
		SetDescription("For testing").
		Save(ctx)
	require.NoError(t, err)

	// Create the user
	userID := uuid.Must(uuid.NewV7())

	_, err = client.User.Create().
		SetID(userID).
		SetFirstName(firstName).
		SetLastName(lastName).
		SetDepartment(dept).
		SetRole(role).
		Save(ctx)
	require.NoError(t, err)

	return sesc.User{
		ID:        userID,
		FirstName: firstName,
		LastName:  lastName,
		Department: sesc.Department{
			ID:   dept.ID,
			Name: deptName,
		},
		Role: role,
	}
}

// CreateTestDepartment creates a test department directly in the database
func CreateTestDepartment(ctx context.Context, t *testing.T, client *ent.Client) sesc.Department {
	t.Helper()

	deptName := "Test Department " + strconv.Itoa(rand.Int())
	dept, err := client.Department.Create().
		SetName(deptName).
		SetDescription("For testing").
		Save(ctx)
	require.NoError(t, err)

	return sesc.Department{
		ID:   dept.ID,
		Name: deptName,
	}
}

// CreateTestAchievementGroup creates a test achievement group directly in the database
func CreateTestAchievementGroup(ctx context.Context, t *testing.T, client *ent.Client) achievement.Group {
	t.Helper()

	groupName := "Test Group " + strconv.Itoa(rand.Int())
	group, err := client.AchievementGroup.Create().
		SetName(groupName).
		SetDescription("For testing").
		Save(ctx)
	require.NoError(t, err)

	return achievement.Group{
		ID:   group.ID,
		Name: groupName,
	}
}

// CreateTestAchievementTemplate creates a test achievement template directly in the database
func CreateTestAchievementTemplate(
	ctx context.Context,
	t *testing.T,
	client *ent.Client,
	kind achievement.Kind,
) achievement.Template {
	t.Helper()

	// Create a group first
	group := CreateTestAchievementGroup(ctx, t, client)

	// Create the template
	templateName := "Test Template " + strconv.Itoa(rand.Int())
	template, err := client.AchievementTemplate.Create().
		SetName(templateName).
		SetDescription("For testing").
		SetPointsLimit(10).
		SetKind(kind).
		SetGroupID(group.ID).
		Save(ctx)
	require.NoError(t, err)

	return achievement.Template{
		ID:          template.ID,
		Name:        templateName,
		Description: "For testing",
		PointsLimit: 10,
		Kind:        kind,
		GroupID:     group.ID,
	}
}

// CreateTestAchievement creates a test achievement directly in the database
func CreateTestAchievement(
	ctx context.Context,
	t *testing.T,
	client *ent.Client,
	user sesc.User,
	template achievement.Template,
	status achievement.Status,
) achievement.Achievement {
	t.Helper()

	achievementID := uuid.Must(uuid.NewV7())
	err := client.Achievement.Create().
		SetID(achievementID).
		SetStatus(string(status)).
		SetPoints(0).
		SetOwnerID(user.ID).
		SetTemplateID(template.ID).
		Exec(ctx)
	require.NoError(t, err)

	return achievement.Achievement{
		ID:       achievementID,
		Status:   status,
		Points:   0,
		Owner:    user,
		Template: template,
	}
}

// CreateTestFile creates a test file entry directly in the database
func CreateTestFile(ctx context.Context, t *testing.T, client *ent.Client) sesc.File {
	t.Helper()

	fileID := uuid.Must(uuid.NewV7())
	fileName := "test-file-" + strconv.Itoa(rand.Int()) + ".pdf"
	objectKey := "test-files/" + fileID.String()

	_, err := client.File.Create().
		SetID(fileID).
		SetName(fileName).
		SetSize(1024).
		SetURL("https://example.com/" + fileName).
		SetS3ObjectKey(objectKey).
		Save(ctx)
	require.NoError(t, err)

	return sesc.File{
		ID:   fileID,
		Name: fileName,
		Size: 1024,
		URL:  "https://example.com/" + fileName,
	}
}

// CreateTestContext creates a new context with an event record for testing
func CreateTestContext(t *testing.T) (context.Context, *event.Record) {
	ctx := t.Context()
	return event.NewRecord(ctx, "test")
}

// TestContext contains shared testing resources
type TestContext struct {
	Client   *ent.Client
	User     sesc.User
	Template achievement.Template
	Group    achievement.Group
	File     sesc.File
	Dept     sesc.Department
}

// SetupTestContext creates a TestContext with all common test resources initialized
func SetupTestContext(t *testing.T) *TestContext {
	t.Helper()

	// Create a new database client
	client := SetupDatabase(t)

	// Create a new context with event recording
	ctx := t.Context()
	ctx, _ = event.NewRecord(ctx, "test")

	// Create department
	dept := CreateTestDepartment(ctx, t, client)

	// Create user with teacher role
	user := CreateTestUser(ctx, t, client, "Test", "User", 1) // 1 is teacher role

	// Create achievement group
	group := CreateTestAchievementGroup(ctx, t, client)

	// Create achievement template
	template := CreateTestAchievementTemplate(ctx, t, client, achievement.Kind("olympiad"))

	// Create file
	file := CreateTestFile(ctx, t, client)

	return &TestContext{
		Client:   client,
		User:     user,
		Template: template,
		Group:    group,
		File:     file,
		Dept:     dept,
	}
}
