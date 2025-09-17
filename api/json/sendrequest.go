package json

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aatuh/pureapi-core/logging"
	"github.com/aatuh/pureapi-framework/defaults"
	"github.com/aatuh/pureapi-framework/util"
	"github.com/aatuh/urlcodec"
)

// Constants for HTTP headers and content types.
const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

// Response represents the response from a client request.
type Response[Output any] struct {
	Response *http.Response // The HTTP response object.
	Output   *Output        // The output data of the API response.
}

// RequestOption defines a function type that modifies a RequestData.
type RequestOption func(*util.RequestData)

// WithSpanID is an option that adds an X-Span-ID header if not already present.
//
// Returns:
//   - RequestOption: The request option.
func WithSpanID() RequestOption {
	return func(rd *util.RequestData) {
		if rd.Headers == nil {
			rd.Headers = make(map[string]string)
		}
		if _, exists := rd.Headers[defaults.XSpanID]; !exists {
			rd.Headers[defaults.XSpanID] = defaults.GenerateSpanID()
		}
	}
}

// WithoutSpanID is an option that ensures the X-Span-ID header is not set.
//
// Returns:
//   - RequestOption: The request option.
func WithoutSpanID() RequestOption {
	return func(rd *util.RequestData) {
		if rd.Headers != nil {
			delete(rd.Headers, defaults.XSpanID)
		}
	}
}

// SendRequest sends a request to the target host. It will setup the input
// headers, cookies, and body based on the input. By default, WithSpanID is
// applied, unless overridden.
//
// Parameters:
//   - ctx: The context for the request.
//   - host: The host to send the request to.
//   - url: The URL to send the request to.
//   - method: The HTTP method to use for the request.
//   - i: The input to use for the request.
//   - loggerFactoryFn: An optional function that returns a logger.
//   - opts: Optional request options. By default, WithSpanID is applied.
//
// Returns:
//   - *Response[Output]: The response from the request.
//   - error: An error if the request fails.
func SendRequest[Input any, Output any](
	ctx context.Context,
	host string,
	url string,
	method string,
	i *Input,
	loggerFactoryFn logging.CtxLoggerFactoryFn,
	opts ...RequestOption,
) (*Response[Output], error) {
	// Parse the input.
	parsedInput, err := util.ParseInput(method, i)
	if err != nil {
		return nil, err
	}
	var logger logging.ILogger
	if loggerFactoryFn != nil {
		logger = loggerFactoryFn(ctx)
	}
	if logger != nil {
		logger.Trace(
			fmt.Sprintf("Send %s request: %s", method, url), parsedInput,
		)
	}

	// Apply default option to add span id.
	WithSpanID()(parsedInput)
	for _, opt := range opts {
		opt(parsedInput)
	}

	// Send the request.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the full URL with query parameters.
	fullURL := fmt.Sprintf("%s%s", host, url)
	response, err := SendParsedRequest[Output](
		ctx, client, fullURL, method, parsedInput,
	)
	if err != nil && logger != nil {
		defaults.CtxLogger(ctx).Error(
			fmt.Sprintf(
				"Failed to send %s request to %s: %s", method, url, err.Error(),
			),
			parsedInput,
		)
	}
	if logger != nil {
		defaults.CtxLogger(ctx).Trace(
			fmt.Sprintf("Received %s response: %s", method, url),
			response.Output,
		)
	}

	return response, err
}

// SendParsedRequest sends a request to the target host using the parsed input.
// It will setup the input headers, cookies, and body based on the parsed input.
// Body is only set for non-GET requests. The URL parameters are added to the
// URL forming the full URL with query parameters.
//
// Parameters:
//   - ctx: The context for the request.
//   - client: The client to use for the request.
//   - fullURL: The URL with query parameters to send the request to.
//   - method: The HTTP method to use for the request.
//   - parsedInput: The parsed input to use for the request.
//
// Returns:
//   - *Response[Output]: The response from the request.
//   - error: An error if the request fails.
func SendParsedRequest[Output any](
	ctx context.Context,
	client *http.Client,
	fullURL string,
	method string,
	parsedInput *util.RequestData,
) (*Response[Output], error) {
	// Ensure the body is not set for GET requests.
	if parsedInput.Body != nil && method == http.MethodGet {
		return nil, fmt.Errorf("SendRequest: GET request must not have body")
	}

	// Marshal the body into a JSON reader.
	bodyReader, err := marshalBody(parsedInput.Body)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: body marshal error: %v", err)
	}

	// Set the Content-Type header if it is not set.
	if parsedInput.Headers != nil &&
		parsedInput.Headers[contentTypeHeader] == "" {
		parsedInput.Headers[contentTypeHeader] = applicationJSON
	}

	// If URL parameters exist, encode and append them.
	if parsedInput.URLParameters != nil {
		params, err := urlcodec.NewURLEncoder().
			Encode(parsedInput.URLParameters)
		if err != nil {
			return nil, err
		}
		encoded := params.Encode()
		if encoded != "" {
			fullURL = fmt.Sprintf("%s?%s", fullURL, encoded)
		}
	}

	// Create the request.
	req, err := createRequest(
		ctx,
		method,
		fullURL,
		bodyReader,
		parsedInput.Headers,
		parsedInput.Cookies,
	)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: error creating request: %v", err)
	}

	// Send the request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: request error: %v", err)
	}

	// Unmarshal the response body into the output object.
	output, err := responseToPayload(resp, new(Output))
	if err != nil {
		return nil, fmt.Errorf("SendRequest: response unmarshal error: %v", err)
	}
	return &Response[Output]{
		Response: resp,
		Output:   output,
	}, nil
}

// createRequest creates a new request with the specified request data.
func createRequest(
	ctx context.Context,
	method string,
	url string,
	bodyReader io.Reader,
	headers map[string]string,
	cookies []http.Cookie,
) (*http.Request, error) {
	var body io.Reader
	if bodyReader != nil {
		body = bodyReader
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("createRequest: error creating request: %v", err)
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}
	return req, nil
}

// marshalBody marshals the body into a JSON reader.
func marshalBody(body any) (*bytes.Reader, error) {
	if body == nil {
		return bytes.NewReader(nil), nil
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshalBody: JSON marshal error: %v", err)
	}
	return bytes.NewReader(bodyBytes), nil
}

// responseToPayload unmarshals the response body into the output object.
func responseToPayload[T any](r *http.Response, output *T) (*T, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("responseToPayload: read error: %v", err)
	}

	// Handle non-JSON responses.
	ct := r.Header.Get(contentTypeHeader)
	if !strings.HasPrefix(ct, applicationJSON) {
		return nil, fmt.Errorf(
			"responseToPayload: expected JSON response, got Content-Type: %s, body: %s",
			ct,
			string(body),
		)
	}

	// For JSON responses, attempt to unmarshal.
	if err := json.Unmarshal(body, output); err != nil {
		return nil, fmt.Errorf(
			"responseToPayload: JSON unmarshal error: %v, body: %s",
			err,
			string(body),
		)
	}
	return output, nil
}
