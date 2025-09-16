package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/aatuh/pureapi-core/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// dummyEmitterLogger is a simple emitter logger that records events.
type dummyEmitterLogger struct {
	events []string
}

func (d *dummyEmitterLogger) Debug(e *event.Event, params ...any) {
	d.events = append(d.events, "Debug:"+e.Message)
}
func (d *dummyEmitterLogger) Trace(e *event.Event, params ...any) {
	d.events = append(d.events, "Trace:"+e.Message)
}
func (d *dummyEmitterLogger) Info(e *event.Event, params ...any) {
	d.events = append(d.events, "Info:"+e.Message)
}
func (d *dummyEmitterLogger) Warn(e *event.Event, params ...any) {
	d.events = append(d.events, "Warn:"+e.Message)
}
func (d *dummyEmitterLogger) Error(e *event.Event, params ...any) {
	d.events = append(d.events, "Error:"+e.Message)
}
func (d *dummyEmitterLogger) Fatal(e *event.Event, params ...any) {
	d.events = append(d.events, "Fatal:"+e.Message)
}

// failResponseWriter simulates a ResponseWriter that fails on Write.
type failResponseWriter struct{}

func (f *failResponseWriter) Header() http.Header {
	return http.Header{}
}
func (f *failResponseWriter) Write(b []byte) (int, error) {
	return 0, fmt.Errorf("write failure")
}
func (f *failResponseWriter) WriteHeader(statusCode int) {}

// JSONOutputTestSuite is a test suite for the JSONOutput.
type JSONOutputTestSuite struct {
	suite.Suite
	logger dummyEmitterLogger
	output JSONOutput
}

// TestJSONOutputTestSuite runs the test suite.
func TestJSONOutputTestSuite(t *testing.T) {
	suite.Run(t, new(JSONOutputTestSuite))
}

// SetupTest sets up the test suite.
func (s *JSONOutputTestSuite) SetupTest() {
	s.output = NewJSONOutput(&s.logger, "test_origin")
}

// TestHandle_Success verifies that a valid output is correctly marshaled.
func (s *JSONOutputTestSuite) TestHandle_Success() {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	outData := "Success"
	err := s.output.Handle(rr, req, outData, nil, http.StatusOK)
	require.NoError(s.T(), err)
	// Verify status code and content type.
	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Equal(s.T(), "application/json", rr.Header().Get("Content-Type"))
	// Unmarshal response.
	var apiOut APIOutput[any]
	err = json.Unmarshal(rr.Body.Bytes(), &apiOut)
	require.NoError(s.T(), err)
	// Check that the payload equals "Success" and error is nil.
	require.NotNil(s.T(), apiOut.Payload)
	payload, ok := (*apiOut.Payload).(string)
	require.True(s.T(), ok, "Expected payload to be a string")
	assert.Equal(s.T(), "Success", payload)
	assert.Nil(s.T(), apiOut.Error)
}

// TestHandle_WithAPIError verifies that when an API error is provided, it
// appears in the JSON output.
func (s *JSONOutputTestSuite) TestHandle_WithAPIError() {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	// Create a sample API error.
	apiErr := apierror.NewAPIError("SAMPLE_ERROR").WithData("error detail")
	outData := "Output Data"
	err := s.output.Handle(rr, req, outData, apiErr, http.StatusBadRequest)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	// Unmarshal response.
	var apiOut APIOutput[any]
	err = json.Unmarshal(rr.Body.Bytes(), &apiOut)
	require.NoError(s.T(), err)
	// Verify payload.
	require.NotNil(s.T(), apiOut.Payload)
	payload, ok := (*apiOut.Payload).(string)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "Output Data", payload)
	// Verify that the error is present and its ID is SAMPLE_ERROR.
	require.NotNil(s.T(), apiOut.Error)
	assert.Equal(s.T(), "SAMPLE_ERROR", apiOut.Error.ID())
}

// TestHandle_NonMarshalableOutput verifies that if the output cannot be
// marshaled, an error is returned.
func (s *JSONOutputTestSuite) TestHandle_NonMarshalableOutput() {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	// Create an output that is not marshalable (e.g., a channel).
	nonMarshalable := make(chan int)
	err := s.output.Handle(rr, req, nonMarshalable, nil, http.StatusOK)
	require.Error(s.T(), err)
	// Expect that an internal server error is written.
	assert.Equal(s.T(), http.StatusInternalServerError, rr.Code)
}

// TestHandle_WriteError simulates a failure when writing to the ResponseWriter.
func (s *JSONOutputTestSuite) TestHandle_WriteError() {
	rw := &failResponseWriter{}
	req := httptest.NewRequest("GET", "http://example.com", nil)
	err := s.output.Handle(rw, req, "Data", nil, http.StatusOK)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "write failure")
}

// TestHandle_NonAPIError verifies that if a non-API error is provided, the
// InvalidOutputErrorType is used.
func (s *JSONOutputTestSuite) TestHandle_NonAPIError() {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com", nil)
	// Provide a non-API error.
	nonAPIErr := errors.New("some error")
	err := s.output.Handle(rr, req, "Data", nonAPIErr, http.StatusBadRequest)
	require.NoError(s.T(), err)
	// Unmarshal response.
	var apiOut APIOutput[any]
	err = json.Unmarshal(rr.Body.Bytes(), &apiOut)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), apiOut.Error)
	assert.Equal(s.T(), InvalidOutputErrorType.ID(), apiOut.Error.ID())
}
