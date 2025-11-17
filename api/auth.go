package api

import (
	"encoding/json"
	"net/http"

	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type CredentialsRequest struct {
	Username string `json:"username" example:"johndoe"   validate:"required"`
	Password string `json:"password" example:"secret123" validate:"required"`
}

type TokenResponse struct {
	Token string        `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
	User  *respond.User `json:"user"                                                    validate:"required"`
}

// Login godoc
// @Summary User login
// @Description Verifies user credentials and returns a JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body CredentialsRequest true "User credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} respond.Error "Invalid request format"
// @Failure 401 {object} respond.Error "Invalid credentials or user does not exist"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /auth/login [post]
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	var credsReq CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&credsReq); err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	token, err := a.iam.Login(ctx, credsReq.Username, credsReq.Password)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	user, err := a.iam.ImWatermelon(ctx, token)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	a.writeJSON(ctx, w, TokenResponse{Token: token, User: respond.WithUser(user)})
}

// ValidateToken godoc
// @Summary Validate JWT token
// @Description Validates a JWT token and returns the identity information
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Param Authorization header string false "Bearer JWT token"
// @Success 200 {object} respond.User
// @Failure 401 {object} respond.Error "Invalid token"
// @Failure 500 {object} respond.Error "Internal server error"
// @Router /auth/validate [get]
func (a *API) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u := CurrentUser(ctx)

	a.writeJSON(ctx, w, respond.WithUser(u))
}
