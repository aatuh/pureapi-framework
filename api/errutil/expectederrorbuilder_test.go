package errutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ExpectedErrorBuilderTestSuite is a suite of tests for ExpectedErrorBuilder.
type ExpectedErrorBuilderTestSuite struct {
	suite.Suite
}

// TestExpectedErrorBuilderTestSuite runs the test suite.
func TestExpectedErrorBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ExpectedErrorBuilderTestSuite))
}

// TestNewExpectedErrorBuilder tests that NewExpectedErrorBuilder returns a new
// ExpectedErrorBuilder.
func (s *ExpectedErrorBuilderTestSuite) TestNewExpectedErrorBuilder() {
	systemID := "builder_origin"
	builder := NewExpectedErrorBuilder(systemID)
	// Without adding any errors, Build should return an empty slice.
	errs := builder.Build()
	assert.Len(s.T(), errs, 0)
}

// TestWithErrorsAndBuild tests that WithErrors and Build work together.
func (s *ExpectedErrorBuilderTestSuite) TestWithErrorsAndBuild() {
	systemID := "builder_origin"
	builder := NewExpectedErrorBuilder(systemID)
	// Create expected errors with dummy origins.
	e1 := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	e2 := NewExpectedError("ERR2", http.StatusInternalServerError, "origin2")
	// Add errors to the builder.
	builder = builder.WithErrors(ExpectedErrors{e1})
	builder = builder.WithErrors(ExpectedErrors{e2})
	errs := builder.Build()
	// Build should return errors with origin equal to the systemID.
	assert.Len(s.T(), errs, 2)
	for _, e := range errs {
		assert.Equal(s.T(), systemID, e.Origin)
	}
	// Verify that the IDs and statuses remain unchanged.
	assert.Equal(s.T(), "ERR1", errs[0].ID)
	assert.Equal(s.T(), http.StatusBadRequest, errs[0].Status)
	assert.Equal(s.T(), "ERR2", errs[1].ID)
	assert.Equal(s.T(), http.StatusInternalServerError, errs[1].Status)
}
