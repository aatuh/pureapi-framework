package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pureapi/pureapi-framework/custom"
	"github.com/pureapi/pureapi-framework/input"
	"github.com/pureapi/pureapi-framework/json"
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
//   - *json.Response[json.APIOutput[Output]]: The response from the request.
//   - error: An error if the request fails.
func SendRequest[Input any, Output any](
	ctx context.Context, host string, url string, method string, i *Input,
) (*json.Response[json.APIOutput[Output]], error) {
	// Parse the input.
	parsedInput, err := ParseInput(method, i)
	if err != nil {
		return nil, err
	}
	custom.CtxLogger(ctx).Trace(
		fmt.Sprintf("Send %s request: %s", method, url), parsedInput,
	)
	// Send the request.
	httpClient := json.NewJSONClient[json.APIOutput[Output]](host, 0)
	response, err := SendParsedRequest(
		ctx, httpClient, url, method, parsedInput,
	)
	if err != nil {
		custom.CtxLogger(ctx).Error(
			fmt.Sprintf(
				"Failed to send %s request to %s: %s", method, url, err.Error(),
			),
			parsedInput,
		)
	}
	custom.CtxLogger(ctx).Trace(
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
//   - httpClient: The HTTP client to use for the request.
//   - url: The URL to send the request to.
//   - method: The HTTP method to use for the request.
//   - parsedInput: The parsed input to use for the request.
//
// Returns:
//   - *json.Response[json.APIOutput[Output]]: The response from the request.
//   - error: An error if the request fails.
func SendParsedRequest[Output any](
	ctx context.Context,
	httpClient *json.JSONClient[json.APIOutput[Output]],
	url string,
	method string,
	parsedInput *RequestData,
) (*json.Response[json.APIOutput[Output]], error) {
	// Set options.
	opts := &json.SendOptions{
		Headers: parsedInput.Headers,
		Cookies: parsedInput.Cookies,
	}
	if method != http.MethodGet {
		opts.Body = parsedInput.Body
	}
	// Add query parameters to the URL.
	urlValues, err := input.NewURLEncoder().Encode(parsedInput.URLParameters)
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
	return httpClient.Send(ctx, fullURL, method, opts)
}
