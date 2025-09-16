package errutil

import (
	"net/http"
	"testing"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExpectedErrorTestSuite struct {
	suite.Suite
}

// TestNewExpectedError creates a new ExpectedError.
func (s *ExpectedErrorTestSuite) TestNewExpectedError() {
	e := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	assert.Equal(s.T(), "ERR1", e.ID)
	assert.Equal(s.T(), http.StatusBadRequest, e.Status)
	assert.Equal(s.T(), "origin1", e.Origin)
}

// TestWithID verifies that WithID updates the ID field.
func (s *ExpectedErrorTestSuite) TestWithID() {
	e := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	updated := e.WithID("ERR2")
	assert.Equal(s.T(), "ERR2", updated.ID)
	assert.Equal(s.T(), "ERR1", e.ID)
}

// TestWithMaskedID verifies that WithMaskedID updates the maskedID field.
func (s *ExpectedErrorTestSuite) TestWithMaskedID() {
	e := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	updated := e.WithMaskedID("MASKED")
	assert.Equal(s.T(), "MASKED", updated.MaskedID)
	assert.Equal(s.T(), "", e.MaskedID)
}

// TestWithStatus verifies that WithStatus updates the status field.
func (s *ExpectedErrorTestSuite) TestWithStatus() {
	e := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	updated := e.WithStatus(http.StatusInternalServerError)
	assert.Equal(s.T(), http.StatusInternalServerError, updated.Status)
	assert.Equal(s.T(), http.StatusBadRequest, e.Status)
}

// TestWithPublicData verifies that WithPublicData updates the publicData field.
func (s *ExpectedErrorTestSuite) TestWithPublicData() {
	e := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	updated := e.WithPublicData(true)
	assert.True(s.T(), updated.PublicData)
	assert.False(s.T(), e.PublicData)
}

// TestWithOrigin verifies that WithOrigin updates the origin field.
func (s *ExpectedErrorTestSuite) TestWithOrigin() {
	e := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	updated := e.WithOrigin("new_origin")
	assert.Equal(s.T(), "new_origin", updated.Origin)
	assert.Equal(s.T(), "origin1", e.Origin)
}

// TestMaskAPIError masks the ID and data of the given API error based on the
// configuration of the ExpectedError.
func (s *ExpectedErrorTestSuite) TestMaskAPIError() {
	cases := []struct {
		name         string
		expectedErr  ExpectedError
		inputAPIErr  apierror.APIError
		expectedID   string
		expectedData any
	}{
		{
			name: "with public data and maskedID",
			expectedErr: NewExpectedError(
				"SAMPLE_ERROR", http.StatusBadRequest, "origin",
			).WithMaskedID("MASKED_ERROR").WithPublicData(true),
			inputAPIErr: apierror.NewAPIError("SAMPLE_ERROR").
				WithData("payload"),
			expectedID:   "MASKED_ERROR",
			expectedData: "payload",
		},
		{
			name: "without public data and no maskedID",
			expectedErr: NewExpectedError(
				"SAMPLE_ERROR", http.StatusBadRequest, "origin",
			).WithPublicData(false),
			inputAPIErr: apierror.NewAPIError("SAMPLE_ERROR").
				WithData("payload"),
			expectedID:   "SAMPLE_ERROR",
			expectedData: nil,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			status, maskedErr := tc.expectedErr.MaskAPIError(tc.inputAPIErr)
			assert.Equal(s.T(), tc.expectedErr.Status, status)
			assert.Equal(s.T(), tc.expectedID, maskedErr.ID())
			assert.Equal(s.T(), tc.expectedErr.Origin, maskedErr.Origin())
			assert.Equal(s.T(), tc.expectedData, maskedErr.Data())
		})
	}
}

// TestExpectedErrorsWithErrors adds new errors to the slice.
func (s *ExpectedErrorTestSuite) TestExpectedErrorsWithErrors() {
	// Test WithErrors on the ExpectedErrors slice.
	errs := ExpectedErrors{
		NewExpectedError("ERR1", http.StatusBadRequest, "origin1"),
	}
	newErr := NewExpectedError(
		"ERR2", http.StatusInternalServerError, "origin2",
	)
	combined := errs.WithErrors(newErr)
	assert.Len(s.T(), combined, 2)
	assert.Equal(s.T(), "ERR1", combined[0].ID)
	assert.Equal(s.T(), "ERR2", combined[1].ID)
}

// TestExpectedErrorsWithOrigin tests that WithOrigin updates the origin for all
// errors in the slice.
func (s *ExpectedErrorTestSuite) TestExpectedErrorsWithOrigin() {
	// WithOrigin should update the origin for all errors in the slice.
	errs := ExpectedErrors{
		NewExpectedError("ERR1", http.StatusBadRequest, "origin1"),
		NewExpectedError("ERR2", http.StatusInternalServerError, "origin2"),
	}
	newOrigin := "new_origin"
	updated := errs.WithOrigin(newOrigin)
	for _, e := range updated {
		assert.Equal(s.T(), newOrigin, e.Origin)
	}
}

// TestExpectedErrorsGetByID tests that GetByID returns the expected error.
func (s *ExpectedErrorTestSuite) TestExpectedErrorsGetByID() {
	e1 := NewExpectedError("ERR1", http.StatusBadRequest, "origin1")
	e2 := NewExpectedError("ERR2", http.StatusInternalServerError, "origin2")
	errs := ExpectedErrors{e1, e2}
	result := errs.GetByID("ERR1")
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), "ERR1", result.ID)

	// Test for a non-existent error ID.
	result = errs.GetByID("NON_EXISTENT")
	assert.Nil(s.T(), result)
}

func TestExpectedErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ExpectedErrorTestSuite))
}
