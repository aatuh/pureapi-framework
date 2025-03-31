package jsonapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pureapi/pureapi-core/util"
	"github.com/pureapi/pureapi-core/util/types"
)

const (
	// EventJSONHandleError event is emitted when an error occurs during JSON
	// output handling.
	EventJSONHandleError = "event_json_handle_error"

	// EventOutput event is emitted when an output is emitted.
	EventOutput = "event_client_output"
)

// InvalidOutputErrorType represents a generic error.
var InvalidOutputErrorType = util.NewAPIError("INVALID_OUTPUT_ERROR_TYPE")

// APIOutput represents the output of a client request.
type APIOutput[T any] struct {
	Payload *T                    `json:"payload,omitempty"`
	Error   *util.DefaultAPIError `json:"error,omitempty"`
}

// JSONOutput represents the output of a client request.
type JSONOutput struct {
	emitterLogger types.EmitterLogger
	errorOrigin   string
}

// NewJSONOutput returns a new JSONOutput.
//
// Parameters:
//   - emitterLogger: The emitter logger.
//   - errorOrigin: The origin of the error.
//
// Returns:
//   - JSONOutput: The new JSONOutput.
func NewJSONOutput(
	emitterLogger types.EmitterLogger, errorOrigin string,
) JSONOutput {
	if emitterLogger == nil {
		emitterLogger = util.NewNoopEmitterLogger()
	}
	return JSONOutput{
		emitterLogger: emitterLogger,
		errorOrigin:   errorOrigin,
	}
}

// Handle marshals the output to JSON and writes it to the response.
// If the logger is not nil, it will log the output. Only APIErrors are allowed
// to be marshaled to JSON.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//   - out: The output data.
//   - outError: The output error.
//   - status: The HTTP status code.
//
// Returns:
//   - error: An error if the request fails.
func (o JSONOutput) Handle(
	w http.ResponseWriter, r *http.Request, out any, outError error, status int,
) error {
	output, err := o.jsonOutput(w, out, outError, status)
	if err != nil {
		o.emitterLogger.Error(
			types.NewEvent(
				EventJSONHandleError,
				fmt.Sprintf("Error handling output JSON: %+v", err),
			).WithData(map[string]any{"err": err}),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	o.emitterLogger.Trace(
		types.NewEvent(
			EventOutput,
			fmt.Sprintf("Output: %+v", out),
		).WithData(map[string]any{"output": output}),
	)
	return nil
}

// jsonOutput marshals the output to JSON and writes it to the response.
func (o JSONOutput) jsonOutput(
	w http.ResponseWriter, outputData any, outputError error, statusCode int,
) (*APIOutput[any], error) {
	// Marshal output to JSON.
	output := APIOutput[any]{
		Payload: &outputData,
		Error:   o.handleError(outputError),
	}
	jsonData, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	// Set content type and status code.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Write error to response.
	_, err = w.Write(jsonData)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

// handleError returns the API error if it is an *util.DefaultAPIError or
// an utiltypes.util. Returns InvalidOutputErrorType otherwise.
func (o JSONOutput) handleError(outputError error) *util.DefaultAPIError {
	if outputError == nil {
		return nil
	}
	if val, ok := outputError.(*util.DefaultAPIError); ok {
		return val
	}
	if val, ok := outputError.(types.APIError); ok {
		return util.APIErrorFrom(val)
	}
	return InvalidOutputErrorType
}
