package errutil

import (
	"errors"
	"net/http"
	"testing"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestErrorHandlerTestSuite is a test suite for the error handler.
func TestErrorHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlerTestSuite))
}

// ErrorHandlerTestSuite runs the test suite.
type ErrorHandlerTestSuite struct {
	suite.Suite
}

// errorHandlerTestCase defines a table-driven test case for the error handler.
type errorHandlerTestCase struct {
	name           string
	expectedErrs   []ExpectedError
	systemID       *string // optional system ID to set on the handler
	inputError     error
	expectedStatus int
	expectedID     string
	expectedOrigin string
	expectedData   any
}

// TestHandle tests the Handle method of the error handler.
func (s *ErrorHandlerTestSuite) TestHandle() {
	systemID := "systemX"
	testCases := []errorHandlerTestCase{
		{
			name:           "non-api error",
			expectedErrs:   []ExpectedError{},
			inputError:     errors.New("plain error"),
			expectedStatus: http.StatusInternalServerError,
			// For non-api errors don’t inspect individual fields.
		},
		{
			name: "expected error match with origin preset",
			expectedErrs: []ExpectedError{
				{
					ID:         "SAMPLE_ERROR",
					MaskedID:   "MASKED_SAMPLE_ERROR",
					Status:     http.StatusBadRequest,
					PublicData: true,
					Origin:     "test_origin",
				},
			},
			inputError: apierror.NewAPIError("SAMPLE_ERROR").
				WithOrigin("test_origin").
				WithData("some data"),
			expectedStatus: http.StatusBadRequest,
			expectedID:     "MASKED_SAMPLE_ERROR",
			expectedOrigin: "test_origin",
			expectedData:   "some data",
		},
		{
			name: "expected error match with systemID applied",
			expectedErrs: []ExpectedError{
				{
					ID:         "SAMPLE_ERROR",
					MaskedID:   "MASKED_SAMPLE_ERROR",
					Status:     http.StatusBadRequest,
					PublicData: true,
					Origin:     "systemX",
				},
			},
			systemID: &systemID,
			inputError: apierror.NewAPIError("SAMPLE_ERROR").
				WithData("info"), // no origin provided
			expectedStatus: http.StatusBadRequest,
			expectedID:     "MASKED_SAMPLE_ERROR",
			expectedOrigin: "systemX",
			expectedData:   "info",
		},
		{
			name: "expected error match with public data false",
			expectedErrs: []ExpectedError{
				{
					ID:         "SAMPLE_ERROR",
					MaskedID:   "SAMPLE_ERROR", // no override
					Status:     http.StatusBadRequest,
					PublicData: false,
					Origin:     "test_origin",
				},
			},
			inputError: apierror.NewAPIError("SAMPLE_ERROR").
				WithOrigin("test_origin").
				WithData("secret"),
			expectedStatus: http.StatusBadRequest,
			expectedID:     "SAMPLE_ERROR",
			expectedOrigin: "test_origin",
			expectedData:   nil,
		},
		{
			name: "no matching expected error",
			expectedErrs: []ExpectedError{
				{
					ID:         "OTHER_ERROR",
					MaskedID:   "OTHER_ERROR",
					Status:     http.StatusBadRequest,
					PublicData: true,
					Origin:     "other",
				},
			},
			inputError: apierror.NewAPIError("SAMPLE_ERROR").
				WithOrigin("test_origin").
				WithData("data"),
			expectedStatus: http.StatusInternalServerError,
			expectedID:     "INTERNAL_SERVER_ERROR",
			expectedOrigin: "",
			expectedData:   nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			handler := NewErrorHandler(tc.expectedErrs)
			if tc.systemID != nil {
				handler = handler.WithSystemID(tc.systemID)
			}
			status, outErr := handler.Handle(tc.inputError)
			assert.Equal(s.T(), tc.expectedStatus, status, tc.name)
			// For non-api errors, compare to InternalServerError.
			if tc.name == "non-api error" {
				assert.Equal(s.T(), InternalServerError, outErr, tc.name)
			} else {
				assert.Equal(s.T(), tc.expectedID, outErr.ID(), tc.name)
				assert.Equal(s.T(), tc.expectedOrigin, outErr.Origin(), tc.name)
				assert.Equal(s.T(), tc.expectedData, outErr.Data(), tc.name)
			}
		})
	}
}
