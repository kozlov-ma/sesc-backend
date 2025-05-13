package api

// Error represents a structured error returned by the API.
type Error struct {
	Code      string `json:"code"             example:"INVALID_REQUEST"             validate:"required"`
	Message   string `json:"message"          example:"Invalid request body"        validate:"required"`
	RuMessage string `json:"ruMessage"        example:"Некорректный формат запроса" validate:"required"`
	Details   string `json:"details,omitzero" example:"field X is required"`
}

// WithDetails adds detail information to the error
func (e Error) WithDetails(details string) Error {
	e.Details = details
	return e
}

// Common errors used across the API
var (
	// Request validation errors
	ErrInvalidRequest = Error{
		Code:      "INVALID_REQUEST",
		Message:   "Invalid request body",
		RuMessage: "Некорректный формат запроса",
	}

	ErrInvalidUUID = Error{
		Code:      "INVALID_UUID",
		Message:   "Invalid UUID format",
		RuMessage: "Некорректный формат UUID",
	}

	// Authentication errors
	ErrInvalidAuthHeader = Error{
		Code:      "INVALID_AUTH_HEADER",
		Message:   "Invalid Authorization header format",
		RuMessage: "Неверный формат заголовка авторизации",
	}

	ErrInvalidToken = Error{
		Code:      "INVALID_TOKEN",
		Message:   "Invalid or expired token",
		RuMessage: "Недействительный или просроченный токен",
	}

	ErrAuthError = Error{
		Code:      "AUTH_ERROR",
		Message:   "Error processing authentication",
		RuMessage: "Ошибка обработки аутентификации",
	}

	ErrUnauthorized = Error{
		Code:      "UNAUTHORIZED",
		Message:   "Unauthorized access",
		RuMessage: "Неавторизованный доступ",
	}

	ErrForbidden = Error{
		Code:      "FORBIDDEN",
		Message:   "Forbidden - insufficient permissions",
		RuMessage: "Доступ запрещен - недостаточно прав",
	}

	ErrInvalidCredentials = Error{
		Code:      "INVALID_CREDENTIALS",
		Message:   "Invalid credentials format",
		RuMessage: "Неверный формат учетных данных",
	}

	// User errors
	ErrUserNotFound = Error{
		Code:      "USER_NOT_FOUND",
		Message:   "User does not exist",
		RuMessage: "Пользователь не существует",
	}

	ErrUserExists = Error{
		Code:      "USER_EXISTS",
		Message:   "User with this username already exists",
		RuMessage: "Пользователь с таким именем уже существует",
	}

	ErrCredentialsNotFound = Error{
		Code:      "CREDENTIALS_NOT_FOUND",
		Message:   "User credentials not found",
		RuMessage: "Учетные данные пользователя не найдены",
	}

	// Server errors
	ErrServerError = Error{
		Code:      "SERVER_ERROR",
		Message:   "Internal server error",
		RuMessage: "Внутренняя ошибка сервера",
	}
)
