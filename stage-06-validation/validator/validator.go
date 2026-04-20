// package validator sets up a single shared validator instance and
// provides a helper to convert validation errors into readable API responses.
package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validate is a single shared instance of the validator.
// We reuse one instance across the whole app — it caches struct metadata internally.
var Validate = validator.New()

// ValidationError represents a single field error returned to the client.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatErrors converts the raw validator.ValidationErrors into readable messages.
// Instead of "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
// we return: {"field": "email", "message": "must be a valid email address"}
func FormatErrors(err error) []ValidationError {
	var errors []ValidationError

	// validator returns a list of FieldError — one per failed field
	for _, fe := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   fieldName(fe),
			Message: fieldMessage(fe),
		})
	}

	return errors
}

// fieldName converts the struct field name to a lowercase json-friendly name.
func fieldName(fe validator.FieldError) string {
	// fe.Field() returns the struct field name e.g. "Email"
	// We lowercase it to match JSON convention
	field := fe.Field()
	if len(field) == 0 {
		return field
	}
	return string(field[0]+32) + field[1:] // simple lowercase first letter
}

// fieldMessage returns a human-readable message for each validation tag.
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
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
