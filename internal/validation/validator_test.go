package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Email    string `json:"email"    validate:"required,email"`
	Name     string `json:"name"     validate:"required,min=2,max=50"`
	Age      int    `json:"age"      validate:"gte=0,lte=150"`
	Role     string `json:"role"     validate:"oneof=admin user guest"`
	Password string `json:"password" validate:"required,min=8"`
}

func validTestStruct() testStruct {
	return testStruct{
		Email:    "test@example.com",
		Name:     "John Doe",
		Age:      25,
		Role:     "user",
		Password: "password123",
	}
}

func TestValidator_Validate_Valid(t *testing.T) {
	t.Parallel()

	v := New()
	err := v.Validate(validTestStruct())
	assert.NoError(t, err)
}

func TestValidator_Validate_MissingRequired(t *testing.T) {
	t.Parallel()

	v := New()
	s := testStruct{Name: "John Doe", Age: 25, Role: "user", Password: "password123"}
	err := v.Validate(s)
	require.Error(t, err)

	var validationErrs ValidationErrors
	require.ErrorAs(t, err, &validationErrs)
	assert.Len(t, validationErrs, 1)
	assert.Equal(t, "email", validationErrs[0].Field)
	assert.Equal(t, "is required", validationErrs[0].Message)
}

func TestValidator_Validate_InvalidEmail(t *testing.T) {
	t.Parallel()

	v := New()
	s := validTestStruct()
	s.Email = "not-an-email"
	err := v.Validate(s)
	require.Error(t, err)

	var validationErrs ValidationErrors
	require.ErrorAs(t, err, &validationErrs)
	assert.Len(t, validationErrs, 1)
	assert.Equal(t, "email", validationErrs[0].Field)
	assert.Equal(t, "must be a valid email address", validationErrs[0].Message)
}

func TestValidator_Validate_BelowMin(t *testing.T) {
	t.Parallel()

	v := New()
	s := validTestStruct()
	s.Password = "short"
	err := v.Validate(s)
	require.Error(t, err)

	var validationErrs ValidationErrors
	require.ErrorAs(t, err, &validationErrs)
	assert.Len(t, validationErrs, 1)
	assert.Equal(t, "password", validationErrs[0].Field)
	assert.Equal(t, "must be at least 8 characters", validationErrs[0].Message)
}

func TestValidator_Validate_InvalidOneof(t *testing.T) {
	t.Parallel()

	v := New()
	s := validTestStruct()
	s.Role = "invalid"
	err := v.Validate(s)
	require.Error(t, err)

	var validationErrs ValidationErrors
	require.ErrorAs(t, err, &validationErrs)
	assert.Len(t, validationErrs, 1)
	assert.Equal(t, "role", validationErrs[0].Field)
	assert.Equal(t, "must be one of: admin user guest", validationErrs[0].Message)
}

func TestValidator_Validate_MultipleErrors(t *testing.T) {
	t.Parallel()

	v := New()
	s := testStruct{Email: "not-valid", Name: "A", Age: 200, Role: "unknown"}
	err := v.Validate(s)
	require.Error(t, err)

	var validationErrs ValidationErrors
	require.ErrorAs(t, err, &validationErrs)
	assert.GreaterOrEqual(t, len(validationErrs), 4)
}

func TestValidationErrors_Error(t *testing.T) {
	t.Parallel()

	t.Run("empty errors returns default message", func(t *testing.T) {
		t.Parallel()

		ve := ValidationErrors{}
		assert.Equal(t, "validation failed", ve.Error())
	})

	t.Run("single error formatted correctly", func(t *testing.T) {
		t.Parallel()

		ve := ValidationErrors{
			{Field: "email", Message: "is required"},
		}
		assert.Equal(t, "email: is required", ve.Error())
	})

	t.Run("multiple errors joined with semicolon", func(t *testing.T) {
		t.Parallel()

		ve := ValidationErrors{
			{Field: "email", Message: "is required"},
			{Field: "name", Message: "must be at least 2 characters"},
		}
		assert.Equal(t, "email: is required; name: must be at least 2 characters", ve.Error())
	})
}

func TestValidationErrors_ToMap(t *testing.T) {
	t.Parallel()

	ve := ValidationErrors{
		{Field: "email", Message: "is required"},
		{Field: "name", Message: "must be at least 2 characters"},
	}

	result := ve.ToMap()
	assert.Equal(t, "is required", result["email"])
	assert.Equal(t, "must be at least 2 characters", result["name"])
}

type usernamePatternTestStruct struct {
	Username string `json:"username" validate:"username_pattern"`
}

func TestValidator_UsernamePattern(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	v := New()

	tests := []struct {
		name      string
		username  string
		expectErr bool
	}{
		{
			name:      "Valid - alphanumeric lowercase",
			username:  "user123",
			expectErr: false,
		},
		{
			name:      "Valid - alphanumeric uppercase",
			username:  "USER123",
			expectErr: false,
		},
		{
			name:      "Valid - with underscore",
			username:  "user_name",
			expectErr: false,
		},
		{
			name:      "Valid - all underscores",
			username:  "___",
			expectErr: false,
		},
		{
			name:      "Valid - mixed case with underscores",
			username:  "User_Name_123",
			expectErr: false,
		},
		{
			name:      "Invalid - with @",
			username:  "user@name",
			expectErr: true,
		},
		{
			name:      "Invalid - with space",
			username:  "user name",
			expectErr: true,
		},
		{
			name:      "Invalid - with hyphen",
			username:  "user-name",
			expectErr: true,
		},
		{
			name:      "Invalid - with dot",
			username:  "user.name",
			expectErr: true,
		},
		{
			name:      "Invalid - with special chars",
			username:  "user!@#$",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := v.Validate(usernamePatternTestStruct{Username: tt.username})
			if tt.expectErr {
				require.Error(t, err)

				var validationErrs ValidationErrors
				require.ErrorAs(t, err, &validationErrs)
				assert.Len(t, validationErrs, 1)
				assert.Equal(t, "username", validationErrs[0].Field)
				assert.Contains(t, validationErrs[0].Message, "alphanumeric")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
