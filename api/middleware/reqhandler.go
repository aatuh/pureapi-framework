package middleware

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/logging"
	"github.com/aatuh/pureapi-framework/util"
	"github.com/aatuh/pureapi-framework/util/ctxutil"
)

// Context keys for various data.
var (
	responseDataKey = ctxutil.NewDataKey()
	requestDataKey  = ctxutil.NewDataKey()
	requestIDKey    = ctxutil.NewDataKey()
)

// ReqWrap defines the interface for request wrappers.
type ReqWrap interface {
	GetRequest() *http.Request
	GetBody() []byte
}

// ResWrap defines the interface for response wrappers.
type ResWrap interface {
	http.ResponseWriter
	StatusCode() int
	Header() http.Header
	WriteHeader(statusCode int)
	Body() []byte
	Write(data []byte) (int, error)
	Flush()
	Hijack() (net.Conn, *bufio.ReadWriter, error)
}

// PanicHandler defines the interface for panic handlers.
type PanicHandler interface {
	HandlePanic(w http.ResponseWriter, r *http.Request, err any)
}

// RequestLog defines the structure for request logs.
type RequestLog struct {
	StartTime     time.Time `json:"start_time"`     // When the request started.
	RemoteAddress string    `json:"remote_address"` // Client IP address.
	Protocol      string    `json:"protocol"`       // HTTP protocol (e.g., HTTP/1.1).
	HTTPMethod    string    `json:"http_method"`    // HTTP method used.
	URL           string    `json:"url"`            // Request URL.
}

// RequestMetadata defines the structure for request metadata.
type RequestMetadata struct {
	TimeStart     time.Time // Request start time.
	TraceID       string    // Unique identifier for the request.
	SpanID        string    // Span identifier for the request.
	RemoteAddress string    // Client IP address.
	Protocol      string    // HTTP protocol.
	HTTPMethod    string    // HTTP method.
	URL           string    // Request URL.
}

// ReqHandlerOptions represents options for the request handler middleware.
type ReqHandlerOptions struct {
	// Optional factory function to create a unique trace ID for the request.
	TraceIDFactoryFn func(r *http.Request) string
	// Optional factory function to create a span ID for the request.
	SpanIDFactoryFn func(r *http.Request) string
	// Optional factory function that returns a logger for request logs.
	CtxLoggerFactoryFn logging.CtxLoggerFactoryFn
	// Optional panic handler.
	PanicHandler PanicHandler
	// Optional factory function for request wrapper.
	ReqWrapFactoryFn func(r *http.Request) (ReqWrap, error)
	// Optional factory function for response wrapper.
	ResWrapFactoryFn func(w http.ResponseWriter) ResWrap
}

// ReqHandler creates a request handler middleware. The middleware handles
// multiple tasks: It injects a new context, wraps the request and response
// for inspection and reuse, attaches a trace ID and metadata, recovers from
// panics, and logs the request start andcompletion.
//
// Parameters:
//   - opts: Options for the request handler middleware.
//
// Returns:
//   - endpointtypes.Middleware: The configured request handler middleware.
func ReqHandler(
	opts *ReqHandlerOptions,
) endpoint.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxutil.NewContext(r.Context())
			r = r.WithContext(ctx)
			// Use the panic handler to recover from panics.
			defer func() {
				if rec := recover(); rec != nil {
					// Handle or re-throw the panic.
					if opts.PanicHandler != nil {
						opts.PanicHandler.HandlePanic(w, r, rec)
					} else {
						panic(rec)
					}
				}
			}()

			// Wrap response and request.
			var reqWrapper ReqWrap
			if opts.ReqWrapFactoryFn != nil {
				var err error
				reqWrapper, err = opts.ReqWrapFactoryFn(r)
				if err != nil {
					if opts.CtxLoggerFactoryFn != nil {
						opts.CtxLoggerFactoryFn(r.Context()).Error(
							"Failed to wrap request", err,
						)
					}
					http.Error(
						w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
					return
				}
				if err = setRequestWrapper(r, reqWrapper); err != nil {
					if opts.CtxLoggerFactoryFn != nil {
						opts.CtxLoggerFactoryFn(r.Context()).Error(
							"Failed to set request wrapper into context", err,
						)
					}
					http.Error(
						w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
					return
				}
			}

			var resWrapper ResWrap
			if opts.ResWrapFactoryFn != nil {
				resWrapper = opts.ResWrapFactoryFn(w)
				if err := setResponseWrapper(r, resWrapper); err != nil {
					if opts.CtxLoggerFactoryFn != nil {
						opts.CtxLoggerFactoryFn(r.Context()).Error(
							"Failed to set response wrapper into context", err,
						)
					}
					http.Error(w, http.StatusText(
						http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
					return
				}
			}

			// Create and attach trace and span metadata.
			var traceId string
			if opts.TraceIDFactoryFn != nil {
				traceId = opts.TraceIDFactoryFn(r)
			}
			var spanId string
			if opts.SpanIDFactoryFn != nil {
				spanId = opts.SpanIDFactoryFn(r)
			}

			// Create and attach request metadata.
			reqMeta := &RequestMetadata{
				TimeStart:     time.Now().UTC(),
				TraceID:       traceId,
				SpanID:        spanId,
				RemoteAddress: util.RequestIPAddress(r),
				Protocol:      r.Proto,
				HTTPMethod:    r.Method,
				URL:           fmt.Sprintf("%s%s", r.Host, r.URL.Path),
			}
			_, err := ctxutil.SetContextValue(
				r.Context(), requestIDKey, reqMeta,
			)
			if err != nil {
				if opts.CtxLoggerFactoryFn != nil {
					opts.CtxLoggerFactoryFn(r.Context()).Error(
						"Failed to set request metadata", err,
					)
				}
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}

			// Log the start of the request.
			logRequestStart(r, reqMeta, opts.CtxLoggerFactoryFn)

			// Get theresponse and request from the wrappers or the original.
			var res http.ResponseWriter
			if resWrapper != nil {
				res = resWrapper
			} else {
				res = w
			}
			var req *http.Request
			if reqWrapper != nil {
				req = reqWrapper.GetRequest()
			} else {
				req = r
			}

			// Execute the next handler.
			next.ServeHTTP(res, req)
			if opts.CtxLoggerFactoryFn != nil {
				opts.CtxLoggerFactoryFn(r.Context()).Info("Request completed")
			}
		})
	}
}

// GetRequestMetadata retrieves the request metadata from the context.
func GetRequestMetadata(ctx context.Context) *RequestMetadata {
	return ctxutil.GetContextValue[*RequestMetadata](ctx, requestIDKey, nil)
}

// setResponseWrapper saves the response wrapper in the request context.
func setResponseWrapper(r *http.Request, rw ResWrap) error {
	_, err := ctxutil.SetContextValue(r.Context(), responseDataKey, rw)
	return err
}

// setRequestWrapper saves the request wrapper in the request context.
func setRequestWrapper(r *http.Request, rw ReqWrap) error {
	_, err := ctxutil.SetContextValue(r.Context(), requestDataKey, rw)
	return err
}

// GetResponseWrapper retrieves the response wrapper from the request context.
func GetResponseWrapper(r *http.Request) ResWrap {
	return ctxutil.GetContextValue[ResWrap](
		r.Context(), responseDataKey, nil,
	)
}

// logRequestStart logs the beginning of the request.
func logRequestStart(
	r *http.Request,
	meta *RequestMetadata,
	ctxLoggerFactoryFn logging.CtxLoggerFactoryFn,
) {
	if ctxLoggerFactoryFn == nil {
		return
	}
	if meta == nil {
		ctxLoggerFactoryFn(r.Context()).Warn(
			"Request started: Request metadata not found",
		)
	} else {
		ctxLoggerFactoryFn(r.Context()).Info(
			"Request started",
			RequestLog{
				StartTime:     time.Now().UTC(),
				RemoteAddress: meta.RemoteAddress,
				Protocol:      meta.Protocol,
				HTTPMethod:    meta.HTTPMethod,
				URL:           meta.URL,
			},
		)
	}
}
