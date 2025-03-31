package jsonapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pureapi/pureapi-framework/defaults"
	"github.com/pureapi/pureapi-framework/util"
	"github.com/pureapi/pureapi-util/urlencoder"
)

// SendRequest sends a request to the target host. It will setup the input
// headers, cookies, and body based on the input.
//
// Parameters:
//   - ctx: The context for the request.
//   - host: The host to send the request to.
//   - url: The URL to send the request to.
//   - method: The HTTP method to use for the request.
//   - i: The input to use for the request.
//
// Returns:
//   - *Response[Output]: The response from the request.
//   - error: An error if the request fails.
func SendRequest[Input any, Output any](
	ctx context.Context, host string, url string, method string, i *Input,
) (*Response[Output], error) {
	// Parse the input.
	parsedInput, err := util.ParseInput(method, i)
	if err != nil {
		return nil, err
	}
	defaults.CtxLogger(ctx).Trace(
		fmt.Sprintf("Send %s request: %s", method, url), parsedInput,
	)
	// Send the request.
	httpClient := NewJSONClient[Output](host, 0)
	response, err := SendParsedRequest(
		ctx, httpClient, url, method, parsedInput,
	)
	if err != nil {
		defaults.CtxLogger(ctx).Error(
			fmt.Sprintf(
				"Failed to send %s request to %s: %s", method, url, err.Error(),
			),
			parsedInput,
		)
	}
	defaults.CtxLogger(ctx).Trace(
		fmt.Sprintf("Received %s response: %s", method, url), response.Output,
	)
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
//   - url: The URL to send the request to.
//   - method: The HTTP method to use for the request.
//   - parsedInput: The parsed input to use for the request.
//
// Returns:
//   - *Response[Output]: The response from the request.
//   - error: An error if the request fails.
func SendParsedRequest[Output any](
	ctx context.Context,
	client *JSONClient[Output],
	url string,
	method string,
	parsedInput *util.RequestData,
) (*Response[Output], error) {
	// Ensure headers map is initialized.
	if parsedInput.Headers == nil {
		parsedInput.Headers = make(map[string]string)
	}
	// Add span id header if not already set.
	if _, exists := parsedInput.Headers[defaults.XSpanID]; !exists {
		parsedInput.Headers[defaults.XSpanID] = defaults.GenerateSpanID()
	}

	// Set options.
	opts := &SendOptions{
		Headers: parsedInput.Headers,
		Cookies: parsedInput.Cookies,
	}
	if method != http.MethodGet {
		opts.Body = parsedInput.Body
	}
	// Add query parameters to the URL.
	urlValues, err := urlencoder.NewURLEncoder().
		Encode(parsedInput.URLParameters)
	if err != nil {
		return nil, err
	}
	// Add query parameters to the URL.
	var fullURL string
	if len(urlValues) == 0 {
		fullURL = url
	} else {
		fullURL = fmt.Sprintf("%s?%s", url, urlValues.Encode())
	}
	// Send request.
	return client.Send(ctx, fullURL, method, opts)
}
