// package validator wraps go-playground/validator to provide:
//   1. A single shared Validate instance (creating it is expensive — do it once)
//   2. Human-readable error messages (validator's default messages are terse)
//   3. Lowercase field names in errors (Go structs use PascalCase, JSON uses camelCase)
package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validate is the shared validator instance.
// Use this everywhere: v.Validate.Struct(req) or v.Validate.Var(val, "uuid4")
var Validate = validator.New()

// ValidationError is one field's error in a 422 response.
// Example: {"field": "email", "message": "must be a valid email address"}
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatErrors converts validator.ValidationErrors into our API error format.
// Called after v.Validate.Struct(req) returns a non-nil error.
func FormatErrors(err error) []ValidationError {
	var errors []ValidationError
	for _, fe := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   lowerFirst(fe.Field()), // "Name" → "name"
			Message: fieldMessage(fe),
		})
	}
	return errors
}

// lowerFirst lowercases the first character of a string.
// Go struct fields are PascalCase; JSON expects camelCase.
// "Name" → "name", "Email" → "email"
func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	// ASCII lowercase: add 32 to the byte value of the first character
	return string(s[0]+32) + s[1:]
}

// fieldMessage returns a human-readable message for a validation failure.
// The tag is the rule that failed: "required", "email", "min", "max", "uuid4"
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
