package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestInput defines a sample input struct with various source tags.
type TestInput struct {
	// Explicitly placed in URL.
	ID int `json:"id" source:"url"`
	// No source tag, so default placement depends on method.
	Name string `json:"name"`
	// Explicitly placed in header.
	AuthToken string `json:"auth_token" source:"header"`
	// Explicitly placed in cookie.
	SessionID string `json:"session_id" source:"cookie"`
	// No source tag.
	Flag bool `json:"flag"`
}

// BadInput has an empty JSON tag, which should cause an error.
type BadInput struct {
	Field int `json:"" source:"url"`
}

// ClientInputParserTestSuite is a test suite for ClientInputParser.
type ClientInputParserTestSuite struct {
	suite.Suite
}

// TestClientInputParserTestSuite runs the test suite.
func TestClientInputParserTestSuite(t *testing.T) {
	suite.Run(t, new(ClientInputParserTestSuite))
}

// TestParseInput_Nil verifies that a nil input returns an empty RequestData.
func (s *ClientInputParserTestSuite) TestParseInput_Nil() {
	rd, err := ParseInput("GET", nil)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), rd)
	// When no input is provided, all maps and slices remain nil.
	assert.Nil(s.T(), rd.URLParameters)
	assert.Nil(s.T(), rd.Headers)
	assert.Nil(s.T(), rd.Cookies)
	assert.Nil(s.T(), rd.Body)
}

// TestParseInput_InvalidInput verifies that non-pointer inputs return an error.
func (s *ClientInputParserTestSuite) TestParseInput_InvalidInput() {
	_, err := ParseInput("GET", "not a struct")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "input must be a pointer")
}

// TestParseInput_GET_DefaultPlacement tests a GET request.
// For GET, the default placement for fields with no explicit source is "url".
func (s *ClientInputParserTestSuite) TestParseInput_GET_DefaultPlacement() {
	input := &TestInput{
		ID:        42,
		Name:      "Alice",
		AuthToken: "Bearer token",
		SessionID: "sess123",
		Flag:      true,
	}
	rd, err := ParseInput(http.MethodGet, input)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), rd)

	// Fields with explicit source "url" go to URLParameters.
	require.NotNil(s.T(), rd.URLParameters)
	assert.EqualValues(s.T(), 42, rd.URLParameters["id"])

	// For GET the default placement is URL.
	assert.Equal(s.T(), "Alice", rd.URLParameters["name"])

	// Header placement.
	require.NotNil(s.T(), rd.Headers)
	assert.Equal(s.T(), "Bearer token", rd.Headers["auth_token"])

	// Cookie placement.
	require.NotNil(s.T(), rd.Cookies)
	var found bool
	for _, c := range rd.Cookies {
		if c.Name == "session_id" && c.Value == "sess123" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected cookie session_id to be present")

	// Boolean field, default for GET is URL.
	assert.Equal(s.T(), true, rd.URLParameters["flag"])
}

// TestParseInput_POST_DefaultPlacement tests a POST request.
// For POST the default placement is "body".
func (s *ClientInputParserTestSuite) TestParseInput_POST_DefaultPlacement() {
	input := &TestInput{
		ID:        101,
		Name:      "Bob",
		AuthToken: "Token123",
		SessionID: "cookie456",
		Flag:      false,
	}
	rd, err := ParseInput(http.MethodPost, input)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), rd)

	// Explicit "url" tag remains unchanged.
	require.NotNil(s.T(), rd.URLParameters)
	assert.EqualValues(s.T(), 101, rd.URLParameters["id"])

	// For POST the default placement is "body".
	require.NotNil(s.T(), rd.Body)
	assert.Equal(s.T(), "Bob", rd.Body["name"])

	// Headers.
	require.NotNil(s.T(), rd.Headers)
	assert.Equal(s.T(), "Token123", rd.Headers["auth_token"])

	// Cookies.
	require.NotNil(s.T(), rd.Cookies)
	var found bool
	for _, c := range rd.Cookies {
		if c.Name == "session_id" && c.Value == "cookie456" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected cookie session_id to be present")

	// Boolean field goes to body.
	assert.Equal(s.T(), false, rd.Body["flag"])
}

// TestPlaceField_InvalidSourceTag tests that an invalid source tag causes an
// error.
func (s *ClientInputParserTestSuite) TestPlaceField_InvalidSourceTag() {
	rd := NewRequestData()
	err := rd.placeFieldValue("invalid", "field", "value")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid source tag")
}
