package api

import (
	"bytes"
	"context"
	"io"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (

	// IAMService defines the authentication interface required by the API
	IAMService interface {
		// RegisterCredentials assigns username/password to an existing userID, returns authID.
		// Returns ErrUserDoesNotExist if user does not exist, ErrUserAlreadyExists if username exists,
		// or ErrInvalidCredentials if credentials are invalid.
		RegisterCredentials(
			ctx context.Context,
			userID uuid.UUID,
			creds iam.Credentials,
		) (uuid.UUID, error)
		// Login verifies credentials and returns signed JWT token string
		Login(ctx context.Context, creds iam.Credentials) (string, error)
		// LoginAdmin checks token for being an admin token
		LoginAdmin(ctx context.Context, creds iam.Credentials) (string, error)
		// ImWatermelon parses tokenString, returns Identity or error
		ImWatermelon(ctx context.Context, tokenString string) (iam.Identity, error)
		// DropCredentials deletes credentials by userID
		DropCredentials(ctx context.Context, userID uuid.UUID) error
		// Credentials returns username/password for a userID
		Credentials(ctx context.Context, userID uuid.UUID) (iam.Credentials, error)
	}

	// SESC defines the business logic interface required by the API
	SESC interface {
		// Department operations
		Departments(ctx context.Context) ([]sesc.Department, error)
		DepartmentByID(ctx context.Context, id uuid.UUID) (sesc.Department, error)
		CreateDepartment(ctx context.Context, name string, description string) (sesc.Department, error)
		UpdateDepartment(ctx context.Context, id uuid.UUID, name string, description string) error
		DeleteDepartment(ctx context.Context, id uuid.UUID) error

		// User operations
		Users(ctx context.Context) ([]sesc.User, error)
		User(ctx context.Context, id uuid.UUID) (sesc.User, error)
		UserByID(ctx context.Context, id uuid.UUID) (sesc.User, error)
		CreateUser(ctx context.Context, options sesc.UserUpdateOptions) (sesc.User, error)
		UpdateUser(ctx context.Context, id uuid.UUID, options sesc.UserUpdateOptions) (sesc.User, error)
		UpdateProfilePicture(ctx context.Context, id uuid.UUID, pictureURL string) error

		// Achievement group operations
		AchievementGroups(
			ctx context.Context,
			options achievement.GroupSearchOptions,
		) ([]achievement.Group, error)
		AchievementGroupByID(ctx context.Context, id uuid.UUID) (achievement.Group, error)
		CreateAchievementGroup(
			ctx context.Context,
			options achievement.GroupCreateOptions,
		) (achievement.Group, error)
		UpdateAchievementGroup(
			ctx context.Context,
			id uuid.UUID,
			options achievement.GroupUpdateOptions,
		) (achievement.Group, error)

		// Achievement template operations
		AchievementTemplates(
			ctx context.Context,
			options achievement.TemplateSearchOptions,
		) ([]achievement.Template, error)
		AchievementTemplateByID(ctx context.Context, id uuid.UUID) (achievement.Template, error)
		CreateAchievementTemplate(
			ctx context.Context,
			options achievement.TemplateCreateOptions,
		) (achievement.Template, error)
		UpdateAchievementTemplate(
			ctx context.Context,
			id uuid.UUID,
			options achievement.TemplateUpdateOptions,
		) (achievement.Template, error)

		GetAchievement(ctx context.Context, id uuid.UUID) (achievement.Achievement, error)
		GetUserAchievements(
			ctx context.Context,
			userID uuid.UUID,
			offset, limit int,
		) ([]achievement.Achievement, int, error)
		GetAchievementsForUser(
			ctx context.Context,
			userID uuid.UUID,
			offset, limit int,
		) ([]achievement.Achievement, int, error)
		GetGroupedAchievements(
			ctx context.Context,
			offset, limit int,
		) (map[uuid.UUID][]achievement.Achievement, int, error)

		CreateAchievement(ctx context.Context, opt achievement.CreateOptions) (achievement.Achievement, error)
		DeleteAchievement(ctx context.Context, opt achievement.DeleteOptions) error
		AddDocument(ctx context.Context, opt achievement.AddDocumentOptions) (achievement.Document, error)
		RemoveDocument(ctx context.Context, opt achievement.RemoveDocumentOptions) error

		SubmitAchievement(
			ctx context.Context,
			opt achievement.SubmitOptions,
		) (achievement.Achievement, error)

		ReviewAchievement(
			ctx context.Context,
			opt achievement.ReviewOptions,
		) (achievement.Achievement, error)

		// GenerateUserPointsReport generates an Excel report with user achievement points
		GenerateUserPointsReport(ctx context.Context) (*bytes.Buffer, error)

		// MarkAchievementsAsAccounted marks achievements with "done" status as "accounted"
		MarkAchievementsAsAccounted(ctx context.Context, achievementIDs []uuid.UUID) error

		// MarkAllDoneAchievementsAsAccounted marks all achievements with "done" status as "accounted"
		MarkAllDoneAchievementsAsAccounted(ctx context.Context) (int, error)
	}

	// FileService defines the file operations interface required by the API
	FileService interface {
		// Search searches for files with the given options
		Search(ctx context.Context, opts sesc.FileSearchOptions) ([]sesc.File, int, error)
		// Create uploads a new file
		Create(ctx context.Context, reader io.Reader, opts sesc.FileCreateOptions) (sesc.File, error)
		// Delete deletes a file
		Delete(ctx context.Context, id uuid.UUID) error
		// ByID returns a file by its ID
		ByID(ctx context.Context, id uuid.UUID) (sesc.File, error)
	}

	// EventSink is used by the API to log events
	EventSink interface {
		RecordHTTPRequest(ctx context.Context, event *event.Record)
	}
)
