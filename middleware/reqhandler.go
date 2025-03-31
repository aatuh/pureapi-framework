package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/middleware/types"
	"github.com/pureapi/pureapi-framework/util"
	"github.com/pureapi/pureapi-framework/util/ctxutil"
)

var (
	responseDataKey = ctxutil.NewDataKey()
	requestDataKey  = ctxutil.NewDataKey()
	requestIDKey    = ctxutil.NewDataKey()
)

type requestLog struct {
	StartTime     time.Time `json:"start_time"`     // When the request started.
	RemoteAddress string    `json:"remote_address"` // Client IP address.
	Protocol      string    `json:"protocol"`       // HTTP protocol (e.g., HTTP/1.1).
	HTTPMethod    string    `json:"http_method"`    // HTTP method used.
	URL           string    `json:"url"`            // Request URL.
}

type requestMetadata struct {
	TimeStart     time.Time // Request start time.
	TraceID       string    // Unique identifier for the request.
	SpanID        string    // Span identifier for the request.
	RemoteAddress string    // Client IP address.
	Protocol      string    // HTTP protocol.
	HTTPMethod    string    // HTTP method.
	URL           string    // Request URL.
}

// ReqHandler creates a request handler middleware. The middleware handles
// multiple tasks: It injects a new context, wraps the request and response
// for inspection and reuse, attaches a trace ID and metadata, recovers from
// panics, and logs the request start andcompletion.
//
// Parameters:
//   - traceIDFn: Function to generate a unique trace ID for the request.
//   - spanIDFn: Function to generate a span ID for the request.
//   - loggerFn: Function that returns a logger for request logs.
//   - panicHandler: Panic handler.
//   - reqWrapFactoryFn: Factory function for request wrapper.
//   - resWrapFactoryFn: Factory function for response wrapper.
//
// Returns:
//   - endpointtypes.Middleware: The configured request handler middleware.
func ReqHandler(
	traceIDFn func(r *http.Request) string,
	spanIDFn func(r *http.Request) string,
	loggerFn utiltypes.CtxLoggerFactoryFn,
	panicHandler types.PanicHandler,
	reqWrapFactoryFn func(r *http.Request) (types.ReqWrap, error),
	resWrapFactoryFn func(w http.ResponseWriter) types.ResWrap,
) endpointtypes.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxutil.NewContext(r.Context())
			r = r.WithContext(ctx)
			// Use the injected PanicHandler to recover from panics.
			defer func() {
				if rec := recover(); rec != nil {
					panicHandler.HandlePanic(w, r, rec)
				}
			}()

			// Wrap response and request.
			reqWrapper, err := reqWrapFactoryFn(r)
			if err != nil {
				loggerFn(r.Context()).Error("Failed to wrap request", err)
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}

			// Store the wrappers in the request context.
			rw := resWrapFactoryFn(w)
			if err = setResponseWrapper(r, rw); err != nil {
				loggerFn(r.Context()).
					Error("Failed to set response wrapper into context", err)
				http.Error(w, http.StatusText(
					http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			if err = setRequestWrapper(r, reqWrapper); err != nil {
				loggerFn(r.Context()).
					Error("Failed to set request wrapper into context", err)
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
				SpanID:        spanIDFn(r),
				RemoteAddress: util.RequestIPAddress(r),
				Protocol:      r.Proto,
				HTTPMethod:    r.Method,
				URL:           fmt.Sprintf("%s%s", r.Host, r.URL.Path),
			}
			_, err = ctxutil.SetContextValue(r.Context(), requestIDKey, reqMeta)
			if err != nil {
				loggerFn(r.Context()).
					Error("Failed to set request metadata", err)
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}

			// Log the start of the request.
			logRequestStart(r, reqMeta, loggerFn)

			// Execute the next handler.
			next.ServeHTTP(rw, reqWrapper.GetRequest())
			loggerFn(r.Context()).Info("Request completed")
		})
	}
}

// GetRequestMetadata retrieves the request metadata from the context.
func GetRequestMetadata(ctx context.Context) *requestMetadata {
	return ctxutil.GetContextValue[*requestMetadata](ctx, requestIDKey, nil)
}

// setResponseWrapper saves the response wrapper in the request context.
func setResponseWrapper(r *http.Request, rw types.ResWrap) error {
	_, err := ctxutil.SetContextValue(r.Context(), responseDataKey, rw)
	return err
}

// setRequestWrapper saves the request wrapper in the request context.
func setRequestWrapper(r *http.Request, rw types.ReqWrap) error {
	_, err := ctxutil.SetContextValue(r.Context(), requestDataKey, rw)
	return err
}

// GetResponseWrapper retrieves the response wrapper from the request context.
func GetResponseWrapper(r *http.Request) types.ResWrap {
	return ctxutil.GetContextValue[types.ResWrap](
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
		requestLoggerFn(r.Context()).
			Warn("Request started: Request metadata not found")
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
