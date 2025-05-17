package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// Error represents a structured error returned by the API.
type Error struct {
	Code       string `json:"code"             example:"INVALID_REQUEST"             validate:"required"`
	Message    string `json:"message"          example:"Invalid request body"        validate:"required"`
	RuMessage  string `json:"ruMessage"        example:"Некорректный формат запроса" validate:"required"`
	Details    string `json:"details,omitzero" example:"field X is required"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e Error) WithDetails(details string) Error {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e Error) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return e
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s (%s); %s", e.Code, e.Message, e.RuMessage, e.Details)
}

type SpecificError interface {
	InvalidRequestError | InvalidUUIDError | InvalidAuthHeaderError |
		InvalidTokenError | AuthError | UnauthorizedError |
		ForbiddenError | InvalidCredentialsError | UserNotFoundError |
		UserExistsError | CredentialsNotFoundError | ServerError |
		InvalidRoleError | InvalidNameError | DepartmentExistsError |
		InvalidDepartmentIDError | InvalidDepartmentError | DepartmentNotFoundError |
		CannotRemoveDepartmentError | Error
}

// InvalidRequestError represents an invalid request error
type InvalidRequestError struct {
	Code       string `json:"code"             example:"INVALID_REQUEST"`
	Message    string `json:"message"          example:"Invalid request body"`
	RuMessage  string `json:"ruMessage"        example:"Некорректный формат запроса"`
	Details    string `json:"details,omitzero" example:"field X is required"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidRequestError) WithDetails(details string) InvalidRequestError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidRequestError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// InvalidUUIDError represents an invalid UUID format error
type InvalidUUIDError struct {
	Code       string `json:"code"             example:"INVALID_UUID"`
	Message    string `json:"message"          example:"Invalid UUID format"`
	RuMessage  string `json:"ruMessage"        example:"Некорректный формат UUID"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidUUIDError) WithDetails(details string) InvalidUUIDError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidUUIDError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// InvalidAuthHeaderError represents an invalid authentication header error
type InvalidAuthHeaderError struct {
	Code       string `json:"code"             example:"INVALID_AUTH_HEADER"`
	Message    string `json:"message"          example:"Invalid Authorization header format"`
	RuMessage  string `json:"ruMessage"        example:"Неверный формат заголовка авторизации"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidAuthHeaderError) WithDetails(details string) InvalidAuthHeaderError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidAuthHeaderError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// InvalidTokenError represents an invalid or expired token error
type InvalidTokenError struct {
	Code       string `json:"code"             example:"INVALID_TOKEN"`
	Message    string `json:"message"          example:"Invalid or expired token"`
	RuMessage  string `json:"ruMessage"        example:"Недействительный или просроченный токен"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidTokenError) WithDetails(details string) InvalidTokenError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidTokenError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// AuthError represents an authentication processing error
type AuthError struct {
	Code       string `json:"code"             example:"AUTH_ERROR"`
	Message    string `json:"message"          example:"Error processing authentication"`
	RuMessage  string `json:"ruMessage"        example:"Ошибка обработки аутентификации"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e AuthError) WithDetails(details string) AuthError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e AuthError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// UnauthorizedError represents an unauthorized access error
type UnauthorizedError struct {
	Code       string `json:"code"             example:"UNAUTHORIZED"`
	Message    string `json:"message"          example:"Unauthorized access"`
	RuMessage  string `json:"ruMessage"        example:"Неавторизованный доступ"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e UnauthorizedError) WithDetails(details string) UnauthorizedError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e UnauthorizedError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// ForbiddenError represents a forbidden access error
type ForbiddenError struct {
	Code       string `json:"code"             example:"FORBIDDEN"`
	Message    string `json:"message"          example:"Forbidden - insufficient permissions"`
	RuMessage  string `json:"ruMessage"        example:"Доступ запрещен - недостаточно прав"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e ForbiddenError) WithDetails(details string) ForbiddenError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e ForbiddenError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// InvalidCredentialsError represents invalid credentials format error
type InvalidCredentialsError struct {
	Code       string `json:"code"             example:"INVALID_CREDENTIALS"`
	Message    string `json:"message"          example:"Invalid credentials format"`
	RuMessage  string `json:"ruMessage"        example:"Неверный формат учетных данных"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidCredentialsError) WithDetails(details string) InvalidCredentialsError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidCredentialsError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// UserNotFoundError represents a user not found error
type UserNotFoundError struct {
	Code       string `json:"code"             example:"USER_NOT_FOUND"`
	Message    string `json:"message"          example:"User does not exist"`
	RuMessage  string `json:"ruMessage"        example:"Пользователь не существует"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e UserNotFoundError) WithDetails(details string) UserNotFoundError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e UserNotFoundError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// UserExistsError represents a user already exists error
type UserExistsError struct {
	Code       string `json:"code"             example:"USER_EXISTS"`
	Message    string `json:"message"          example:"User with this username already exists"`
	RuMessage  string `json:"ruMessage"        example:"Пользователь с таким именем уже существует"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e UserExistsError) WithDetails(details string) UserExistsError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e UserExistsError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// CredentialsNotFoundError represents user credentials not found error
type CredentialsNotFoundError struct {
	Code       string `json:"code"             example:"CREDENTIALS_NOT_FOUND"`
	Message    string `json:"message"          example:"User credentials not found"`
	RuMessage  string `json:"ruMessage"        example:"Учетные данные пользователя не найдены"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e CredentialsNotFoundError) WithDetails(details string) CredentialsNotFoundError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e CredentialsNotFoundError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// ServerError represents an internal server error
type ServerError struct {
	Code       string `json:"code"             example:"SERVER_ERROR"`
	Message    string `json:"message"          example:"Internal server error"`
	RuMessage  string `json:"ruMessage"        example:"Внутренняя ошибка сервера"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e ServerError) WithDetails(details string) ServerError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e ServerError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// InvalidRoleError represents an invalid role error
type InvalidRoleError struct {
	Code       string `json:"code"             example:"INVALID_ROLE"`
	Message    string `json:"message"          example:"Invalid role ID specified"`
	RuMessage  string `json:"ruMessage"        example:"Указана некорректная роль"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidRoleError) WithDetails(details string) InvalidRoleError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidRoleError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// InvalidNameError represents an invalid name error
type InvalidNameError struct {
	Code       string `json:"code"             example:"INVALID_NAME"`
	Message    string `json:"message"          example:"Invalid name specified"`
	RuMessage  string `json:"ruMessage"        example:"Указано некорректное имя"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e InvalidNameError) WithDetails(details string) InvalidNameError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidNameError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// The DepartmentExistsError is already declared in departments.go

var (
	ErrInvalidRequest = InvalidRequestError{
		Code:      "INVALID_REQUEST",
		Message:   "Invalid request body",
		RuMessage: "Некорректный формат запроса",
	}

	ErrInvalidUUID = InvalidUUIDError{
		Code:      "INVALID_UUID",
		Message:   "Invalid UUID format",
		RuMessage: "Некорректный формат UUID",
	}

	ErrInvalidAuthHeader = InvalidAuthHeaderError{
		Code:      "INVALID_AUTH_HEADER",
		Message:   "Invalid Authorization header format",
		RuMessage: "Неверный формат заголовка авторизации",
	}

	ErrInvalidToken = InvalidTokenError{
		Code:      "INVALID_TOKEN",
		Message:   "Invalid or expired token",
		RuMessage: "Недействительный или просроченный токен",
	}

	ErrAuthError = AuthError{
		Code:      "AUTH_ERROR",
		Message:   "Error processing authentication",
		RuMessage: "Ошибка обработки аутентификации",
	}

	ErrUnauthorized = UnauthorizedError{
		Code:      "UNAUTHORIZED",
		Message:   "Unauthorized access",
		RuMessage: "Неавторизованный доступ",
	}

	ErrForbidden = ForbiddenError{
		Code:      "FORBIDDEN",
		Message:   "Forbidden - insufficient permissions",
		RuMessage: "Доступ запрещен - недостаточно прав",
	}

	ErrInvalidCredentials = InvalidCredentialsError{
		Code:      "INVALID_CREDENTIALS",
		Message:   "Invalid credentials format",
		RuMessage: "Неверный формат учетных данных",
	}

	ErrUserNotFound = UserNotFoundError{
		Code:      "USER_NOT_FOUND",
		Message:   "User does not exist",
		RuMessage: "Пользователь не существует",
	}

	ErrUserExists = UserExistsError{
		Code:      "USER_EXISTS",
		Message:   "User with this credentials already exists",
		RuMessage: "Пользователь с такими учетными данными уже существует",
	}

	ErrCredentialsNotFound = CredentialsNotFoundError{
		Code:      "CREDENTIALS_NOT_FOUND",
		Message:   "User credentials not found",
		RuMessage: "Учетные данные пользователя не найдены",
	}

	ErrServerError = ServerError{
		Code:      "SERVER_ERROR",
		Message:   "Internal server error",
		RuMessage: "Внутренняя ошибка сервера",
	}
)

// Convert SESC domain errors to API errors
func sescError(err error) Error {
	switch {
	case errors.Is(err, sesc.ErrInvalidRole):
		return InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}.WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrUserNotFound):
		return ErrUserNotFound.WithStatus(http.StatusNotFound)
	case errors.Is(err, sesc.ErrCannotRemoveDepartment):
		return ErrCannotRemoveDepartment.WithStatus(http.StatusConflict)
	case errors.Is(err, sesc.ErrInvalidDepartment):
		return ErrInvalidDepartment.WithStatus(http.StatusConflict)
	case errors.Is(err, sesc.ErrInvalidPermission):
		return ErrForbidden.WithStatus(http.StatusForbidden)
	case errors.Is(err, sesc.ErrInvalidRoleChange):
		return InvalidRoleError{
			Code:      "INVALID_ROLE_CHANGE",
			Message:   "Invalid role change",
			RuMessage: "Недопустимое изменение роли",
		}.WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrInvalidUserName):
		return InvalidNameError{
			Code:      "INVALID_NAME",
			Message:   "Invalid or missing user name",
			RuMessage: "Указано некорректное или отсутствует имя пользователя",
		}.WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrInvalidDepartmentName):
		return InvalidNameError{
			Code:      "INVALID_NAME",
			Message:   "Invalid or missing department name",
			RuMessage: "Указано некорректное или отсутствует название кафедры",
		}.WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrEmptyDepartment):
		return ErrInvalidDepartment.WithDetails("Department is empty").WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrDepartmentNotFound):
		return ErrDepartmentNotFound.WithStatus(http.StatusNotFound)
	case errors.Is(err, sesc.ErrInvalidUserID):
		return ErrInvalidUUID.WithDetails("Invalid user ID").WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrInvalidDepartmentID):
		return ErrInvalidDepartmentID.WithStatus(http.StatusBadRequest)
	default:
		return ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError)
	}
}

// Convert IAM domain errors to API errors
func iamError(err error) Error {
	if errors.Is(err, iam.ErrInvalidCredentials) {
		return ErrInvalidCredentials.WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrCredentialsAlreadyExist) {
		return ErrUserExists.WithStatus(http.StatusConflict)
	}
	if errors.Is(err, iam.ErrInvalidToken) {
		return ErrInvalidToken.WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrUserNotFound) {
		return ErrUserNotFound.WithStatus(http.StatusNotFound)
	}
	if errors.Is(err, iam.ErrEmptyUsername) {
		return ErrInvalidCredentials.WithDetails("Username cannot be empty").WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrEmptyPassword) {
		return ErrInvalidCredentials.WithDetails("Password cannot be empty").WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrInvalidUserID) {
		return ErrInvalidUUID.WithDetails("Invalid user ID").WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrCredentialsNotFound) {
		return ErrCredentialsNotFound.WithStatus(http.StatusNotFound)
	}
	if errors.Is(err, iam.ErrInvalidRole) {
		return InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}.WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrUnauthorized) {
		return ErrUnauthorized.WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrTokenExpired) {
		return ErrInvalidToken.WithDetails("Token has expired").WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrInvalidTokenFormat) {
		return ErrInvalidToken.WithDetails("Invalid token format").WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrTokenSignature) {
		return ErrInvalidToken.WithDetails("Invalid token signature").WithStatus(http.StatusUnauthorized)
	}

	return ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError)
}
