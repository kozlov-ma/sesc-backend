package respond

import "encoding/json"

type Status[T any] struct {
	Value      T
	StatusCode int
}

func (s *Status[T]) AsHTTPStatusCode() int {
	return s.StatusCode
}

func (s *Status[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}

func WithStatus[T any](obj T, code int) *Status[T] {
	return &Status[T]{
		Value:      obj,
		StatusCode: code,
	}
}
