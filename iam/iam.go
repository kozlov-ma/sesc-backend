package iam

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/authuser"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrCredentialsAlreadyExist = errors.New("user with similar credentials already exists")
	ErrInvalidToken            = errors.New("invalid token")
	ErrUserNotFound            = errors.New("user not found")
)

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
	AuthID uuid.UUID
	Role   Role
	ID     uuid.UUID
}

// IAM handles authentication using Ent for persistence.
type IAM struct {
	client           *ent.Client
	adminCredentials []Credentials
	tokenDuration    time.Duration
	jwtkey           []byte
}

// New creates a new IAM with the given Ent client.
func New(
	client *ent.Client,
	tokenDuration time.Duration,
	adminCredentials []Credentials,
	jwtkey []byte,
) *IAM {
	return &IAM{
		client:           client,
		adminCredentials: adminCredentials,
		tokenDuration:    tokenDuration,
		jwtkey:           jwtkey,
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
	rec := event.Get(ctx).Sub("iam/register_credentials")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"user_id", userID,
		"username", creds.Username,
	)

	if err := creds.Validate(); err != nil {
		rec.Add(events.Error, err)
		return UUID{}, err
	}

	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()

	// Start a transaction with serializable isolation
	tx, err := i.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		txrec.Add(events.Error, err)
		return UUID{}, fmt.Errorf("couldn't start transaction: %w", err)
	}

	rollback := func(err error) (UUID, error) {
		txrec.Set("rollback", true)
		if rbErr := tx.Rollback(); rbErr != nil {
			txrec.Add(events.Error, err)
			txrec.Set("rollback_failed", true)
			return UUID{}, fmt.Errorf("%w: rollback failed: %w", err, rbErr)
		}
		return UUID{}, err
	}

	statrec.Add(events.PostgresQueries, 1)
	userExists, err := tx.User.Query().
		Where(user.ID(userID)).
		Exist(ctx)
	if err != nil {
		err := fmt.Errorf("error checking user existence: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}
	if !userExists {
		rec.Set("user_exists", false)
		return rollback(ErrUserNotFound)
	}

	statrec.Add(events.PostgresQueries, 1)
	exists, err := tx.AuthUser.
		Query().
		Where(authuser.UsernameEQ(creds.Username), authuser.Password(creds.Password)).
		Exist(ctx)
	if err != nil {
		err := fmt.Errorf("failed to check if username exists: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}
	if exists {
		rec.Set("duplicate_credentials", true)
		return rollback(ErrCredentialsAlreadyExist)
	}

	statrec.Add(events.PostgresQueries, 1)
	_, err = tx.AuthUser.Delete().Where(authuser.UserID(userID)).Exec(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't delete old credentials: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	statrec.Add(events.PostgresQueries, 1)
	authID := uuid.Must(uuid.NewV7())
	_, err = tx.AuthUser.
		Create().
		SetUsername(creds.Username).
		SetPassword(creds.Password).
		SetAuthID(authID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't create AuthUser: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	rec.Sub("auth_user").Add(
		"user_id", userID,
		"auth_id", authID,
		"username", creds.Username,
	)

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))
	rec.Set("success", true)

	return authID, nil
}

// Login verifies credentials and returns signed JWT token string.
func (i *IAM) Login(ctx context.Context, creds Credentials) (string, error) {
	rec := event.Get(ctx).Sub("iam/login")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"username", creds.Username,
	)

	if err := creds.Validate(); err != nil {
		rec.Set("credentials_validation_error", true)
		return "", err
	}

	pgTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	authRec, err := i.client.AuthUser.
		Query().
		Where(authuser.Username(creds.Username), authuser.Password(creds.Password)).
		Only(ctx)
	statrec.Add(events.PostgresTime, pgTime)

	if ent.IsNotFound(err) {
		rec.Set("auth_user", nil)
		return "", ErrUserNotFound
	} else if err != nil {
		err := fmt.Errorf("couldn't query databse for auth data: %w", err)
		rec.Set(events.Error, err)
		return "", err
	}

	rec.Sub("auth_user").Set(
		"user_id", authRec.UserID,
		"auth_id", authRec.AuthID,
		"username", authRec.Username,
	)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": authRec.AuthID.String(),
		"role":    string(RoleUser),
		"exp":     time.Now().Add(i.tokenDuration).Unix(),
	})

	signed, err := token.SignedString(i.jwtkey)
	if err != nil {
		err := fmt.Errorf("couldn't sign token: %w", err)
		rec.Add(events.Error, err)
		return "", err
	}

	return signed, nil
}

// LoginAdmin checks token for being an admin token.
// Returns ErrInvalidCredentials if the token is not valid.
func (i *IAM) LoginAdmin(ctx context.Context, creds Credentials) (string, error) {
	rec := event.Get(ctx).Sub("iam/login_admin")

	rec.Sub("params").Set("username", creds.Username)
	rec.Set("credentials_found", false)

	if !slices.Contains(i.adminCredentials, creds) {
		return "", ErrUserNotFound
	}

	rec.Set("credentials_found", true)

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.Nil.String(), // Add user_id claim for admin
		"role":    string(RoleAdmin),
		"exp":     time.Now().Add(i.tokenDuration).Unix(),
	})

	// Use SignedString with jwtKey instead of SigningString
	signed, err := tok.SignedString(i.jwtkey)
	if err != nil {
		err := fmt.Errorf("couldn't sign token: %w", err)
		rec.Add(events.Error, err)
		return "", err
	}

	return signed, nil
}

// ImWatermelon parses tokenString, returns Identity or ErrInvalidToken.
func (i *IAM) ImWatermelon(ctx context.Context, tokenString string) (Identity, error) {
	rec := event.Get(ctx).Sub("iam/im_watermelon")
	statrec := event.Get(ctx).Sub("stats")

	rec.Set("auth_user", nil)
	rec.Set("token_string_valid", false)
	rec.Set("token_content_valid", false)
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return i.jwtkey, nil
	})

	if err != nil || !parsed.Valid {
		rec.Add("error", err)
		return Identity{}, errors.Join(ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, ErrInvalidToken
	}

	rec.Set("token_string_valid", true)

	authIDStr, ok1 := claims["user_id"].(string)
	roleStr, ok2 := claims["role"].(string)
	if !ok1 || !ok2 {
		rec.Sub("claims").Set(
			"auth_id", authIDStr,
			"role_str", roleStr,
		)
		return Identity{}, ErrInvalidToken
	}

	if roleStr == string(RoleAdmin) {
		rec.Set("token_content_valid", true)
		return Identity{
			AuthID: uuid.Nil,
			Role:   RoleAdmin,
		}, nil
	}

	aid, err := uuid.FromString(authIDStr)
	if err != nil {
		return Identity{}, ErrInvalidToken
	}

	statrec.Add(events.PostgresQueries, 1)

	rec.Set("token_content_valid", true)

	pgTime := time.Now()
	res, err := i.client.AuthUser.Query().Where(authuser.AuthID(aid)).Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(pgTime))

	switch {
	case ent.IsNotFound(err):
		return Identity{}, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't get user id from auth id: %w", err)
		rec.Add(events.Error, err)
		return Identity{}, err
	}

	rec.Sub("auth_user").Set(
		"user_id", res.UserID,
		"role", Role(roleStr),
		"auth_id", res.AuthID,
		"username", res.Username,
	)

	identity := Identity{
		AuthID: res.AuthID,
		Role:   Role(roleStr),
		ID:     res.UserID,
	}
	return identity, nil
}

// DropCredentials deletes credentials by userID; returns ErrUserNotFound if credentials missing,
// or ErrUserDoesNotExist if the user doesn't exist.
func (i *IAM) DropCredentials(ctx context.Context, userID UUID) error {
	rec := event.Get(ctx).Sub("iam/drop_credentials")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"user_id", userID,
	)

	txStart := time.Now()
	txrec := rec.Sub("pg_transaction")

	// Start a transaction with serializable isolation
	tx, err := i.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		err := fmt.Errorf("couldn't start transaction: %w", err)
		txrec.Add(events.Error, err)
		return err
	}

	rollback := func(err error) error {
		txrec.Set("rollback", true)
		if rbErr := tx.Rollback(); rbErr != nil {
			txrec.Add(events.Error, err)
			txrec.Set("rollback_failed", true)
			return fmt.Errorf("%w: rollback failed: %w", err, rbErr)
		}
		return err
	}

	rec.Set("user", nil)
	statrec.Add(events.PostgresQueries, 1)
	user, err := tx.User.Get(ctx, userID)
	switch {
	case ent.IsNotFound(err):
		return rollback(ErrUserNotFound)
	case err != nil:
		err := fmt.Errorf("error checking user existence: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	rec.Sub("user").Set(
		"id", user.ID,
		"first_name", user.FirstName,
		"last_name", user.LastName,
		"middle_name", user.MiddleName,
		"suspended", user.Suspended,
		"role_id", user.RoleID,
	)

	rec.Set("auth_user", nil)
	statrec.Add(events.PostgresQueries, 1)
	authUser, err := tx.AuthUser.Query().
		Where(authuser.UserID(userID)).
		Only(ctx)
	switch {
	case ent.IsNotFound(err):
		return rollback(ErrUserNotFound)
	case err != nil:
		err := fmt.Errorf("error checking credentials existence: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	rec.Sub("auth_user").Set(
		"user_id", authUser.UserID,
		"auth_id", authUser.AuthID,
		"username", authUser.Username,
	)

	statrec.Add(events.PostgresQueries, 1)
	_, err = tx.AuthUser.
		Delete().
		Where(authuser.UserID(userID)).
		Exec(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't drop user's credentials: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	err = tx.Commit()
	if err != nil {
		err := fmt.Errorf("couldn't commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return rollback(err)
	}

	statrec.Add(events.PostgresTime, time.Since(txStart))

	return nil
}

func (i *IAM) Credentials(ctx context.Context, userID UUID) (Credentials, error) {
	rec := event.Get(ctx).Sub("iam/credentials")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("user_id", userID)

	statrec.Add(events.PostgresQueries, 1)
	rec.Set("auth_user", nil)

	startTime := time.Now()
	res, err := i.client.AuthUser.Query().Where(authuser.UserID(userID)).Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		return Credentials{}, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't get credentials: %w", err)
		rec.Add(events.Error, err)
		return Credentials{}, err
	}

	return Credentials{
		Username: res.Username,
		Password: res.Password,
	}, nil
}
