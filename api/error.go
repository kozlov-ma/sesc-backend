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

// ToError converts any specific error type to the generic Error type
func ToError[T SpecificError](specificError T) Error {
	var baseError Error

	// Handle different specific error types
	switch e := any(specificError).(type) {
	case Error:
		baseError = e
	case InvalidRequestError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidUUIDError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidAuthHeaderError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidTokenError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case AuthError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case UnauthorizedError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case ForbiddenError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidCredentialsError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case UserNotFoundError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case UserExistsError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case CredentialsNotFoundError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case ServerError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidRoleError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidNameError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case DepartmentExistsError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidDepartmentIDError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case InvalidDepartmentError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case DepartmentNotFoundError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	case CannotRemoveDepartmentError:
		baseError = Error{
			Code:      e.Code,
			Message:   e.Message,
			RuMessage: e.RuMessage,
			Details:   e.Details,
		}
	default:
		// Fallback to default error
		baseError = Error{
			Code:      "SERVER_ERROR",
			Message:   "Internal server error",
			RuMessage: "Внутренняя ошибка сервера",
			Details:   fmt.Sprintf("%v", specificError),
		}
	}

	// Set default status code if not set
	if baseError.StatusCode == 0 {
		baseError.StatusCode = http.StatusInternalServerError
	}

	return baseError
}

// Define specific error types for each error scenario

// InvalidRequestError represents an invalid request error
type InvalidRequestError struct {
	Code      string `json:"code"             example:"INVALID_REQUEST"`
	Message   string `json:"message"          example:"Invalid request body"`
	RuMessage string `json:"ruMessage"        example:"Некорректный формат запроса"`
	Details   string `json:"details,omitzero" example:"field X is required"`
}

// WithDetails adds detail information to the error
func (e InvalidRequestError) WithDetails(details string) InvalidRequestError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidRequestError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// InvalidUUIDError represents an invalid UUID format error
type InvalidUUIDError struct {
	Code      string `json:"code"             example:"INVALID_UUID"`
	Message   string `json:"message"          example:"Invalid UUID format"`
	RuMessage string `json:"ruMessage"        example:"Некорректный формат UUID"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e InvalidUUIDError) WithDetails(details string) InvalidUUIDError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidUUIDError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// InvalidAuthHeaderError represents an invalid authentication header error
type InvalidAuthHeaderError struct {
	Code      string `json:"code"             example:"INVALID_AUTH_HEADER"`
	Message   string `json:"message"          example:"Invalid Authorization header format"`
	RuMessage string `json:"ruMessage"        example:"Неверный формат заголовка авторизации"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e InvalidAuthHeaderError) WithDetails(details string) InvalidAuthHeaderError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidAuthHeaderError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// InvalidTokenError represents an invalid or expired token error
type InvalidTokenError struct {
	Code      string `json:"code"             example:"INVALID_TOKEN"`
	Message   string `json:"message"          example:"Invalid or expired token"`
	RuMessage string `json:"ruMessage"        example:"Недействительный или просроченный токен"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e InvalidTokenError) WithDetails(details string) InvalidTokenError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidTokenError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// AuthError represents an authentication processing error
type AuthError struct {
	Code      string `json:"code"             example:"AUTH_ERROR"`
	Message   string `json:"message"          example:"Error processing authentication"`
	RuMessage string `json:"ruMessage"        example:"Ошибка обработки аутентификации"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e AuthError) WithDetails(details string) AuthError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e AuthError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// UnauthorizedError represents an unauthorized access error
type UnauthorizedError struct {
	Code      string `json:"code"             example:"UNAUTHORIZED"`
	Message   string `json:"message"          example:"Unauthorized access"`
	RuMessage string `json:"ruMessage"        example:"Неавторизованный доступ"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e UnauthorizedError) WithDetails(details string) UnauthorizedError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e UnauthorizedError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// ForbiddenError represents a forbidden access error
type ForbiddenError struct {
	Code      string `json:"code"             example:"FORBIDDEN"`
	Message   string `json:"message"          example:"Forbidden - insufficient permissions"`
	RuMessage string `json:"ruMessage"        example:"Доступ запрещен - недостаточно прав"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e ForbiddenError) WithDetails(details string) ForbiddenError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e ForbiddenError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// InvalidCredentialsError represents invalid credentials format error
type InvalidCredentialsError struct {
	Code      string `json:"code"             example:"INVALID_CREDENTIALS"`
	Message   string `json:"message"          example:"Invalid credentials format"`
	RuMessage string `json:"ruMessage"        example:"Неверный формат учетных данных"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e InvalidCredentialsError) WithDetails(details string) InvalidCredentialsError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidCredentialsError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// UserNotFoundError represents a user not found error
type UserNotFoundError struct {
	Code      string `json:"code"             example:"USER_NOT_FOUND"`
	Message   string `json:"message"          example:"User does not exist"`
	RuMessage string `json:"ruMessage"        example:"Пользователь не существует"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e UserNotFoundError) WithDetails(details string) UserNotFoundError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e UserNotFoundError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// UserExistsError represents a user already exists error
type UserExistsError struct {
	Code      string `json:"code"             example:"USER_EXISTS"`
	Message   string `json:"message"          example:"User with this username already exists"`
	RuMessage string `json:"ruMessage"        example:"Пользователь с таким именем уже существует"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e UserExistsError) WithDetails(details string) UserExistsError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e UserExistsError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// CredentialsNotFoundError represents user credentials not found error
type CredentialsNotFoundError struct {
	Code      string `json:"code"             example:"CREDENTIALS_NOT_FOUND"`
	Message   string `json:"message"          example:"User credentials not found"`
	RuMessage string `json:"ruMessage"        example:"Учетные данные пользователя не найдены"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e CredentialsNotFoundError) WithDetails(details string) CredentialsNotFoundError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e CredentialsNotFoundError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// ServerError represents an internal server error
type ServerError struct {
	Code      string `json:"code"             example:"SERVER_ERROR"`
	Message   string `json:"message"          example:"Internal server error"`
	RuMessage string `json:"ruMessage"        example:"Внутренняя ошибка сервера"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e ServerError) WithDetails(details string) ServerError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e ServerError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// InvalidRoleError represents an invalid role error
type InvalidRoleError struct {
	Code      string `json:"code"             example:"INVALID_ROLE"`
	Message   string `json:"message"          example:"Invalid role ID specified"`
	RuMessage string `json:"ruMessage"        example:"Указана некорректная роль"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e InvalidRoleError) WithDetails(details string) InvalidRoleError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidRoleError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// InvalidNameError represents an invalid name error
type InvalidNameError struct {
	Code      string `json:"code"             example:"INVALID_NAME"`
	Message   string `json:"message"          example:"Invalid name specified"`
	RuMessage string `json:"ruMessage"        example:"Указано некорректное имя"`
	Details   string `json:"details,omitzero"`
}

// WithDetails adds detail information to the error
func (e InvalidNameError) WithDetails(details string) InvalidNameError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e InvalidNameError) WithStatus(statusCode int) Error {
	return ToError(e).WithStatus(statusCode)
}

// The DepartmentExistsError is already declared in departments.go

var (
	ErrValidation = InvalidRequestError{
		Code:      "VALIDATION_ERROR",
		Message:   "Validation failed",
		RuMessage: "Ошибка валидации",
	}
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
	if errors.Is(err, sesc.ErrInvalidRole) {
		return ToError(InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, sesc.ErrUserNotFound) {
		return ToError(ErrUserNotFound).WithStatus(http.StatusNotFound)
	}
	if errors.Is(err, sesc.ErrCannotRemoveDepartment) {
		return ToError(ErrCannotRemoveDepartment).WithStatus(http.StatusConflict)
	}
	if errors.Is(err, sesc.ErrInvalidDepartment) {
		return ToError(ErrInvalidDepartment).WithStatus(http.StatusConflict)
	}
	if errors.Is(err, sesc.ErrInvalidPermission) {
		return ToError(ErrForbidden).WithStatus(http.StatusForbidden)
	}
	if errors.Is(err, sesc.ErrInvalidRoleChange) {
		return ToError(InvalidRoleError{
			Code:      "INVALID_ROLE_CHANGE",
			Message:   "Invalid role change",
			RuMessage: "Недопустимое изменение роли",
		}).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, sesc.ErrInvalidUserName) {
		return ToError(InvalidNameError{
			Code:      "INVALID_NAME",
			Message:   "Invalid or missing user name",
			RuMessage: "Указано некорректное или отсутствует имя пользователя",
		}).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, sesc.ErrInvalidDepartmentName) {
		return ToError(InvalidNameError{
			Code:      "INVALID_NAME",
			Message:   "Invalid or missing department name",
			RuMessage: "Указано некорректное или отсутствует название кафедры",
		}).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, sesc.ErrEmptyDepartment) {
		return ToError(ErrInvalidDepartment.WithDetails("Department is empty")).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, sesc.ErrDepartmentNotFound) {
		return ToError(ErrDepartmentNotFound).WithStatus(http.StatusNotFound)
	}
	if errors.Is(err, sesc.ErrInvalidUserID) {
		return ToError(ErrInvalidUUID.WithDetails("Invalid user ID")).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, sesc.ErrInvalidDepartmentID) {
		return ToError(ErrInvalidDepartmentID).WithStatus(http.StatusBadRequest)
	}

	return ToError(ErrServerError.WithDetails(err.Error())).WithStatus(http.StatusInternalServerError)
}

// Convert IAM domain errors to API errors
func iamError(err error) Error {
	if errors.Is(err, iam.ErrInvalidCredentials) {
		return ToError(ErrInvalidCredentials).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrCredentialsAlreadyExist) {
		return ToError(ErrUserExists).WithStatus(http.StatusConflict)
	}
	if errors.Is(err, iam.ErrInvalidToken) {
		return ToError(ErrInvalidToken).WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrUserNotFound) {
		return ToError(ErrUserNotFound).WithStatus(http.StatusNotFound)
	}
	if errors.Is(err, iam.ErrEmptyUsername) {
		return ToError(ErrInvalidCredentials.WithDetails("Username cannot be empty")).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrEmptyPassword) {
		return ToError(ErrInvalidCredentials.WithDetails("Password cannot be empty")).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrInvalidUserID) {
		return ToError(ErrInvalidUUID.WithDetails("Invalid user ID")).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrCredentialsNotFound) {
		return ToError(ErrCredentialsNotFound).WithStatus(http.StatusNotFound)
	}
	if errors.Is(err, iam.ErrInvalidRole) {
		return ToError(InvalidRoleError{
			Code:      "INVALID_ROLE",
			Message:   "Invalid role ID specified",
			RuMessage: "Указана некорректная роль",
		}).WithStatus(http.StatusBadRequest)
	}
	if errors.Is(err, iam.ErrUnauthorized) {
		return ToError(ErrUnauthorized).WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrTokenExpired) {
		return ToError(ErrInvalidToken.WithDetails("Token has expired")).WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrInvalidTokenFormat) {
		return ToError(ErrInvalidToken.WithDetails("Invalid token format")).WithStatus(http.StatusUnauthorized)
	}
	if errors.Is(err, iam.ErrTokenSignature) {
		return ToError(ErrInvalidToken.WithDetails("Invalid token signature")).WithStatus(http.StatusUnauthorized)
	}

	return ToError(ErrServerError.WithDetails(err.Error())).WithStatus(http.StatusInternalServerError)
}
