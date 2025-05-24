package api

import (
	"context"
	"io"

	"github.com/gofrs/uuid/v5"
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
			options sesc.AchievementGroupSearchOptions,
		) ([]sesc.AchievementGroup, error)
		AchievementGroupByID(ctx context.Context, id uuid.UUID) (sesc.AchievementGroup, error)
		CreateAchievementGroup(
			ctx context.Context,
			options sesc.AchievementGroupCreateOptions,
		) (sesc.AchievementGroup, error)
		UpdateAchievementGroup(
			ctx context.Context,
			id uuid.UUID,
			options sesc.AchievementGroupUpdateOptions,
		) (sesc.AchievementGroup, error)

		// Achievement template operations
		AchievementTemplates(
			ctx context.Context,
			options sesc.AchievementTemplateSearchOptions,
		) ([]sesc.AchievementTemplate, error)
		AchievementTemplateByID(ctx context.Context, id uuid.UUID) (sesc.AchievementTemplate, error)
		CreateAchievementTemplate(
			ctx context.Context,
			options sesc.AchievementTemplateCreateOptions,
		) (sesc.AchievementTemplate, error)
		UpdateAchievementTemplate(
			ctx context.Context,
			id uuid.UUID,
			options sesc.AchievementTemplateUpdateOptions,
		) (sesc.AchievementTemplate, error)
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
