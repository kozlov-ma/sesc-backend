package api

// APIError represents a structured error returned by the API.
type APIError struct {
	Code      string `json:"code"             example:"INVALID_REQUEST"             validate:"required"`
	Message   string `json:"message"          example:"Invalid request body"        validate:"required"`
	RuMessage string `json:"ruMessage"        example:"Некорректный формат запроса" validate:"required"`
	Details   string `json:"details,omitzero" example:"field X is required"`
}

// ErrInvalidRequest is used when the request payload cannot be parsed.
var ErrInvalidRequest = APIError{
	Code:      "INVALID_REQUEST",
	Message:   "Invalid request body",
	RuMessage: "Некорректный формат запроса",
}
