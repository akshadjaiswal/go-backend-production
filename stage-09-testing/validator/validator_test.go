package validator_test

// Why "validator_test" and not "validator"?
//
// In Go you have two choices for test package names:
//   1. Same package:  "package validator"   — can access unexported functions
//   2. Black-box:     "package validator_test" — tests the PUBLIC API only
//
// We use the black-box style here because FormatErrors is the exported public API.
// If our tests pass using only what external callers can see, the package is well-designed.
// This also avoids accidental reliance on internal details.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v "github.com/akshadjaiswal/go-backend-production/stage-09-testing/validator"
)

// testStruct is a simple struct we use to trigger specific validation failures.
// The validate tags tell go-playground/validator what rules to apply.
type testStruct struct {
	Name  string `validate:"required,min=2,max=10"`
	Email string `validate:"required,email"`
}

// TestFormatErrors is a table-driven test for the FormatErrors function.
//
// Table-driven test pattern — the Go idiom:
//   - Define a slice of test cases (the "table") as an anonymous struct
//   - Each case has: a name, input, and expected output
//   - Loop over all cases with t.Run — each gets its own subtest
//
// Why table-driven?
//   - All cases for one function live together — easy to scan
//   - Adding a new edge case = one new line in the table
//   - If one case fails, others still run — you see all failures at once
func TestFormatErrors(t *testing.T) {
	tests := []struct {
		name           string
		input          any          // the struct to validate
		expectedField  string       // which field name we expect in the error
		expectedMsg    string       // which message we expect
		expectNoErrors bool         // true when the input is valid — no errors expected
	}{
		{
			name:          "required field missing",
			input:         testStruct{Email: "test@example.com"}, // Name is empty
			expectedField: "name",                                 // FormatErrors lowercases the first letter
			expectedMsg:   "Name is required",                     // fieldMessage for "required" tag
		},
		{
			name:          "invalid email",
			input:         testStruct{Name: "Akshad", Email: "not-an-email"},
			expectedField: "email",
			expectedMsg:   "must be a valid email address",
		},
		{
			name:          "min length violation",
			input:         testStruct{Name: "A", Email: "a@b.com"}, // Name too short (min=2)
			expectedField: "name",
			expectedMsg:   "minimum length is 2 characters",
		},
		{
			name:          "max length violation",
			input:         testStruct{Name: "TooLongNameXYZ", Email: "a@b.com"}, // Name too long (max=10)
			expectedField: "name",
			expectedMsg:   "maximum length is 10 characters",
		},
		{
			name:           "valid input — no errors",
			input:          testStruct{Name: "Akshad", Email: "akshad@example.com"},
			expectNoErrors: true,
		},
	}

	for _, tc := range tests {
		// t.Run creates a subtest with the given name.
		// If it fails, the output shows "TestFormatErrors/required_field_missing" — easy to find.
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate.Struct(tc.input)

			if tc.expectNoErrors {
				// require.NoError stops this subtest immediately if err != nil.
				// We use require here because if there's an error, the rest of the
				// test would panic trying to call FormatErrors on a non-nil error.
				require.NoError(t, err, "expected valid input to produce no validation errors")
				return
			}

			// For invalid inputs, we expect a validation error
			require.Error(t, err, "expected validation error but got nil")

			// FormatErrors converts the raw validation error into our JSON-friendly format
			errs := v.FormatErrors(err)
			require.NotEmpty(t, errs, "FormatErrors should return at least one error")

			// Find the error for our expected field
			// There may be multiple errors (e.g. both required + email fail) so we search
			found := false
			for _, e := range errs {
				if e.Field == tc.expectedField {
					// assert.Equal shows a nice diff if they don't match:
					// Expected: "Name is required"
					// Actual:   "name is required"
					assert.Equal(t, tc.expectedMsg, e.Message,
						"wrong message for field %q", tc.expectedField)
					found = true
					break
				}
			}
			assert.True(t, found, "expected error for field %q but it wasn't in: %v",
				tc.expectedField, errs)
		})
	}
}

// TestFormatErrors_MultipleErrors verifies that when multiple fields fail,
// FormatErrors returns all of them — not just the first one.
func TestFormatErrors_MultipleErrors(t *testing.T) {
	// Both Name and Email are missing/invalid
	input := testStruct{Name: "", Email: "bad"}

	err := v.Validate.Struct(input)
	require.Error(t, err)

	errs := v.FormatErrors(err)

	// Should have at least 2 errors (Name required + Email invalid)
	assert.GreaterOrEqual(t, len(errs), 2,
		"expected multiple validation errors, got %d: %v", len(errs), errs)
}

// TestValidationError_FieldIsLowercased verifies the lowerFirst behavior specifically.
// The go-playground validator gives us "Name" (capitalized from the struct field name).
// Our FormatErrors lowercases the first letter → "name" for JSON output.
func TestValidationError_FieldIsLowercased(t *testing.T) {
	type S struct {
		UserName string `validate:"required"` // two words, PascalCase
	}

	err := v.Validate.Struct(S{})
	require.Error(t, err)

	errs := v.FormatErrors(err)
	require.Len(t, errs, 1)

	// "UserName" → "uUserName"? No. lowerFirst only lowercases the FIRST letter.
	// So "UserName" → "userName"
	assert.Equal(t, "userName", errs[0].Field,
		"FormatErrors should lowercase only the first letter of the field name")
}

// TestFormatErrors_UUIDTag verifies the uuid4 tag produces the right message.
// This tag is used in handler UUID validation: v.Validate.Var(id, "required,uuid4")
func TestFormatErrors_UUIDTag(t *testing.T) {
	type S struct {
		ID string `validate:"required,uuid4"`
	}

	err := v.Validate.Struct(S{ID: "not-a-uuid"})
	require.Error(t, err)

	errs := v.FormatErrors(err)

	// Find the uuid4 error
	var uuidErr *v.ValidationError
	for i := range errs {
		if errs[i].Field == "iD" || errs[i].Field == "id" || errs[i].Field == "ID" {
			uuidErr = &errs[i]
			break
		}
	}

	// The field name depends on the struct — just check the message
	found := false
	for _, e := range errs {
		if e.Message == "must be a valid UUID" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'must be a valid UUID' message, got: %v, uuidErr: %v", errs, uuidErr)
}
