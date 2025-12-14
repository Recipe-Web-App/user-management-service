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
