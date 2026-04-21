package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatErrors(err error) []ValidationError {
	var errors []ValidationError
	for _, fe := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   lowerFirst(fe.Field()),
			Message: fieldMessage(fe),
		})
	}
	return errors
}

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]+32) + s[1:]
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("minimum length is %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("maximum length is %s characters", fe.Param())
	case "uuid4":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
