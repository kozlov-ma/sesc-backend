package param

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
)

func QueryInt(r *http.Request, name string) (int, error) {
	i, err := strconv.Atoi(r.FormValue(name))
	if err != nil {
		return 0, ErrInvalid(name)
	}

	return i, nil
}

func QueryString(r *http.Request, name string) (string, error) {
	v := r.FormValue(name)
	if v == "" {
		return "", ErrInvalid(name)
	}

	return v, nil
}

// QueryStringOrZero returns a named query param of type string or "".
//
// Why this exists if you can just call r.FormValue?
//
// Because I was drunk AF and spent at least 30 minutes on debugging an issue
// that turned out to be calling r.PathValue instead of r.FormValue.
func QueryStringOrZero(r *http.Request, name string) string {
	return r.FormValue(name)
}

func QueryIntOrZero(r *http.Request, name string) int {
	i, _ := strconv.Atoi(r.FormValue(name))
	return i
}

func QueryBool(r *http.Request, name string) (bool, error) {
	b, err := strconv.ParseBool(r.FormValue(name))
	if err != nil {
		return false, ErrInvalid(name)
	}
	return b, nil
}

func QueryBoolOrFalse(r *http.Request, name string) bool {
	b, _ := strconv.ParseBool(r.FormValue(name))
	return b
}

// QueryPagination parses offset and limit parameters with defaults
func QueryPagination(r *http.Request) (offset, limit int, err error) {
	offset = 0
	limit = 10

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, 0, ErrInvalid("offset")
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, ErrInvalid("limit")
		}
	}

	return offset, limit, nil
}

func QueryUUID(r *http.Request, name string) (uuid.UUID, error) {
	if name == "" {
		panic("empty string cannot be a query param key")
	}

	ids := r.FormValue(name)
	if ids == "" {
		return uuid.UUID{}, ErrInvalid(fmt.Sprintf("Query: %s", name))
	}

	id, err := uuid.FromString(ids)
	if err != nil {
		return id, ErrInvalid(fmt.Sprintf("Query: %s", name))
	}

	return id, nil
}
