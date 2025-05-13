package api

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/iam"
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
		LoginAdmin(ctx context.Context, token string) (string, error)
		// ImWatermelon parses tokenString, returns Identity or error
		ImWatermelon(ctx context.Context, tokenString string) (iam.Identity, error)
		// DropCredentials deletes credentials by userID
		DropCredentials(ctx context.Context, userID uuid.UUID) error
		// Credentials returns username/password for a userID
		Credentials(ctx context.Context, userID uuid.UUID) (iam.Credentials, error)
	}

	SESC interface {
		// UpdateUser updates user with the new fields.
		//
		// Returns an ErrInvalidRole if the new role id is invalid.
		// Returns an ErrInvalidName if the first or last name is missing.
		UpdateUser(ctx context.Context, id sesc.UUID, upd sesc.UserUpdateOptions) (sesc.User, error)
		// CreateUser creates a new User with a specified role.
		//
		// Returns an ErrInvalidName if the first or last name is missing.z
		CreateUser(ctx context.Context, opt sesc.UserUpdateOptions) (sesc.User, error)
		// Return a sesc.DepartmentAlreadyExists if the department already exists
		CreateDepartment(ctx context.Context, name, description string) (sesc.Department, error)
		UpdateDepartment(ctx context.Context, id sesc.UUID, name, description string) error
		// User returns a User by ID. If the user does not exist, returns a sesc.ErrUserNotFound.
		User(ctx context.Context, id sesc.UUID) (sesc.User, error)

		// Users returns all the users currently registered within the system.
		Users(ctx context.Context) ([]sesc.User, error)

		// Departments returns all the departments currently registered within the system.
		Departments(ctx context.Context) ([]sesc.Department, error)
		DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error)
		DeleteDepartment(ctx context.Context, id sesc.UUID) error
		UpdateProfilePicture(ctx context.Context, id sesc.UUID, pictureURL string) error
	}
)
