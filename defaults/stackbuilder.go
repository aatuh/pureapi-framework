package defaults

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/middleware"
	"github.com/pureapi/pureapi-framework/middleware/types"
	"github.com/pureapi/pureapi-framework/util/httpwrap"
	"github.com/pureapi/pureapi-framework/util/paniclog"
	"github.com/pureapi/pureapi-util/envvar"
)

const (
	// Middleware wrapper IDs.
	CORSID       = "cors"
	ReqHandlerID = "reqhandler"

	// HTTP headers.
	XSpanID = "X-Span-ID"

	// Defaults for request handling.
	maxRequestBodySize   = 10 * 1024 * 1024 // 10 MB
	maxPanicDumpPartSize = 10 * 1024 * 1024 // 10 MB
)

// StackEnvConfig holds the names of the environment variables used.
type StackEnvConfig struct {
	AlwaysNewSpanID string
}

// StackConfig returns the stack configuration.
type StackConfig struct {
	AlwaysNewSpan bool
}

// stackEnvConfig holds the stack environment variable config.
var stackEnvConfig atomic.Value

func init() {
	stackEnvConfig.Store(StackEnvConfig{
		AlwaysNewSpanID: "ALWAYS_NEW_SPAN_ID",
	})
}

// SetStackEnvConfig allows clients to override the default stack env var names.
//
// Parameters:
//   - cfg: The stack environment variable names.
func SetStackEnvConfig(cfg StackEnvConfig) {
	stackEnvConfig.Store(cfg)
}

// GetStackEnvConfig returns the current stack env var configuration.
//
// Returns:
//   - StackEnvConfig: The stack environment variable configuration.
func GetStackEnvConfig() StackEnvConfig {
	return stackEnvConfig.Load().(StackEnvConfig)
}

// NewStackConfig returns a default stack config.
//
// Returns:
//   - *StackConfig: The stack config.
func NewStackConfig() *StackConfig {
	return &StackConfig{
		AlwaysNewSpan: envvar.MustGetBool(GetStackEnvConfig().AlwaysNewSpanID),
	}
}

// StackBuilder builds a middleware stack.
type StackBuilder struct {
	stack endpointtypes.Stack
}

// NewStackBuilder returns a default stack builder.
//
// Returns:
//   - *StackBuilder: The stack builder.
func NewStackBuilder() *StackBuilder {
	return &StackBuilder{
		stack: endpoint.NewStack(
			ReqHandlerWrapper(
				func(r *http.Request) string {
					return (&UUIDGen{}).MustRandom().String()
				},
				extractOrGenerateSpanID,
				func(ctx context.Context) utiltypes.ILogger {
					return CtxLogger(ctx)
				},
			),
		),
	}
}

// Build returns a new stack from the builder.
//
// Returns:
//   - endpointtypes.Stack: The new stack.
func (b *StackBuilder) Build() endpointtypes.Stack {
	newStack := b.stack.Clone()
	return newStack
}

// Clone returns a copy of the stack builder.
//
// Returns:
//   - *StackBuilder: The cloned stack builder.
func (b *StackBuilder) Clone() *StackBuilder {
	return &StackBuilder{
		stack: b.stack.Clone(),
	}
}

// CORSWrapper creates a new Wrapper for the CORS endpoint.
//
// Parameters:
//   - opts: The CORS options.
//
// Returns:
//   - endpointtypes.Wrapper: The CORS wrapper.
func CORSWrapper(opts *middleware.CORSOptions) endpointtypes.Wrapper {
	return endpoint.NewWrapper(CORSID, middleware.CORS(*opts))
}

// ReqHandlerWrapper creates a new Wrapper for the reqhandler endpoint.
//
// Parameters:
//   - traceIDFn: The trace ID function.
//   - spanIDFn: The span ID function.
//   - loggerFactoryFn: The logger factory function.
//
// Returns:
//   - endpointtypes.Wrapper: The reqhandler wrapper.
func ReqHandlerWrapper(
	traceIDFn func(r *http.Request) string,
	spanIDFn func(r *http.Request) string,
	loggerFactoryFn utiltypes.CtxLoggerFactoryFn,
) endpointtypes.Wrapper {
	return endpoint.NewWrapper(
		ReqHandlerID,
		middleware.ReqHandler(
			traceIDFn,
			spanIDFn,
			loggerFactoryFn,
			paniclog.NewPanicLog(
				loggerFactoryFn,
				maxPanicDumpPartSize,
				func(r *http.Request) httpwrap.ResWrap {
					return middleware.GetResponseWrapper(r)
				},
			),
			func(r *http.Request) (types.ReqWrap, error) {
				return httpwrap.NewReqWrap(r, maxRequestBodySize)
			},
			func(w http.ResponseWriter) types.ResWrap {
				return httpwrap.NewResWrap(w)
			},
		),
	)
}

// GenerateSpanID generates a new span ID.
//
// Returns:
//   - string: The new span ID.
func GenerateSpanID() string {
	return (&UUIDGen{}).MustRandom().String()
}

// extractOrGenerateSpanID checks for a span ID in the request header and
// returns it if available, or generates a new one.
func extractOrGenerateSpanID(r *http.Request) string {
	cfg := NewStackConfig()
	if cfg.AlwaysNewSpan {
		return GenerateSpanID()
	}
	if span := r.Header.Get(XSpanID); span != "" {
		return span
	}
	return GenerateSpanID()
}
