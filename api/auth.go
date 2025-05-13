package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/iam"
)

type CredentialsRequest struct {
	Username string `json:"username" example:"johndoe"   validate:"required"`
	Password string `json:"password" example:"secret123" validate:"required"`
}

type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
}

type AdminLoginRequest struct {
	Token string `json:"token" example:"admin-secret-token" validate:"required"`
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
// @Failure 400 {object} APIError "Invalid credentials or request format"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 403 {object} APIError "Forbidden - admin role required"
// @Failure 404 {object} APIError "User does not exist"
// @Failure 409 {object} APIError "User already exists"
// @Failure 500 {object} APIError "Internal server error"
// @Router /users/{id}/credentials [put]
func (a *API) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get userID from path parameter
	idStr := r.PathValue("id")
	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	var credsReq CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&credsReq); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	creds := iam.Credentials{
		Username: credsReq.Username,
		Password: credsReq.Password,
	}

	authID, err := a.iam.RegisterCredentials(ctx, userID, creds)
	switch {
	case errors.Is(err, iam.ErrInvalidCredentials):
		a.writeError(w, APIError{
			Code:      "INVALID_CREDENTIALS",
			Message:   "Invalid credentials format",
			RuMessage: "Неверный формат учетных данных",
		}, http.StatusBadRequest)
		return
	case errors.Is(err, iam.ErrUserDoesNotExist):
		a.writeError(w, APIError{
			Code:      "USER_NOT_FOUND",
			Message:   "User does not exist",
			RuMessage: "Пользователь не существует",
		}, http.StatusNotFound)
		return
	case errors.Is(err, iam.ErrUserAlreadyExists):
		a.writeError(w, APIError{
			Code:      "USER_EXISTS",
			Message:   "User with this username already exists",
			RuMessage: "Пользователь с таким именем уже существует",
		}, http.StatusConflict)
		return
	case err != nil:
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to register user",
			RuMessage: "Ошибка регистрации пользователя",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, map[string]uuid.UUID{"authId": authID}, http.StatusCreated)
}

// Login godoc
// @Summary User login
// @Description Verifies user credentials and returns a JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body CredentialsRequest true "User credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} APIError "Invalid credentials format"
// @Failure 401 {object} APIError "Invalid credentials"
// @Failure 500 {object} APIError "Internal server error"
// @Router /auth/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var credsReq CredentialsRequest

	if err := json.NewDecoder(r.Body).Decode(&credsReq); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	creds := iam.Credentials{
		Username: credsReq.Username,
		Password: credsReq.Password,
	}

	token, err := a.iam.Login(ctx, creds)
	switch {
	case errors.Is(err, iam.ErrInvalidCredentials):
		a.writeError(w, APIError{
			Code:      "INVALID_CREDENTIALS",
			Message:   "Invalid username or password",
			RuMessage: "Неверное имя пользователя или пароль",
		}, http.StatusUnauthorized)
		return
	case err != nil:
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to login",
			RuMessage: "Ошибка авторизации",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, TokenResponse{Token: token}, http.StatusOK)
}

// LoginAdmin godoc
// @Summary Admin login
// @Description Verifies admin token and returns a JWT token with admin privileges
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} APIError "Invalid request format"
// @Failure 401 {object} APIError "Invalid admin token"
// @Failure 500 {object} APIError "Internal server error"
// @Router /auth/admin/login [post]
func (a *API) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req AdminLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	token, err := a.iam.LoginAdmin(ctx, req.Token)
	switch {
	case errors.Is(err, iam.ErrInvalidCredentials):
		a.writeError(w, APIError{
			Code:      "INVALID_ADMIN_TOKEN",
			Message:   "Invalid admin token",
			RuMessage: "Неверный токен администратора",
		}, http.StatusUnauthorized)
		return
	case err != nil:
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to login as admin",
			RuMessage: "Ошибка авторизации администратора",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, TokenResponse{Token: token}, http.StatusOK)
}

// DeleteCredentials godoc
// @Summary Delete user credentials
// @Description Deletes credentials for a user
// @Tags authentication
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "User UUID"
// @Success 204 "No content"
// @Failure 400 {object} APIError "Invalid UUID format"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 403 {object} APIError "Forbidden - admin role required"
// @Failure 404 {object} APIError "User not found or does not exist"
// @Failure 500 {object} APIError "Internal server error"
// @Router /auth/credentials/{id} [delete]
func (a *API) DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	err = a.iam.DropCredentials(ctx, userID)
	switch {
	case errors.Is(err, iam.ErrUserNotFound):
		a.writeError(w, APIError{
			Code:      "CREDENTIALS_NOT_FOUND",
			Message:   "User credentials not found",
			RuMessage: "Учетные данные пользователя не найдены",
		}, http.StatusNotFound)
		return
	case errors.Is(err, iam.ErrUserDoesNotExist):
		a.writeError(w, APIError{
			Code:      "USER_NOT_FOUND",
			Message:   "User does not exist",
			RuMessage: "Пользователь не существует",
		}, http.StatusNotFound)
		return
	case err != nil:
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to delete credentials",
			RuMessage: "Ошибка удаления учетных данных",
		}, http.StatusInternalServerError)
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
// @Failure 400 {object} APIError "Invalid UUID format"
// @Failure 401 {object} APIError "Unauthorized"
// @Failure 403 {object} APIError "Forbidden - admin role required"
// @Failure 404 {object} APIError "User not found or does not exist"
// @Failure 500 {object} APIError "Internal server error"
// @Router /auth/credentials/{id} [get]
func (a *API) GetCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	userID, err := uuid.FromString(idStr)
	if err != nil {
		a.writeError(w, APIError{
			Code:      "INVALID_UUID",
			Message:   "Invalid user ID format",
			RuMessage: "Некорректный формат ID пользователя",
		}, http.StatusBadRequest)
		return
	}

	creds, err := a.iam.Credentials(ctx, userID)
	switch {
	case errors.Is(err, iam.ErrUserNotFound):
		a.writeError(w, APIError{
			Code:      "CREDENTIALS_NOT_FOUND",
			Message:   "User credentials not found",
			RuMessage: "Учетные данные пользователя не найдены",
		}, http.StatusNotFound)
		return
	case errors.Is(err, iam.ErrUserDoesNotExist):
		a.writeError(w, APIError{
			Code:      "USER_NOT_FOUND",
			Message:   "User does not exist",
			RuMessage: "Пользователь не существует",
		}, http.StatusNotFound)
		return
	case err != nil:
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to get user credentials",
			RuMessage: "Ошибка получения учетных данных пользователя",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, CredentialsRequest{
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
// @Failure 401 {object} APIError "Invalid token"
// @Failure 500 {object} APIError "Internal server error"
// @Router /auth/validate [get]
func (a *API) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		a.writeError(w, APIError{
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
		a.writeError(w, APIError{
			Code:      "INVALID_TOKEN",
			Message:   "Invalid or expired token",
			RuMessage: "Недействительный или просроченный токен",
		}, http.StatusUnauthorized)
		return
	case err != nil:
		a.writeError(w, APIError{
			Code:      "SERVER_ERROR",
			Message:   "Failed to validate token",
			RuMessage: "Ошибка проверки токена",
		}, http.StatusInternalServerError)
		return
	}

	a.writeJSON(w, IdentityResponse{
		ID:   identity.ID,
		Role: string(identity.Role),
	}, http.StatusOK)
}
