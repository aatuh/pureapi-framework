package errutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APIErrorFactoryTestSuite is a suite of tests for APIErrorFactory.
type APIErrorFactoryTestSuite struct {
	suite.Suite
}

// TestAPIErrorFactoryTestSuite runs the test suite.
func TestAPIErrorFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(APIErrorFactoryTestSuite))
}

// TestNewErrorFactory tests the NewErrorFactory function.
func (s *APIErrorFactoryTestSuite) TestNewErrorFactory() {
	systemID := "test_system"
	factory := NewErrorFactory(systemID)
	assert.Equal(s.T(), systemID, factory.SystemID)
}

// TestAPIError tests the APIError function.
func (s *APIErrorFactoryTestSuite) TestAPIError() {
	systemID := "test_system"
	factory := NewErrorFactory(systemID)
	errID := "TEST_ERROR"
	apiErr := factory.APIError(errID)
	assert.Equal(s.T(), errID, apiErr.ID())
	assert.Equal(s.T(), systemID, apiErr.Origin())
}
