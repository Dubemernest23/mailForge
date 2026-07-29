package httpx

import (
	"encoding/json"
	"mailForgeApi/internal/shared/constants"
	"mailForgeApi/internal/shared/response"
	"net/http"
)

// decodeAndValidate is shared across all four handlers — decode JSON body,
// then run the DTO's own Validate(). One place for this pattern instead of
// repeating decode+validate boilerplate four times.

type Validatable interface {
	Validate() error
}

func DecodeAndValidate[T Validatable](r *http.Request, dst *T) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		return response.NewAppError(constants.StatusBadRequest, response.CodeBadRequest, "invalid request body")
	}

	err = (*dst).Validate()
	if err != nil {
		return response.NewAppError(constants.StatusBadRequest, response.CodeBadRequest, err.Error())
	}
	return nil
}
