package list

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

type CreateListRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Description string `json:"description"`
}

type UpdateListRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1"`
	Description string `json:"description"`
}

type ListResponse struct {
	PublicID    string    `json:"public_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r CreateListRequest) Validate() error {
	return translateErr(validate.Struct(r))
}

func (r UpdateListRequest) Validate() error {
	return translateErr(validate.Struct(r))
}

func newValidator() *validator.Validate {
	return validator.New()
}

func translateErr(err error) error {
	if err == nil {
		return nil
	}

	var valErr validator.ValidationErrors

	if errors.As(err, &valErr) {
		firstErr := valErr[0]
		field := firstErr.Field()

		switch firstErr.Tag() {
		case "required":
			return fmt.Errorf("%s is required", field)
		case "min":
			return fmt.Errorf("%s must be at least %s characters", field, firstErr.Param())
		default:
			return fmt.Errorf("%s is invalid", field)
		}
	}

	return err
}
