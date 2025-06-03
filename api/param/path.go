package param

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

func PathUUID(r *http.Request, name string) (uuid.UUID, error) {
	if name == "" {
		panic("empty string cannot be a path value key")
	}

	ids := r.PathValue(name)
	if ids == "" {
		return uuid.UUID{}, ErrInvalid(fmt.Sprintf("Path: %s", name))
	}

	id, err := uuid.FromString(ids)
	if err != nil {
		return id, ErrInvalid(fmt.Sprintf("Path: %s", name))
	}

	return id, nil
}
