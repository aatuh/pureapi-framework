package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/util"
)

// TODO: Span ID

var (
	// These keys are used to store values in the request context.
	responseDataKey = util.NewDataKey()
	requestDataKey  = util.NewDataKey()
	requestIDKey    = util.NewDataKey()
)

type requestLog struct {
	StartTime     time.Time `json:"start_time"`     // Time when the request started.
	RemoteAddress string    `json:"remote_address"` // Client IP address.
	Protocol      string    `json:"protocol"`       // HTTP protocol (e.g., HTTP/1.1).
	HTTPMethod    string    `json:"http_method"`    // HTTP method used.
	URL           string    `json:"url"`            // Request URL.
}

type requestMetadata struct {
	TimeStart     time.Time // Request start time.
	TraceID       string    // Unique identifier for the request.
	RemoteAddress string    // Client IP address.
	Protocol      string    // HTTP protocol.
	HTTPMethod    string    // HTTP method.
	URL           string    // Request URL.
}

// ReqHandler creates a request handler middleware that provides the following:
//   - Injects a new context into the request.
//   - Wraps the response writer and request for inspection.
//   - Attaches a unique trace ID and additional metadata to the context.
//   - Recovers from panics and logs detailed request/response data along with a
//     stack trace.
//   - Logs the start and completion of the request.
//
// Parameters:
//   - maxRequestBodySize: Maximum size of the request body in bytes.
//   - maxPanicDumpPartSize: Maximum size of each panic dump part in bytes.
//   - traceIDFn: Function to generate a unique trace ID for the request.
//   - loggerFn: Function that returns a logger for request logs.
//
// Returns:
//   - api.Middleware: The configured request handler middleware.
func ReqHandler(
	maxRequestBodySize int64,
	maxPanicDumpPartSize int64,
	traceIDFn func(r *http.Request) string,
	loggerFn utiltypes.CtxLoggerFactoryFn,
) endpointtypes.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Attach a new context.
			ctx := util.NewContext(r.Context())
			r = r.WithContext(ctx)
			// Panic recovery.
			defer func() {
				if rec := recover(); rec != nil {
					util.NewPanicHandler(
						loggerFn,
						maxPanicDumpPartSize,
						getResponseWrapper,
					).HandlePanic(w, r, rec)
				}
			}()
			// Wrap response and request.
			rw := util.NewResWrap(w)
			reqWrapper, err := util.NewReqWrap(r, maxRequestBodySize)
			if err != nil {
				loggerFn(r.Context()).Error("Failed to wrap request", err)
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			// Store wrappers in the request context.
			err = setResponseWrapper(r, rw)
			if err != nil {
				loggerFn(r.Context()).Error(
					"Failed to set response wrapper into context", err,
				)
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			err = setRequestWrapper(r, reqWrapper)
			if err != nil {
				loggerFn(r.Context()).Error(
					"Failed to set requst wrapper into context", err,
				)
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			// Create and attach request metadata.
			reqMeta := &requestMetadata{
				TimeStart:     time.Now().UTC(),
				TraceID:       traceIDFn(r),
				RemoteAddress: util.RequestIPAddress(r),
				Protocol:      r.Proto,
				HTTPMethod:    r.Method,
				URL:           fmt.Sprintf("%s%s", r.Host, r.URL.Path),
			}
			_, err = util.SetContextValue(r.Context(), requestIDKey, reqMeta)
			if err != nil {
				loggerFn(r.Context()).Error(
					"Failed to set request metadata", err,
				)
				http.Error(
					w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			// Log request start.
			logRequestStart(r, reqMeta, loggerFn)
			// Execute the next handler.
			next.ServeHTTP(rw, reqWrapper.GetRequest())
			// Log request completion.
			loggerFn(r.Context()).Info("Request completed")
		})
	}
}

// GetRequestMetadata retrieves the request metadata from the context.
//
// Parameters:
//   - ctx: The request context.
//
// Returns:
//   - *requestMetadata: The request metadata, or nil if not found.
func GetRequestMetadata(ctx context.Context) *requestMetadata {
	return util.GetContextValue[*requestMetadata](ctx, requestIDKey, nil)
}

// setResponseWrapper saves the response wrapper in the request context.
func setResponseWrapper(r *http.Request, rw util.ResWrap) error {
	_, err := util.SetContextValue(r.Context(), responseDataKey, rw)
	return err
}

// setRequestWrapper saves the request wrapper in the request context.
func setRequestWrapper(r *http.Request, rw util.ReqWrap) error {
	_, err := util.SetContextValue(r.Context(), requestDataKey, rw)
	return err
}

// getResponseWrapper retrieves the response wrapper from the request context.
func getResponseWrapper(r *http.Request) util.ResWrap {
	return util.GetContextValue[util.ResWrap](
		r.Context(), responseDataKey, nil,
	)
}

// logRequestStart logs the beginning of the request.
func logRequestStart(
	r *http.Request,
	meta *requestMetadata,
	requestLoggerFn utiltypes.CtxLoggerFactoryFn,
) {
	if meta == nil {
		requestLoggerFn(r.Context()).Info(
			"Request started", "Request metadata not found",
		)
		return
	}
	requestLoggerFn(r.Context()).Info(
		"Request started",
		requestLog{
			StartTime:     time.Now().UTC(),
			RemoteAddress: meta.RemoteAddress,
			Protocol:      meta.Protocol,
			HTTPMethod:    meta.HTTPMethod,
			URL:           meta.URL,
		},
	)
}
