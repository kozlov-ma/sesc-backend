package api

import (
	"bytes"
	"context"
	"io"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
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
		Departments(ctx context.Context) (ent.Departments, error)
		DepartmentByID(ctx context.Context, id uuid.UUID) (*ent.Department, error)
		CreateDepartment(ctx context.Context, name string, description string) (*ent.Department, error)
		UpdateDepartment(ctx context.Context, id uuid.UUID, name string, description string) (*ent.Department, error)
		DeleteDepartment(ctx context.Context, id uuid.UUID) error

		// User operations
		Users(ctx context.Context, offset, limit int, search string) (ent.Users, int, error)
		User(ctx context.Context, id uuid.UUID) (*ent.User, error)
		CreateUser(ctx context.Context, options sesc.UserUpdateOptions) (*ent.User, error)
		UpdateUser(ctx context.Context, id uuid.UUID, options sesc.UserUpdateOptions) (*ent.User, error)
		UpdateProfilePicture(ctx context.Context, id uuid.UUID, pictureURL string) error

		// Achievement group operations
		AchievementGroups(
			ctx context.Context,
			options achievement.GroupSearchOptions,
		) (ent.AchievementGroups, error)
		AchievementGroupByID(ctx context.Context, id uuid.UUID) (*ent.AchievementGroup, error)
		CreateAchievementGroup(
			ctx context.Context,
			options achievement.GroupCreateOptions,
		) (*ent.AchievementGroup, error)
		UpdateAchievementGroup(
			ctx context.Context,
			id uuid.UUID,
			options achievement.GroupUpdateOptions,
		) (*ent.AchievementGroup, error)

		// Achievement template operations
		AchievementTemplates(
			ctx context.Context,
			options achievement.TemplateSearchOptions,
		) (*ent.AchievementTemplate, error)
		AchievementTemplateByID(ctx context.Context, id uuid.UUID) (*ent.AchievementTemplate, error)
		CreateAchievementTemplate(
			ctx context.Context,
			options achievement.TemplateCreateOptions,
		) (*ent.AchievementTemplate, error)
		UpdateAchievementTemplate(
			ctx context.Context,
			id uuid.UUID,
			options achievement.TemplateUpdateOptions,
		) (*ent.AchievementTemplate, error)

		GetAchievement(ctx context.Context, id uuid.UUID) (*ent.Achievement, error)
		GetUserAchievements(
			ctx context.Context,
			userID uuid.UUID,
			offset, limit int,
		) (ent.Achievements, int, error)
		GetUsersWithAchievements(
			ctx context.Context,
			whosAsking uuid.UUID,
			offset, limit int,
		) (ent.Users, int, error)

		CreateAchievement(ctx context.Context, opt achievement.CreateOptions) (*ent.Achievement, error)
		DeleteAchievement(ctx context.Context, opt achievement.DeleteOptions) error
		AddDocument(ctx context.Context, opt achievement.AddDocumentOptions) (*ent.AchievementDocument, error)
		RemoveDocument(ctx context.Context, opt achievement.RemoveDocumentOptions) error

		SubmitAchievement(
			ctx context.Context,
			opt achievement.SubmitOptions,
		) (*ent.Achievement, error)

		ReviewAchievement(
			ctx context.Context,
			opt achievement.ReviewOptions,
		) (*ent.Achievement, error)

		// GenerateUserPointsReport generates an Excel report with user achievement points
		GenerateUserPointsReport(ctx context.Context) (*bytes.Buffer, error)

		// MarkAllDoneAchievementsAsAccounted marks all achievements with "done" status as "accounted"
		MarkAllDoneAchievementsAsAccounted(ctx context.Context) (int, error)
	}

	// FileService defines the file operations interface required by the API
	FileService interface {
		// Search searches for files with the given options
		Search(ctx context.Context, opts sesc.FileSearchOptions) (ent.Files, int, error)
		// Create uploads a new file
		Create(ctx context.Context, reader io.Reader, opts sesc.FileCreateOptions) (*ent.File, error)
		// Delete deletes a file
		Delete(ctx context.Context, id uuid.UUID) error
		// ByID returns a file by its ID
		ByID(ctx context.Context, id uuid.UUID) (*ent.File, error)
	}

	// EventSink is used by the API to log events
	EventSink interface {
		RecordHTTPRequest(ctx context.Context, event *event.Record)
	}
)
