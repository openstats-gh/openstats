package validation

import (
	"fmt"
	"net/mail"
	"slices"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
)

const (
	MaxDisplayNameLength = 64
	MinDisplayNameLength = 1
	MaxSlugNameLength    = 64
	MinSlugNameLength    = 2
	MaxPasswordLength    = 32
	MinPasswordLength    = 10
)

var ValidSlugSpecialCharacters = []rune("!@#$%^&*")

var goValidator = validator.New(validator.WithRequiredStructEnabled())

type ConflictError struct {
	Field string
	Value string
}

func (v *ConflictError) Error() string {
	return fmt.Sprintf("Conflict on '%s' with value '%s'", v.Field, v.Value)
}

func Conflict(field string, value string) *ConflictError {
	return &ConflictError{
		Field: field,
		Value: value,
	}
}

type ValidationFieldError struct {
	Field string
	Tag   string
	Value interface{}
}

type ValidationError struct {
	Errors []ValidationFieldError
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("Validation Errors: %v", v.Errors)
}

func Validate(data interface{}) error {
	var resultErrors []ValidationFieldError

	errs := goValidator.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			resultErrors = append(resultErrors, ValidationFieldError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: err.Value(),
			})
		}
	}

	if len(resultErrors) > 0 {
		return &ValidationError{
			Errors: resultErrors,
		}
	}

	return nil
}

var validationDetails = map[string]string{
	"email": "must be a valid email address",
	"slug":  fmt.Sprintf("must be at least %v in length, be no more than %v in length; must be all lowercase, start with a lowercase letter, and may only contain letters, numbers, and non-consecutive dashes", MinSlugNameLength, MaxSlugNameLength),

	//   - password is at least MinPasswordLength and no more than MaxPasswordLength in length
	//   - password contains only latin characters, numbers, or some special characters: !@#$%^&*
	"password":    fmt.Sprintf("must be at least %v in length, be no more than %v in length; must only contain letters, numbers, or some special characters: !@#$%%^&*", MinPasswordLength, MaxPasswordLength),
	"displayName": fmt.Sprintf("must be at least %v in length, be no more than %v in length; may contain any renderable unicode character", MinDisplayNameLength, MaxDisplayNameLength),
}

func GetValidationDetail(tag string) string {
	value, ok := validationDetails[tag]
	if ok {
		return value
	}

	return fmt.Sprintf("needs to implement '%s'", tag)
}

func SetupValidations() error {
	err := goValidator.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return ValidSlug(fl.Field().String())
	})
	if err != nil {
		return err
	}

	err = goValidator.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		return ValidPassword(fl.Field().String())
	})
	if err != nil {
		return err
	}

	err = goValidator.RegisterValidation("displayName", func(fl validator.FieldLevel) bool {
		return ValidDisplayName(fl.Field().String())
	})
	if err != nil {
		return err
	}

	return nil
}

func ValidEmailAddress(email string) bool {
	_, emailErr := mail.ParseAddress(email)
	return emailErr == nil
}

func ValidDisplayName(displayName string) bool {
	return len(displayName) >= MinDisplayNameLength && len(displayName) <= MaxDisplayNameLength
}

// ValidSlug returns true if all of these rules are followed:
//   - slug is at least MinSlugNameLength and no more than MaxSlugNameLength in length
//   - slug is all lowercase
//   - slug contains only latin characters, numbers, or a dash
func ValidSlug(slug string) bool {
	if len(slug) < MinSlugNameLength || len(slug) > MaxSlugNameLength {
		return false
	}

	for _, r := range []rune(slug) {
		if !unicode.IsLower(r) && !unicode.IsNumber(r) && !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}

	return true
}

// ValidPassword returns true if all of these rules are followed:
//   - password is at least MinPasswordLength and no more than MaxPasswordLength in length
//   - password contains only latin characters, numbers, or some special characters: !@#$%^&*
func ValidPassword(password string) bool {
	if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
		return false
	}

	for _, r := range []rune(password) {
		if !unicode.IsNumber(r) && !unicode.IsLetter(r) && !slices.Contains(ValidSlugSpecialCharacters, r) {
			return false
		}
	}

	return true
}

func ErrorIsAny(err error, errs ...error) bool {
	for _, other := range errs {
		if eris.Is(err, other) {
			return true
		}
	}

	return false
}
