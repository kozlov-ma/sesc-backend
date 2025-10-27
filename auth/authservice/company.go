package authservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kozlov-ma/sesc-backend/auth"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/kozlov-ma/sesc-backend/company/companyservice"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type Company struct {
	company       companyservice.S
	tokenDuration time.Duration
	jwtSecret     []byte
}

func NewCompany(companysvc companyservice.S, tokenDuration time.Duration, jwtSecret []byte) *Company {
	return &Company{company: companysvc, tokenDuration: tokenDuration, jwtSecret: jwtSecret}
}

// Login verifies credentials and returns signed JWT token string.
func (c *Company) Login(ctx context.Context, id, password string) (string, error) {
	rec := event.Get(ctx).Sub("authservice/company/login")

	rec.Sub("params").Set(
		"id", id,
	)

	u, err := c.company.User(ctx, companyquery.User{
		ID:       id,
		Password: password,
	})

	if err != nil {
		return "", fmt.Errorf("failed to query a user: %w", err)
	}

	ctx = rec.Sub("generate_token").Wrap(ctx)
	token, err := c.generateUserToken(ctx, u)
	if err != nil {
		return "", fmt.Errorf("failed to generate a user auth token: %w", err)
	}

	rec.Set("success", true)
	return token, nil
}

// generateUserToken generates a JWT token for a user
func (c *Company) generateUserToken(
	ctx context.Context,
	u company.User,
) (string, error) {
	rec := event.Get(ctx).Sub("generate_token")
	rec.Set(
		"user_id", u.ID,
		"roles", strings.Join(u.RoleStrings(), "+"),
	)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   u.ID,
		"role":      strings.Join(u.RoleStrings(), "+"),
		"exp":       time.Now().Add(c.tokenDuration).Unix(),
		"user_hash": u.Hash(),
	})

	signed, err := token.SignedString(c.jwtSecret)
	if err != nil {
		err := fmt.Errorf("failed to sign an authentication token: %w", err)
		rec.Add(events.Error, err)
		rec.Set("success", false)
		return "", err
	}

	rec.Set("success", true)
	return signed, nil
}

// ImWatermelon parses tokenString, returns a company.User or an auth.ErrInvalidToken.
func (c *Company) ImWatermelon(ctx context.Context, tokenString string) (company.User, error) {
	rec := event.Get(ctx).Sub("authservice/company/im_watermelon")

	ctx = rec.Sub("parse_token").Wrap(ctx)
	claims, err := c.parseAndValidateToken(ctx, tokenString)
	if err != nil {
		return company.User{}, err
	}

	ctx = rec.Sub("extract_claims").Wrap(ctx)
	userID, role, hash, err := c.extractTokenClaims(ctx, claims)
	if err != nil {
		return company.User{}, err
	}

	u, err := c.company.User(ctx, companyquery.User{
		ID: userID,
	})
	if err != nil {
		return company.User{}, fmt.Errorf("failed to query a company user: %w", err)
	}

	if u.ID != userID || strings.Join(u.RoleStrings(), "+") != role || u.Hash() != hash {
		return company.User{}, fmt.Errorf("%w: user info changed, a re-login is required", auth.ErrInvalidToken)
	}

	rec.Set("success", true)
	return u, nil
}

func (c *Company) parseAndValidateToken(
	ctx context.Context,
	tokenString string,
) (jwt.MapClaims, error) {
	rec := event.Get(ctx).Sub("parse_token")

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, auth.ErrInvalidToken
		}
		return c.jwtSecret, nil
	})

	if err != nil || !parsed.Valid {
		rec.Add(events.Error, err)
		rec.Set("valid", false)
		return nil, errors.Join(iam.ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		rec.Set("valid", false)
		return nil, iam.ErrInvalidToken
	}

	rec.Set("valid", true)
	return claims, nil
}

// extractTokenClaims extracts and validates token claims
func (c *Company) extractTokenClaims(
	ctx context.Context,
	claims jwt.MapClaims,
) (string, string, string, error) {
	rec := event.Get(ctx).Sub("extract_claims")

	userID, ok1 := claims["user_id"].(string)
	role, ok2 := claims["role"].(string)
	hash, ok3 := claims["user_hash"].(string)
	if !ok1 || !ok2 || !ok3 {
		rec.Set(
			"valid", false,
			"user_id", userID,
			"role_str", role,
			"hash_str", hash,
		)
		return "", "", "", auth.ErrInvalidToken
	}

	rec.Set(
		"valid", true,
		"user_id", userID,
		"role_str", role,
		"hash_str", hash,
	)

	return userID, role, hash, nil
}
