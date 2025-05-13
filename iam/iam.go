package iam

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/authuser"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

// JWT secret key (should be in config).
var jwtKey = []byte("dinahu")

type Credentials struct {
	Username string
	Password string
}

func (c Credentials) Validate() error {
	if c.Username == "" || c.Password == "" {
		return ErrInvalidCredentials
	}
	return nil
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Identity struct {
	ID   uuid.UUID
	Role Role
}

// IAM handles authentication using Ent for persistence.
type IAM struct {
	log           *slog.Logger
	client        *ent.Client
	adminTokens   []string
	tokenDuration time.Duration
}

// New creates a new IAM with the given Ent client.
func New(
	log *slog.Logger,
	client *ent.Client,
	tokenDuration time.Duration,
	adminTokens []string,
) *IAM {
	return &IAM{
		log:           log,
		client:        client,
		adminTokens:   adminTokens,
		tokenDuration: tokenDuration,
	}
}

type UUID = uuid.UUID

// RegisterCredentials assigns username/password to an existing userID, returns authID.
// Returns ErrUserDoesNotExist if user does not exist, ErrUserAlreadyExists if username exists,
// or ErrInvalidCredentials if creds invalid.
func (i *IAM) RegisterCredentials(
	ctx context.Context,
	userID UUID,
	creds Credentials,
) (UUID, error) {
	logger := i.log.With("method", "RegisterCredentials", "user_id", userID)
	logger.DebugContext(ctx, "Starting credentials registration")

	if err := creds.Validate(); err != nil {
		logger.ErrorContext(ctx, "Invalid credentials provided", "error", err)
		return UUID{}, err
	}

	// Start a transaction with serializable isolation
	tx, err := i.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return UUID{}, fmt.Errorf("couldn't start transaction: %w", err)
	}

	// Function to rollback transaction and wrap error
	rollback := func(err error) (UUID, error) {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.ErrorContext(ctx, "Failed to rollback transaction", "error", rbErr)
			return UUID{}, fmt.Errorf("%w: rollback failed: %w", err, rbErr)
		}
		return UUID{}, err
	}

	// Check if the user exists first
	userExists, err := tx.User.Query().
		Where(user.ID(userID)).
		Exist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check if user exists", "error", err)
		return rollback(fmt.Errorf("error checking user existence: %w", err))
	}
	if !userExists {
		logger.ErrorContext(
			ctx,
			"Cannot register credentials for non-existent user",
			"user_id",
			userID,
		)
		return rollback(ErrUserNotFound)
	}

	// Check if the username already exists
	exists, err := tx.AuthUser.
		Query().
		Where(authuser.UsernameEQ(creds.Username)).
		Exist(ctx)
	if err != nil {
		logger.ErrorContext(
			ctx,
			"Failed to check if username exists",
			"username",
			creds.Username,
			"error",
			err,
		)
		return rollback(err)
	}
	if exists {
		logger.ErrorContext(ctx, "User already exists", "username", creds.Username)
		return rollback(ErrUserAlreadyExists)
	}

	// Create the auth user entry
	authID := uuid.Must(uuid.NewV7())
	_, err = tx.AuthUser.
		Create().
		SetUsername(creds.Username).
		SetPassword(creds.Password).
		SetAuthID(authID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create auth user", "error", err)
		return rollback(err)
	}

	err = tx.Commit()
	if err != nil {
		logger.ErrorContext(ctx, "Failed to commit transaction", "error", err)
		return rollback(fmt.Errorf("couldn't commit transaction: %w", err))
	}

	logger.InfoContext(
		ctx,
		"User credentials registered successfully",
		"auth_id",
		authID,
		"username",
		creds.Username,
	)
	return authID, nil
}

// Login verifies credentials and returns signed JWT token string.
func (i *IAM) Login(ctx context.Context, creds Credentials) (string, error) {
	logger := i.log.With("method", "Login", "username", creds.Username)
	logger.DebugContext(ctx, "Attempting user login")

	if err := creds.Validate(); err != nil {
		logger.ErrorContext(ctx, "Invalid credentials provided", "error", err)
		return "", err
	}

	authRec, err := i.client.AuthUser.
		Query().
		Where(authuser.Username(creds.Username), authuser.Password(creds.Password)).
		Only(ctx)
	if ent.IsNotFound(err) {
		logger.ErrorContext(ctx, "User not found during login attempt")
		return "", ErrInvalidCredentials
	} else if err != nil {
		logger.ErrorContext(ctx, "Database error during login", "error", err)
		return "", err
	}

	if authRec.Password != creds.Password {
		logger.ErrorContext(ctx, "Incorrect password provided")
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": authRec.UserID.String(),
		"role":    string(RoleUser),
		"exp":     time.Now().Add(i.tokenDuration).Unix(),
	})

	signed, err := token.SignedString(jwtKey)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to sign token", "error", err)
		return "", fmt.Errorf("couldn't sign token: %w", err)
	}

	logger.InfoContext(ctx, "User logged in successfully", "user_id", authRec.UserID)
	return signed, nil
}

// LoginAdmin checks token for being an admin token.
// Returns ErrInvalidCredentials if the token is not valid.
func (i *IAM) LoginAdmin(ctx context.Context, token string) (string, error) {
	logger := i.log.With("method", "LoginAdmin")
	logger.DebugContext(ctx, "Attempting admin login")

	if !slices.Contains(i.adminTokens, token) {
		logger.ErrorContext(ctx, "Invalid admin token provided")
		return "", ErrInvalidCredentials
	}

	// Create a system UUID for admin
	adminID := uuid.Must(uuid.NewV7())

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": adminID.String(), // Add user_id claim for admin
		"role":    string(RoleAdmin),
		"exp":     time.Now().Add(i.tokenDuration).Unix(),
	})

	// Use SignedString with jwtKey instead of SigningString
	signed, err := tok.SignedString(jwtKey)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to sign admin token", "error", err)
		return "", fmt.Errorf("couldn't sign token: %w", err)
	}

	logger.InfoContext(ctx, "Admin logged in successfully", "admin_id", adminID)
	return signed, nil
}

// ImWatermelon parses tokenString, returns Identity or ErrInvalidToken.
func (i *IAM) ImWatermelon(ctx context.Context, tokenString string) (Identity, error) {
	logger := i.log.With("method", "ImWatermelon")
	logger.DebugContext(ctx, "Verifying token")

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			logger.ErrorContext(ctx, "Invalid token signing method", "method", t.Method)
			return nil, ErrInvalidToken
		}
		return jwtKey, nil
	})

	if err != nil || !parsed.Valid {
		logger.ErrorContext(ctx, "Invalid or expired token", "error", err)
		return Identity{}, errors.Join(ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		logger.ErrorContext(ctx, "Failed to parse token claims")
		return Identity{}, ErrInvalidToken
	}

	userIDstr, ok1 := claims["user_id"].(string)
	roleStr, ok2 := claims["role"].(string)
	if !ok1 || !ok2 {
		logger.ErrorContext(ctx, "Missing required claims in token",
			"has_user_id", ok1, "has_role", ok2)
		return Identity{}, ErrInvalidToken
	}

	uid, err := uuid.FromString(userIDstr)
	if err != nil {
		logger.ErrorContext(
			ctx,
			"Failed to parse user ID from token",
			"user_id_str",
			userIDstr,
			"error",
			err,
		)
		return Identity{}, ErrInvalidToken
	}

	identity := Identity{ID: uid, Role: Role(roleStr)}
	logger.DebugContext(ctx, "Token verified successfully", "user_id", uid, "role", roleStr)
	return identity, nil
}

// DropCredentials deletes credentials by userID; returns ErrUserNotFound if credentials missing,
// or ErrUserDoesNotExist if the user doesn't exist.
func (i *IAM) DropCredentials(ctx context.Context, userID UUID) error {
	logger := i.log.With("method", "DropCredentials", "user_id", userID)
	logger.DebugContext(ctx, "Attempting to drop user credentials")

	// Start a transaction with serializable isolation
	tx, err := i.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return fmt.Errorf("couldn't start transaction: %w", err)
	}

	// Function to rollback transaction and wrap error
	rollback := func(err error) error {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.ErrorContext(ctx, "Failed to rollback transaction", "error", rbErr)
			return fmt.Errorf("%w: rollback failed: %w", err, rbErr)
		}
		return err
	}

	// Check if the user exists first
	userExists, err := tx.User.Query().
		Where(user.ID(userID)).
		Exist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check if user exists", "error", err)
		return rollback(fmt.Errorf("error checking user existence: %w", err))
	}
	if !userExists {
		logger.ErrorContext(ctx, "Cannot drop credentials for non-existent user", "user_id", userID)
		return rollback(ErrUserNotFound)
	}

	// Check if credentials exist before trying to delete
	credExists, err := tx.AuthUser.Query().
		Where(authuser.UserID(userID)).
		Exist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check if credentials exist", "error", err)
		return rollback(fmt.Errorf("error checking credentials existence: %w", err))
	}
	if !credExists {
		logger.ErrorContext(ctx, "User credentials not found", "user_id", userID)
		return rollback(ErrUserNotFound)
	}

	// Delete the credentials
	_, err = tx.AuthUser.
		Delete().
		Where(authuser.UserID(userID)).
		Exec(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to drop user credentials", "error", err)
		return rollback(fmt.Errorf("couldn't drop user's credentials: %w", err))
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		logger.ErrorContext(ctx, "Failed to commit transaction", "error", err)
		return rollback(fmt.Errorf("couldn't commit transaction: %w", err))
	}

	logger.InfoContext(ctx, "User credentials dropped successfully")
	return nil
}

func (i *IAM) Credentials(ctx context.Context, userID UUID) (Credentials, error) {
	logger := i.log.With("method", "Credentials", "user_id", userID)
	logger.DebugContext(ctx, "Retrieving user credentials")

	// Check if the user exists first
	userExists, err := i.client.User.Query().
		Where(user.ID(userID)).
		Exist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check if user exists", "error", err)
		return Credentials{}, fmt.Errorf("error checking user existence: %w", err)
	}
	if !userExists {
		logger.ErrorContext(
			ctx,
			"Cannot retrieve credentials for non-existent user",
			"user_id",
			userID,
		)
		return Credentials{}, ErrUserNotFound
	}

	res, err := i.client.AuthUser.Query().Where(authuser.UserID(userID)).Only(ctx)
	switch {
	case ent.IsNotFound(err):
		logger.ErrorContext(ctx, "User not found when retrieving credentials")
		return Credentials{}, ErrUserNotFound
	case err != nil:
		logger.ErrorContext(ctx, "Failed to retrieve user credentials", "error", err)
		return Credentials{}, fmt.Errorf("couldn't get credentials: %w", err)
	}

	logger.DebugContext(ctx, "User credentials retrieved successfully")
	return Credentials{
		Username: res.Username,
		Password: res.Password,
	}, nil
}
