package param

import (
	"net/http"
	"strconv"
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

func QueryIntOrZero(r *http.Request, name string) int {
	i, _ := strconv.Atoi(r.FormValue(name))
	return i
}
