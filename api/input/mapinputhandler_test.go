package input

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DummyInput is a simple struct used for testing.
type DummyInput struct {
	Name string `json:"name"`
}

// NonStructInput is used to simulate a decoding failure.
type NonStructInput int

// MapInputHandlerTestSuite tests the MapInputHandler.
type MapInputHandlerTestSuite struct {
	suite.Suite
}

// TestMapInputHandlerTestSuite runs the test suite.
func TestMapInputHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MapInputHandlerTestSuite))
}

// factoryDummyInput returns a new DummyInput.
func factoryDummyInput() *DummyInput {
	return new(DummyInput)
}

// factoryNonStructInput returns a new NonStructInput.
func factoryNonStructInput() *NonStructInput {
	return new(NonStructInput)
}

// TestHandle_Success verifies that a valid request is correctly processed.
func (s *MapInputHandlerTestSuite) TestHandle_Success() {
	// Define APIFields with a required "name" field.
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
		},
	)
	// No conversion or custom validation rules.
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Create the handler for DummyInput.
	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	).MustValidateAPIFields()

	// Create a request with a query parameter "name=John".
	req := httptest.NewRequest("GET", "http://example.com/?name=John", nil)
	if err := req.ParseForm(); err != nil {
		s.T().Fatalf("Failed to parse form: %v", err)
	}
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), input)
	assert.Equal(s.T(), "John", input.Name)
}

// TestHandle_MissingRequiredField verifies that missing required fields lead to
// validation errors.
func (s *MapInputHandlerTestSuite) TestHandle_MissingRequiredField() {
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
		},
	)
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)

	// Create a request with no "name" parameter.
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	if err := req.ParseForm(); err != nil {
		s.T().Fatalf("Failed to parse form: %v", err)
	}
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.Error(s.T(), err)
	assert.Nil(s.T(), input)
	apierr := err.(apierror.APIError)
	errData := apierr.Data().(ErrValidationData)
	assert.Contains(
		s.T(), errData.Errors[0].Message,
		"field \"name\" is required",
	)
}

// TestHandle_InvalidMapping verifies that if mapToObject fails decoding,
// the handler returns an error.
func (s *MapInputHandlerTestSuite) TestHandle_InvalidMapping() {
	// Use NonStructInput (int) so that decoding fails.
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "int", // Expecting an integer.
		},
	)
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryNonStructInput,
	)

	// Provide a value that cannot be converted to int.
	req := httptest.NewRequest("GET", "http://example.com/?name=abc", nil)
	if err := req.ParseForm(); err != nil {
		s.T().Fatalf("Failed to parse form: %v", err)
	}
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.Error(s.T(), err)
	assert.Nil(s.T(), input)
	assert.Contains(s.T(), err.Error(), "error decoding input")
}

// TestHandle_CustomValidationFailure verifies that custom validation rules are
// applied.
func (s *MapInputHandlerTestSuite) TestHandle_CustomValidationFailure() {
	// Create a rule "mustBeJohn" that only accepts the value "John".
	customRules := map[string]func(any) error{
		"mustBeJohn": func(val any) error {
			str, ok := val.(string)
			if !ok || str != "John" {
				return fmt.Errorf("value must be John")
			}
			return nil
		},
	}
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
			Validate: []string{"mustBeJohn"},
		},
	)
	conversionMap := map[string]func(any) any{}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)

	// Create a request with a non-conforming value.
	req := httptest.NewRequest("GET", "http://example.com/?name=Bob", nil)
	if err := req.ParseForm(); err != nil {
		s.T().Fatalf("Failed to parse form: %v", err)
	}
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.Error(s.T(), err)
	assert.Nil(s.T(), input)
	// The error message should indicate the custom validation failure.
	apierr := err.(apierror.APIError)
	errData := apierr.Data().(ErrValidationData)
	msg := errData.Errors[0].Message
	assert.Contains(s.T(), msg, "validation error for field \"name\"")
	assert.Contains(s.T(), msg, "value must be John")
}

// TestValidateAPIFields_Valid verifies that ValidateAPIFields returns no error
// for valid APIFields.
func (s *MapInputHandlerTestSuite) TestValidateAPIFields_Valid() {
	// Define an APIField with a valid validation rule "nonEmpty".
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
			Validate: []string{"nonEmpty"},
		},
	)
	conversionMap := map[string]func(any) any{}
	// Provide a custom rule for "nonEmpty".
	customRules := map[string]func(any) error{
		"nonEmpty": func(val any) error {
			str, ok := val.(string)
			if !ok || str == "" {
				return fmt.Errorf("cannot be empty")
			}
			return nil
		},
	}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)
	h, err := handler.ValidateAPIFields()
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), h)
}

// TestValidateAPIFields_Invalid verifies that ValidateAPIFields returns an
// error when a field has an invalid validation rule.
func (s *MapInputHandlerTestSuite) TestValidateAPIFields_Invalid() {
	// Define an APIField with an invalid rule "nonexistent_rule".
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
			Validate: []string{"nonexistent_rule"},
		},
	)
	conversionMap := map[string]func(any) any{}
	// No custom rule is provided for "nonexistent_rule".
	customRules := map[string]func(any) error{}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)
	_, err := handler.ValidateAPIFields()
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "validateRules")
}

// TestMustValidateAPIFields_Valid verifies that MustValidateAPIFields does not
// panic for valid APIFields.
func (s *MapInputHandlerTestSuite) TestMustValidateAPIFields_Valid() {
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
			Validate: []string{"nonEmpty"},
		},
	)
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{
		"nonEmpty": func(val any) error {
			str, ok := val.(string)
			if !ok || str == "" {
				return fmt.Errorf("cannot be empty")
			}
			return nil
		},
	}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)
	// MustValidateAPIFields should not panic for valid fields.
	assert.NotPanics(s.T(), func() {
		handler.MustValidateAPIFields()
	})
}

// TestMustValidateAPIFields_Invalid verifies that MustValidateAPIFields panics
// when the APIFields are invalid.
func (s *MapInputHandlerTestSuite) TestMustValidateAPIFields_Invalid() {
	fields := NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
			Validate: []string{"nonexistent"},
		},
	)
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{} // No rule for "nonexistent"
	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)
	// MustValidateAPIFields should panic due to invalid validation rule.
	assert.Panics(s.T(), func() {
		handler.MustValidateAPIFields()
	})
}

// TestMapFieldConfigFromAPIFields_DuplicateNestedFields tests that
// mapFieldConfigFromAPIFields returns a map with duplicate field names.
func (s *MapInputHandlerTestSuite) TestMapFieldConfigFromAPIFields_DuplicateNestedFields() {
	// Create nested APIFields with duplicate field names.
	nested := NewAPIFields(
		APIField{
			APIName:  "duplicate",
			Type:     "string",
			Default:  "first",
			Required: false,
		},
		APIField{
			APIName:  "duplicate",
			Type:     "int",
			Default:  "second",
			Required: false,
		},
	)
	// Parent field with nested fields.
	fields := NewAPIFields(
		APIField{
			APIName:  "duplicate",
			Type:     "int32",
			Default:  "third",
			Required: false,
		},
		APIField{
			APIName:  "parent",
			Type:     "object",
			Nested:   nested,
			Required: true,
		},
		APIField{
			APIName:  "duplicate",
			Type:     "int64",
			Default:  "fourth",
			Required: false,
		},
	)
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	)

	// Call the private mapFieldConfigFromAPIFields method.
	cfg, err := handler.mapFieldConfigFromAPIFields(fields)
	require.NoError(s.T(), err)

	// Verify that the "parent" field exists.
	parentCfg, ok := cfg.Fields["parent"]
	require.True(
		s.T(), ok, "parent field should exist in map field config",
	)
	// In the nested configuration, there should be one entry for "duplicate".
	dupCfg, ok := parentCfg.Fields["duplicate"]
	require.True(
		s.T(), ok, "duplicate field should exist in nested map field config",
	)
	// The second definition should have overwritten the first.
	assert.Equal(
		s.T(), "int", dupCfg.ExpectedType,
		"expected the duplicate field's type to be 'int'",
	)
	assert.Equal(
		s.T(), "second", dupCfg.DefaultValue,
		"expected the duplicate field's default to be 'second'",
	)
	assert.Equal(
		s.T(), true, dupCfg.Optional,
		"expected the duplicate field's optional to be false",
	)
}
