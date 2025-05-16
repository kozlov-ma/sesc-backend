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
	statrec := event.Root(ctx).Sub("stats")

	rec.Sub("params").Set(
		"user_id", userID,
		"username", creds.Username,
	)

	// Stage 1: Validate credentials
	ctx = rec.Sub("validate_credentials").Wrap(ctx)
	if err := i.validateCredentials(ctx, creds); err != nil {
		return UUID{}, err
	}

	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)

	txStart := time.Now()

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

	// Stage 2: Check if user exists
	ctx = rec.Sub("check_user_exists").Wrap(ctx)
	if err := i.checkUserExists(ctx, tx, userID); err != nil {
		return rollback(err)
	}

	// Stage 3: Check if username is free
	ctx = rec.Sub("check_username_free").Wrap(ctx)
	if err := i.checkUsernameFree(ctx, tx, creds.Username); err != nil {
		return rollback(err)
	}

	// Stage 4: Delete old credentials
	ctx = rec.Sub("delete_old_credentials").Wrap(ctx)
	if err := i.deleteOldCredentials(ctx, tx, userID); err != nil {
		return rollback(err)
	}

	// Stage 5: Create auth record
	ctx = rec.Sub("create_auth_record").Wrap(ctx)
	authID, err := i.createAuthRecord(ctx, tx, userID, creds)
	if err != nil {
		return rollback(err)
	}

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

// validateCredentials validates the credentials
func (i *IAM) validateCredentials(
	ctx context.Context,
	creds Credentials,
) error {
	rec := event.Get(ctx)
	rec.Set("username", creds.Username)

	if err := creds.Validate(); err != nil {
		rec.Add(events.Error, err)
		rec.Set("valid", false)
		return err
	}

	rec.Set("valid", true)
	return nil
}

// checkUserExists checks if the user exists in the database
func (i *IAM) checkUserExists(
	ctx context.Context,
	tx *ent.Tx,
	userID UUID,
) error {
	rec := event.Get(ctx)
	rec.Set("user_id", userID)

	userExists, err := tx.User.Query().
		Where(user.ID(userID)).
		Exist(ctx)
	if err != nil {
		err := fmt.Errorf("error checking user existence: %w", err)
		rec.Add(events.Error, err)
		rec.Set("exists", false)
		return err
	}

	if !userExists {
		rec.Set("exists", false)
		return ErrUserNotFound
	}

	rec.Set("exists", true)
	return nil
}

// checkUsernameFree checks if the username is available
func (i *IAM) checkUsernameFree(
	ctx context.Context,
	tx *ent.Tx,
	username string,
) error {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Set("username", username)

	statrec.Add(events.PostgresQueries, 1)
	exists, err := tx.AuthUser.
		Query().
		Where(authuser.UsernameEQ(username)).
		Exist(ctx)
	if err != nil {
		err := fmt.Errorf("failed to check if username exists: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	if exists {
		rec.Set("is_free", false)
		return ErrCredentialsAlreadyExist
	}

	rec.Set("is_free", true)
	return nil
}

// deleteOldCredentials deletes any existing credentials for the user
func (i *IAM) deleteOldCredentials(
	ctx context.Context,
	tx *ent.Tx,
	userID UUID,
) error {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Set("user_id", userID)

	statrec.Add(events.PostgresQueries, 1)
	result, err := tx.AuthUser.Delete().Where(authuser.UserID(userID)).Exec(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't delete old credentials: %w", err)
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("deleted_count", result)
	return nil
}

// createAuthRecord creates a new authentication record for the user
func (i *IAM) createAuthRecord(
	ctx context.Context,
	tx *ent.Tx,
	userID UUID,
	creds Credentials,
) (UUID, error) {
	rec := event.Get(ctx)
	rootRec := event.Root(ctx)
	statrec := rootRec.Sub("stats")

	rec.Set(
		"user_id", userID,
		"username", creds.Username,
	)

	statrec.Add(events.PostgresQueries, 1)
	authID := uuid.Must(uuid.NewV7())
	_, err := tx.AuthUser.
		Create().
		SetUsername(creds.Username).
		SetPassword(creds.Password).
		SetAuthID(authID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't create AuthUser: %w", err)
		rec.Add(events.Error, err)
		return UUID{}, err
	}

	rec.Set("auth_id", authID)
	return authID, nil
}

// Login verifies credentials and returns signed JWT token string.
func (i *IAM) Login(ctx context.Context, creds Credentials) (string, error) {
	rec := event.Get(ctx).Sub("iam/login")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set(
		"username", creds.Username,
	)

	// Stage 1: Validate credentials
	if err := i.validateLoginCredentials(ctx, rec, creds); err != nil {
		return "", err
	}

	// Stage 2: Find auth record
	authRec, err := i.findAuthRecord(ctx, rec, statrec, creds)
	if err != nil {
		return "", err
	}

	// Stage 3: Generate token
	token, err := i.generateUserToken(ctx, rec, authRec)
	if err != nil {
		return "", err
	}

	return token, nil
}

// validateLoginCredentials validates the login credentials
func (i *IAM) validateLoginCredentials(
	_ context.Context,
	rec *event.Record,
	creds Credentials,
) error {
	rec = rec.Sub("validate_credentials")
	rec.Set("username", creds.Username)

	if err := creds.Validate(); err != nil {
		rec.Set("valid", false)
		rec.Add(events.Error, err)
		return err
	}

	rec.Set("valid", true)
	return nil
}

// findAuthRecord finds the auth record for the given credentials
func (i *IAM) findAuthRecord(
	ctx context.Context,
	rec *event.Record,
	statrec *event.Record,
	creds Credentials,
) (*ent.AuthUser, error) {
	rec = rec.Sub("find_auth_record")
	rec.Set("username", creds.Username)

	pgTime := time.Now()
	statrec.Add(events.PostgresQueries, 1)
	authRec, err := i.client.AuthUser.
		Query().
		Where(authuser.Username(creds.Username), authuser.Password(creds.Password)).
		Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(pgTime))

	if ent.IsNotFound(err) {
		rec.Set("found", false)
		return nil, ErrUserNotFound
	} else if err != nil {
		err := fmt.Errorf("couldn't query database for auth data: %w", err)
		rec.Add(events.Error, err)
		rec.Set("found", false)
		return nil, err
	}

	rec.Set(
		"found", true,
		"user_id", authRec.UserID,
		"auth_id", authRec.AuthID,
		"username", authRec.Username,
	)

	return authRec, nil
}

// generateUserToken generates a JWT token for a user
func (i *IAM) generateUserToken(
	_ context.Context,
	rec *event.Record,
	authRec *ent.AuthUser,
) (string, error) {
	rec = rec.Sub("generate_token")
	rec.Set(
		"auth_id", authRec.AuthID,
		"role", string(RoleUser),
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
		rec.Set("success", false)
		return "", err
	}

	rec.Set("success", true)
	return signed, nil
}

// LoginAdmin checks token for being an admin token.
// Returns ErrInvalidCredentials if the token is not valid.
func (i *IAM) LoginAdmin(ctx context.Context, creds Credentials) (string, error) {
	rec := event.Get(ctx).Sub("iam/login_admin")

	rec.Sub("params").Set("username", creds.Username)

	// Stage 1: Verify admin credentials
	if err := i.verifyAdminCredentials(ctx, rec, creds); err != nil {
		return "", err
	}

	// Stage 2: Generate admin token
	token, err := i.generateAdminToken(ctx, rec)
	if err != nil {
		return "", err
	}

	return token, nil
}

// verifyAdminCredentials verifies if the credentials match admin credentials
func (i *IAM) verifyAdminCredentials(
	_ context.Context,
	rec *event.Record,
	creds Credentials,
) error {
	rec = rec.Sub("verify_admin_credentials")
	rec.Set("username", creds.Username)

	if !slices.Contains(i.adminCredentials, creds) {
		rec.Set("valid", false)
		return ErrUserNotFound
	}

	rec.Set("valid", true)
	return nil
}

// generateAdminToken generates a JWT token for an admin
func (i *IAM) generateAdminToken(
	_ context.Context,
	rec *event.Record,
) (string, error) {
	rec = rec.Sub("generate_admin_token")

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
		rec.Set("success", false)
		return "", err
	}

	rec.Set("success", true)
	return signed, nil
}

// ImWatermelon parses tokenString, returns Identity or ErrInvalidToken.
func (i *IAM) ImWatermelon(ctx context.Context, tokenString string) (Identity, error) {
	rec := event.Get(ctx).Sub("iam/im_watermelon")
	statrec := event.Get(ctx).Sub("stats")

	// Stage 1: Parse and validate token
	claims, err := i.parseAndValidateToken(rec, tokenString)
	if err != nil {
		return Identity{}, err
	}

	// Stage 2: Extract and validate claims
	authIDStr, roleStr, err := i.extractTokenClaims(rec, claims)
	if err != nil {
		return Identity{}, err
	}

	// Handle admin token
	if roleStr == string(RoleAdmin) {
		return Identity{
			AuthID: uuid.Nil,
			Role:   RoleAdmin,
		}, nil
	}

	// Stage 3: Retrieve auth user for normal user
	identity, err := i.retrieveUserIdentity(ctx, rec, statrec, authIDStr, roleStr)
	if err != nil {
		return Identity{}, err
	}

	return identity, nil
}

// parseAndValidateToken parses and validates the JWT token
func (i *IAM) parseAndValidateToken(
	rec *event.Record,
	tokenString string,
) (jwt.MapClaims, error) {
	rec = rec.Sub("parse_token")

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return i.jwtkey, nil
	})

	if err != nil || !parsed.Valid {
		rec.Add(events.Error, err)
		rec.Set("valid", false)
		return nil, errors.Join(ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		rec.Set("valid", false)
		return nil, ErrInvalidToken
	}

	rec.Set("valid", true)
	return claims, nil
}

// extractTokenClaims extracts and validates token claims
func (i *IAM) extractTokenClaims(
	rec *event.Record,
	claims jwt.MapClaims,
) (string, string, error) {
	rec = rec.Sub("extract_claims")

	authIDStr, ok1 := claims["user_id"].(string)
	roleStr, ok2 := claims["role"].(string)
	if !ok1 || !ok2 {
		rec.Set(
			"valid", false,
			"auth_id", authIDStr,
			"role_str", roleStr,
		)
		return "", "", ErrInvalidToken
	}

	rec.Set(
		"valid", true,
		"auth_id", authIDStr,
		"role", roleStr,
	)

	return authIDStr, roleStr, nil
}

// retrieveUserIdentity retrieves the user identity from the database
func (i *IAM) retrieveUserIdentity(
	ctx context.Context,
	rec *event.Record,
	statrec *event.Record,
	authIDStr string,
	roleStr string,
) (Identity, error) {
	rec = rec.Sub("retrieve_identity")

	aid, err := uuid.FromString(authIDStr)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Set("valid_uuid", false)
		return Identity{}, ErrInvalidToken
	}

	rec.Set("valid_uuid", true)
	statrec.Add(events.PostgresQueries, 1)

	pgTime := time.Now()
	res, err := i.client.AuthUser.Query().Where(authuser.AuthID(aid)).Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(pgTime))

	switch {
	case ent.IsNotFound(err):
		rec.Set("found", false)
		return Identity{}, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't get user id from auth id: %w", err)
		rec.Add(events.Error, err)
		rec.Set("found", false)
		return Identity{}, err
	}

	rec.Set(
		"found", true,
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

	// Stage 1: Check if user exists
	if err := i.checkUserExistsForDrop(ctx, rec, statrec, tx, userID); err != nil {
		return rollback(err)
	}

	// Stage 2: Check if credentials exist
	_, err = i.checkCredentialsExist(ctx, rec, statrec, tx, userID)
	if err != nil {
		return rollback(err)
	}

	// Stage 3: Delete credentials
	if err := i.deleteCredentials(ctx, rec, statrec, tx, userID); err != nil {
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

// checkUserExistsForDrop checks if the user exists in the database
func (i *IAM) checkUserExistsForDrop(
	ctx context.Context,
	rec *event.Record,
	statrec *event.Record,
	tx *ent.Tx,
	userID UUID,
) error {
	rec = rec.Sub("drop_check_user_exists")
	rec.Set("user_id", userID)

	statrec.Add(events.PostgresQueries, 1)
	user, err := tx.User.Get(ctx, userID)
	switch {
	case ent.IsNotFound(err):
		rec.Set("exists", false)
		return ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("error checking user existence: %w", err)
		rec.Add(events.Error, err)
		rec.Set("exists", false)
		return err
	}

	rec.Set(
		"exists", true,
		"id", user.ID,
		"first_name", user.FirstName,
		"last_name", user.LastName,
		"middle_name", user.MiddleName,
		"suspended", user.Suspended,
		"role_id", user.RoleID,
	)

	return nil
}

// checkCredentialsExist checks if credentials exist for the user
func (i *IAM) checkCredentialsExist(
	ctx context.Context,
	rec *event.Record,
	statrec *event.Record,
	tx *ent.Tx,
	userID UUID,
) (*ent.AuthUser, error) {
	rec = rec.Sub("drop_check_credentials_exist")
	rec.Set("user_id", userID)

	statrec.Add(events.PostgresQueries, 1)
	authUser, err := tx.AuthUser.Query().
		Where(authuser.UserID(userID)).
		Only(ctx)
	switch {
	case ent.IsNotFound(err):
		rec.Set("exists", false)
		return nil, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("error checking credentials existence: %w", err)
		rec.Add(events.Error, err)
		rec.Set("exists", false)
		return nil, err
	}

	rec.Set(
		"exists", true,
		"user_id", authUser.UserID,
		"auth_id", authUser.AuthID,
		"username", authUser.Username,
	)

	return authUser, nil
}

// deleteCredentials deletes the credentials for the user
func (i *IAM) deleteCredentials(
	ctx context.Context,
	rec *event.Record,
	statrec *event.Record,
	tx *ent.Tx,
	userID UUID,
) error {
	rec = rec.Sub("drop_delete_credentials")
	rec.Set("user_id", userID)

	statrec.Add(events.PostgresQueries, 1)
	result, err := tx.AuthUser.
		Delete().
		Where(authuser.UserID(userID)).
		Exec(ctx)
	if err != nil {
		err := fmt.Errorf("couldn't drop user's credentials: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return err
	}

	rec.Set("success", true)
	rec.Set("deleted_count", result)
	return nil
}

func (i *IAM) Credentials(ctx context.Context, userID UUID) (Credentials, error) {
	rec := event.Get(ctx).Sub("iam/credentials")
	statrec := event.Get(ctx).Sub("stats")

	rec.Sub("params").Set("user_id", userID)

	// Stage 1: Query credentials
	authUser, err := i.queryUserCredentials(ctx, rec, statrec, userID)
	if err != nil {
		return Credentials{}, err
	}

	// Create and return credentials
	return Credentials{
		Username: authUser.Username,
		Password: authUser.Password,
	}, nil
}

// queryUserCredentials queries the credentials for a user
func (i *IAM) queryUserCredentials(
	ctx context.Context,
	rec *event.Record,
	statrec *event.Record,
	userID UUID,
) (*ent.AuthUser, error) {
	rec = rec.Sub("query_credentials")
	rec.Set("user_id", userID)

	statrec.Add(events.PostgresQueries, 1)

	startTime := time.Now()
	res, err := i.client.AuthUser.Query().Where(authuser.UserID(userID)).Only(ctx)
	statrec.Add(events.PostgresTime, time.Since(startTime))

	switch {
	case ent.IsNotFound(err):
		rec.Set("found", false)
		return nil, ErrUserNotFound
	case err != nil:
		err := fmt.Errorf("couldn't get credentials: %w", err)
		rec.Add(events.Error, err)
		rec.Set("found", false)
		return nil, err
	}

	rec.Set(
		"found", true,
		"user_id", res.UserID,
		"auth_id", res.AuthID,
		"username", res.Username,
	)

	return res, nil
}
