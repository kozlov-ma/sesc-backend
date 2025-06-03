package param

import (
	"net/http"
	"strings"
)

func BearerAuth(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")

	if !strings.HasPrefix(header, "Bearer ") {
		return "", ErrInvalid("Header: Authorization")
	}

	return header[len("Bearer "):], nil
}
