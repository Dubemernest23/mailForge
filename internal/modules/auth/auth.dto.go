package auth

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// RegisterRequest, LoginRequest, RefreshRequest, LogoutRequest, AuthRespons
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	hasNumber     = regexp.MustCompile(`[0-9]`)
	// This matches any character that is NOT a letter, number, or whitespace
	hasSpecialChar = regexp.MustCompile(`[^a-zA-Z0-9\s]`)
)

var validate = newValidator()

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,username_alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,secure_password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,secure_password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,uuid4"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,uuid4"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`  // Signed RS256 JWT
	RefreshToken string `json:"refresh_token"` // Opaque UUID
	ExpiresIn    int    `json:"expires_in"`    // S
}

func (r RegisterRequest) Validate() error {
	return translateErr(validate.Struct(r))
}

func (r LoginRequest) Validate() error {
	return translateErr(validate.Struct(r))
}

func (r LogoutRequest) Validate() error {
	return translateErr(validate.Struct(r))
}

func (r RefreshRequest) Validate() error {
	return translateErr(validate.Struct(r))
}

// NewValidator() configurs and returns a validator instance
func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	// register custom rule
	_ = v.RegisterValidation("username_alphanum", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})

	_ = v.RegisterValidation("secure_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return hasNumber.MatchString(password) && hasSpecialChar.MatchString(password)
	})

	return v
}

// translate the error - the first error from the srray of errmsg should be displayed in human readable form

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
		case "email":
			return fmt.Errorf("%s must be a valid email address", field)
		case "min":
			return fmt.Errorf("%s must be at least %s characters long", field, firstErr.Param())
		case "uuid4":
			return fmt.Errorf("%s must be a valid UUIDv4", field)
		case "username_alphanum":
			return fmt.Errorf("%s can only contain alphanumeric characters and underscores", field)
		case "secure_password":
			return fmt.Errorf("%s must contain at least one number and one special character", field)
		default:
			return fmt.Errorf("%s is invalid", field)
		}

	}
	return err
}
