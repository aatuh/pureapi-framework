package input

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OffensiveDummyInput and Address are used to simulate a more complex input.
type OffensiveDummyInput struct {
	Name    string  `json:"name"`
	Address Address `json:"address"`
}

type Address struct {
	City string `json:"city"`
	Zip  string `json:"zip"`
}

// OffensiveMapInputHandlerTestSuite is an offensive test suite for
// MapInputHandler.
type OffensiveMapInputHandlerTestSuite struct {
	suite.Suite
}

// TestOffensiveMapInputHandlerTestSuite runs the test suite.
func TestOffensiveMapInputHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OffensiveMapInputHandlerTestSuite))
}

// factoryOffensiveDummyInput returns a new OffensiveDummyInput.
func factoryOffensiveDummyInput() *OffensiveDummyInput {
	return new(OffensiveDummyInput)
}

// setupOffensiveFields constructs APIFields for OffensiveDummyInput.
func setupOffensiveFields() APIFields {
	return NewAPIFields(
		APIField{
			APIName:  "name",
			Required: true,
			Type:     "string",
		},
		APIField{
			APIName:  "address",
			Required: true,
			Type:     "object",
			Nested: NewAPIFields(
				APIField{
					APIName:  "city",
					Required: true,
					Type:     "string",
				},
				APIField{
					APIName: "zip",
					Type:    "string",
				},
			),
		},
	)
}

// TestOffensive_LongStringInput sends an extremely long string to test for
// performance and buffer issues.
func (s *OffensiveMapInputHandlerTestSuite) TestOffensive_LongStringInput() {
	fields := setupOffensiveFields()
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Build an extremely long string.
	longName := strings.Repeat("A", 10000)
	req := httptest.NewRequest("GET", "http://example.com/?name="+longName+
		"&address.city=Metropolis&address.zip=12345", nil)
	require.NoError(s.T(), req.ParseForm())

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryOffensiveDummyInput,
	).MustValidateAPIFields()
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), input)
	assert.Equal(s.T(), longName, input.Name)
	assert.Equal(s.T(), "Metropolis", input.Address.City)
	assert.Equal(s.T(), "12345", input.Address.Zip)
}

// TestOffensive_InvalidNestedStructure provides an invalid nested structure
// value to verify that the handler returns a safe error.
func (s *OffensiveMapInputHandlerTestSuite) TestOffensive_InvalidNestedStructure() {
	fields := setupOffensiveFields()
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Set address field to a number instead of an object.
	req := httptest.NewRequest(
		"GET", "http://example.com/?name=John&address=123", nil,
	)
	require.NoError(s.T(), req.ParseForm())

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryOffensiveDummyInput,
	).MustValidateAPIFields()
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.Error(s.T(), err)
	assert.Nil(s.T(), input)
	apiErr := err.(apierror.APIError)
	data := apiErr.Data().(ErrValidationData)
	assert.Contains(
		s.T(), data.Errors[0].Message,
		"validate: field \"address\" is required",
	)

}

// TestOffensive_ExtraQueryParameters verifies that extra unexpected parameters
// are safely ignored.
func (s *OffensiveMapInputHandlerTestSuite) TestOffensive_ExtraQueryParameters() {
	fields := setupOffensiveFields()
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Provide extra parameter "extra" which is not defined in APIFields.
	req := httptest.NewRequest(
		"GET",
		"http://example.com/?name=Alice&address.city=Gotham"+
			"&address.zip=54321&extra=unexpected",
		nil,
	)
	require.NoError(s.T(), req.ParseForm())

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryOffensiveDummyInput,
	).MustValidateAPIFields()
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), input)
	assert.Equal(s.T(), "Alice", input.Name)
	assert.Equal(s.T(), "Gotham", input.Address.City)
	assert.Equal(s.T(), "54321", input.Address.Zip)
}

// TestOffensive_BinaryDataInInput simulates binary (non-UTF8) data in a field.
func (s *OffensiveMapInputHandlerTestSuite) TestOffensive_BinaryDataInInput() {
	fields := setupOffensiveFields()
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Insert binary data in the "name" field.
	binaryData := string([]byte{0xff, 0xfe, 0xfd})
	req := httptest.NewRequest(
		"GET",
		"http://example.com/?name="+binaryData+
			"&address.city=BinaryCity&address.zip=000",
		nil,
	)
	require.NoError(s.T(), req.ParseForm())

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryOffensiveDummyInput,
	).MustValidateAPIFields()
	w := httptest.NewRecorder()

	input, err := handler.Handle(w, req)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), input)
	assert.Equal(s.T(), binaryData, input.Name)
	assert.Equal(s.T(), "BinaryCity", input.Address.City)
	assert.Equal(s.T(), "000", input.Address.Zip)
}

// TestOffensive_DeeplyNestedFields constructs a very deep nesting structure to
// test for stack overflows.
func (s *OffensiveMapInputHandlerTestSuite) TestOffensive_DeeplyNestedFields() {
	// Create a deeply nested structure.
	depth := 50
	var nested APIFields = NewAPIFields(
		APIField{
			APIName:  "value",
			Required: true,
			Type:     "string",
		},
	)
	for i := 0; i < depth; i++ {
		nested = NewAPIFields(
			APIField{
				APIName:  fmt.Sprintf("level%d", i),
				Required: true,
				Type:     "object",
				Nested:   nested,
			},
		)
	}
	fields := NewAPIFields(
		APIField{
			APIName:  "root",
			Required: true,
			Type:     "object",
			Nested:   nested,
		},
	)
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Construct a query string with nested keys.
	q := "root.level0.level1.level2.level3.level4.level5.level6.level7.level8" +
		".level9.level10.level11.level12.level13.level14.level15.level16" +
		".level17.level18.level19.level20.level21.level22.level23.level24" +
		".level25.level26.level27.level28.level29.level30.level31.level32" +
		".level33.level34.level35.level36.level37.level38.level39.level40" +
		".level41.level42.level43.level44.level45.level46.level47.level48" +
		".level49.value=Deep"
	req := httptest.NewRequest("GET", "http://example.com/?"+q, nil)
	require.NoError(s.T(), req.ParseForm())

	// Use DummyInput here because it does not match OffensiveDummyInput.
	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryDummyInput,
	).MustValidateAPIFields()
	w := httptest.NewRecorder()
	_, err := handler.Handle(w, req)
	// The handler should not panic even if the nested chain is very deep.
	s.T().Logf(
		"Deeply nested input produced error (acceptable if graceful): %v", err,
	)
}

// TestOffensive_MalformedQueryParameters simulates malformed query keys.
func (s *OffensiveMapInputHandlerTestSuite) TestOffensive_MalformedQueryParameters() {
	fields := setupOffensiveFields()
	conversionMap := map[string]func(any) any{}
	customRules := map[string]func(any) error{}

	// Manually set req.Form with malformed keys.
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	req.Form = map[string][]string{
		"name":           {"Alice"},
		"address.city":   {"Wonderland"},
		"address.zip":    {"99999"},
		"bad\xffkey":     {"malicious"},
		"another\xfekey": {"payload"},
	}

	handler := NewMapInputHandler(
		fields, conversionMap, customRules, factoryOffensiveDummyInput,
	).MustValidateAPIFields()
	w := httptest.NewRecorder()
	_, err := handler.Handle(w, req)
	require.Error(s.T(), err)
	apiErr, ok := err.(apierror.APIError)
	require.True(s.T(), ok, "expected error to be an APIError")
	assert.Equal(s.T(), "VALIDATION_ERROR", apiErr.ID())
}
