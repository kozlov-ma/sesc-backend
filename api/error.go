package api

// APIError represents a structured error returned by the API.
type APIError struct {
	Code      string `json:"code"        example:"INVALID_REQUEST"`
	Message   string `json:"message"     example:"Invalid request body"`
	RuMessage string `json:"ruMessage"   example:"Некорректный формат запроса"`
	Details   string `json:"details,omitempty" example:"field X is required"`
}

// ErrInvalidRequest is used when the request payload cannot be parsed.
var ErrInvalidRequest = APIError{
	Code:      "INVALID_REQUEST",
	Message:   "Invalid request body",
	RuMessage: "Некорректный формат запроса",
}
