package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// Client is the HTTP client for API testing
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// SetToken sets the authorization token for subsequent requests
func (c *Client) SetToken(token string) {
	c.token = token
}

// makeRequest is a helper method to create and send HTTP requests
func (c *Client) makeRequest(
	ctx context.Context,
	method, endpoint string,
	body any,
	//nolint:unparam // this can be needed in the future.
	query url.Values,
) (*http.Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, endpoint)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// parseResponse is a helper method to parse API responses
func parseResponse(resp *http.Response, out any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiError Error
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("api error: %s (code: %s, status: %d)", apiError.Message, apiError.Code, resp.StatusCode)
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Login performs a login request and returns the token
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	resp, err := c.makeRequest(ctx, http.MethodPost, "/auth/login", LoginRequest{
		Username: username,
		Password: password,
	}, nil)
	if err != nil {
		return "", err
	}

	var loginResp LoginResponse
	if err := parseResponse(resp, &loginResp); err != nil {
		return "", err
	}

	c.token = loginResp.Token
	return loginResp.Token, nil
}

// LoginAdmin performs an admin login request and returns the token
func (c *Client) LoginAdmin(ctx context.Context, username, password string) (string, error) {
	resp, err := c.makeRequest(ctx, http.MethodPost, "/auth/admin/login", LoginRequest{
		Username: username,
		Password: password,
	}, nil)
	if err != nil {
		return "", err
	}

	var loginResp LoginResponse
	if err := parseResponse(resp, &loginResp); err != nil {
		return "", err
	}

	c.token = loginResp.Token
	return loginResp.Token, nil
}

// ValidateToken validates the current token
func (c *Client) ValidateToken(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/auth/validate", nil, nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// GetCurrentUser gets the current user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/users/me", nil, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers gets all users
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/users", nil, nil)
	if err != nil {
		return nil, err
	}

	var usersResp struct {
		Users []User `json:"users"`
	}
	if err := parseResponse(resp, &usersResp); err != nil {
		return nil, err
	}
	return usersResp.Users, nil
}

// GetUser gets a user by ID
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/users/"+id, nil, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	resp, err := c.makeRequest(ctx, http.MethodPost, "/users", req, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// PatchUser updates a user
func (c *Client) PatchUser(ctx context.Context, id string, req PatchUserRequest) (*User, error) {
	resp, err := c.makeRequest(ctx, http.MethodPatch, "/users/"+id, req, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// RegisterUser sets credentials for a user
func (c *Client) RegisterUser(ctx context.Context, userID string, req RegisterUserRequest) error {
	resp, err := c.makeRequest(ctx, http.MethodPut, "/users/"+userID+"/credentials", req, nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// GetDepartments gets all departments
func (c *Client) GetDepartments(ctx context.Context) ([]Department, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/departments", nil, nil)
	if err != nil {
		return nil, err
	}

	var departmentsResp struct {
		Departments []Department `json:"departments"`
	}
	if err := parseResponse(resp, &departmentsResp); err != nil {
		return nil, err
	}
	return departmentsResp.Departments, nil
}

// CreateDepartment creates a new department
func (c *Client) CreateDepartment(ctx context.Context, req CreateDepartmentRequest) (*Department, error) {
	resp, err := c.makeRequest(ctx, http.MethodPost, "/departments", req, nil)
	if err != nil {
		return nil, err
	}

	var department Department
	if err := parseResponse(resp, &department); err != nil {
		return nil, err
	}
	return &department, nil
}

// UpdateDepartment updates a department
func (c *Client) UpdateDepartment(ctx context.Context, id string, req UpdateDepartmentRequest) (*Department, error) {
	resp, err := c.makeRequest(ctx, http.MethodPut, "/departments/"+id, req, nil)
	if err != nil {
		return nil, err
	}

	var department Department
	if err := parseResponse(resp, &department); err != nil {
		return nil, err
	}
	return &department, nil
}

// DeleteDepartment deletes a department
func (c *Client) DeleteDepartment(ctx context.Context, id string) error {
	resp, err := c.makeRequest(ctx, http.MethodDelete, "/departments/"+id, nil, nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// GetRoles gets all roles
func (c *Client) GetRoles(ctx context.Context) ([]Role, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/roles", nil, nil)
	if err != nil {
		return nil, err
	}

	var rolesResp struct {
		Roles []Role `json:"roles"`
	}
	if err := parseResponse(resp, &rolesResp); err != nil {
		return nil, err
	}
	return rolesResp.Roles, nil
}

// GetPermissions gets all permissions
func (c *Client) GetPermissions(ctx context.Context) ([]Permission, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/permissions", nil, nil)
	if err != nil {
		return nil, err
	}

	var permissionsResp struct {
		Permissions []Permission `json:"permissions"`
	}
	if err := parseResponse(resp, &permissionsResp); err != nil {
		return nil, err
	}
	return permissionsResp.Permissions, nil
}
