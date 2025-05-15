package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type CredentialsRequest struct {
	Username string `json:"username" example:"johndoe"   validate:"required"`
	Password string `json:"password" example:"secret123" validate:"required"`
}

type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
}

type IdentityResponse struct {
	ID   uuid.UUID `json:"id"   example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Role string    `json:"role" example:"user"                                 validate:"required"`
}

// RegisterUser godoc
// @Summary Register user credentials
// @Description Assigns username/password credentials to an existing user
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Param request body CredentialsRequest true "User credentials"
// @Success 201 {object} map[string]uuid.UUID "AuthID"
// @Failure 400 {object} InvalidUUIDError "Invalid UUID format"
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 400 {object} InvalidCredentialsError "Invalid credentials format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} UserNotFoundError "User does not exist"
// @Failure 409 {object} UserExistsError "User already exists"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /users/{id}/credentials [put]
func (a *API) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	// Get userID from path parameter
	idStr := r.PathValue("id")
	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, ErrInvalidUUID, http.StatusBadRequest)
		return
	}

	var credsReq CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&credsReq); err != nil {
		writeError(ctx, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	creds := iam.Credentials{
		Username: credsReq.Username,
		Password: credsReq.Password,
	}

	authID, err := a.iam.RegisterCredentials(ctx, userID, creds)
	switch {
	case errors.Is(err, iam.ErrInvalidCredentials):
		writeError(ctx, w, ErrInvalidCredentials, http.StatusBadRequest)
		return
	case errors.Is(err, iam.ErrUserNotFound):
		writeError(ctx, w, ErrUserNotFound, http.StatusNotFound)
		return
	case errors.Is(err, iam.ErrCredentialsAlreadyExist):
		writeError(ctx, w, ErrUserExists, http.StatusConflict)
		return
	case err != nil:
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails("Failed to register user"), http.StatusInternalServerError)
		return
	}

	a.writeJSON(ctx, w, map[string]uuid.UUID{"authId": authID}, http.StatusCreated)
}

// Login godoc
// @Summary User login
// @Description Verifies user credentials and returns a JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body CredentialsRequest true "User credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} CredentialsNotFoundError "Invalid credentials or user does not exist"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /auth/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var credsReq CredentialsRequest

	if err := json.NewDecoder(r.Body).Decode(&credsReq); err != nil {
		writeError(ctx, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	creds := iam.Credentials{
		Username: credsReq.Username,
		Password: credsReq.Password,
	}

	token, err := a.iam.Login(ctx, creds)
	switch {
	case errors.Is(err, iam.ErrUserNotFound):
		writeError(
			ctx,
			w,
			ErrCredentialsNotFound.WithDetails("Invalid username or password"),
			http.StatusUnauthorized,
		)
		return
	case err != nil:
		writeError(ctx, w, ErrServerError.WithDetails("Failed to login"), http.StatusInternalServerError)
		return
	}

	a.writeJSON(ctx, w, TokenResponse{Token: token}, http.StatusOK)
}

// LoginAdmin godoc
// @Summary Admin login
// @Description Verifies admin token and returns a JWT token with admin privileges
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body CredentialsRequest true "Admin credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 401 {object} CredentialsNotFoundError "Invalid admin token or not recognized"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /auth/admin/login [post]
func (a *API) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CredentialsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(ctx, w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	token, err := a.iam.LoginAdmin(ctx, iam.Credentials(req))
	switch {
	case errors.Is(err, iam.ErrUserNotFound):
		writeError(ctx, w, ErrCredentialsNotFound.WithDetails("Invalid admin credentials"), http.StatusUnauthorized)
		return
	case err != nil:
		writeError(ctx, w, ErrServerError.WithDetails("Failed to login as admin"), http.StatusInternalServerError)
		return
	}

	a.writeJSON(ctx, w, TokenResponse{Token: token}, http.StatusOK)
}

// DeleteCredentials godoc
// @Summary Delete user credentials
// @Description Deletes credentials for a user
// @Tags authentication
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Success 204 "No content"
// @Failure 400 {object} InvalidUUIDError "Invalid UUID format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} CredentialsNotFoundError "User credentials not found"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /auth/credentials/{id} [delete]
func (a *API) DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, ErrInvalidUUID, http.StatusBadRequest)
		return
	}

	err = a.iam.DropCredentials(ctx, userID)
	switch {
	case errors.Is(err, iam.ErrUserNotFound):
		writeError(ctx, w, ErrCredentialsNotFound, http.StatusNotFound)
		return
	case err != nil:
		writeError(
			ctx,
			w,
			ErrServerError.WithDetails("Failed to delete credentials"),
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCredentials godoc
// @Summary Get user credentials
// @Description Retrieves credentials for a user
// @Tags authentication
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Success 200 {object} CredentialsRequest
// @Failure 400 {object} InvalidUUIDError "Invalid UUID format"
// @Failure 401 {object} UnauthorizedError "Unauthorized"
// @Failure 403 {object} ForbiddenError "Forbidden - admin role required"
// @Failure 404 {object} CredentialsNotFoundError "User not found or does not exist"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /auth/credentials/{id} [get]
func (a *API) GetCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		writeError(ctx, w, ErrInvalidUUID, http.StatusBadRequest)
		return
	}

	creds, err := a.iam.Credentials(ctx, userID)
	switch {
	case errors.Is(err, iam.ErrUserNotFound):
		writeError(ctx, w, ErrCredentialsNotFound, http.StatusNotFound)
		return
	case err != nil:
		writeError(ctx, w, ErrServerError, http.StatusInternalServerError)
		return
	}

	a.writeJSON(ctx, w, CredentialsRequest{
		Username: creds.Username,
		Password: creds.Password,
	}, http.StatusOK)
}

// ValidateToken godoc
// @Summary Validate JWT token
// @Description Validates a JWT token and returns the identity information
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} IdentityResponse
// @Failure 401 {object} InvalidTokenError "Invalid token"
// @Failure 500 {object} ServerError "Internal server error"
// @Router /auth/validate [get]
func (a *API) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		writeError(ctx, w, InvalidTokenError{
			Code:      "INVALID_TOKEN",
			Message:   "Missing or invalid Authorization header",
			RuMessage: "Отсутствует или некорректный заголовок авторизации",
		}, http.StatusUnauthorized)
		return
	}

	tokenString := authHeader[7:]
	identity, err := a.iam.ImWatermelon(ctx, tokenString)
	switch {
	case errors.Is(err, iam.ErrInvalidToken):
		writeError(ctx, w, ErrInvalidToken, http.StatusUnauthorized)
		return
	case errors.Is(err, iam.ErrUserNotFound):
		writeError(ctx, w, ErrUnauthorized, http.StatusUnauthorized)
		return
	case err != nil:
		writeError(ctx, w, ErrServerError, http.StatusInternalServerError)
		return
	}

	a.writeJSON(ctx, w, IdentityResponse{
		ID:   identity.AuthID,
		Role: string(identity.Role),
	}, http.StatusOK)
}
