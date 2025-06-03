package param

import (
	"fmt"
)

type InvalidParamError struct {
	ParamName string
}

func (e *InvalidParamError) Error() string {
	return fmt.Sprintf("param %q is invalid or was not provided", e.ParamName)
}

func ErrInvalid(name string) *InvalidParamError {
	return &InvalidParamError{
		ParamName: name,
	}
}
