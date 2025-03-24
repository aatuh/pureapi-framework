package defaults

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pureapi/pureapi-core/endpoint"
	endpointtypes "github.com/pureapi/pureapi-core/endpoint/types"
	utiltypes "github.com/pureapi/pureapi-core/util/types"
	"github.com/pureapi/pureapi-framework/custom"
	"github.com/pureapi/pureapi-framework/middleware"
)

const (
	CORSID       = "cors"
	ReqHandlerID = "reqhandler"

	maxRequestBodySize   = 10 * 1024 * 1024 // 10 MB
	maxPanicDumpPartSize = 10 * 1024 * 1024 // 10 MB
)

// CORSWrapper creates a new Wrapper for the CORS endpoint.
func CORSWrapper(
	allowedOrigins []string, allowedMethods []string, allowedHeaders []string,
) endpointtypes.Wrapper {
	return endpoint.NewWrapper(
		CORSID,
		middleware.CORS(
			middleware.CORSOptions{
				AllowedOrigins:   allowedOrigins,
				AllowedMethods:   allowedMethods,
				AllowedHeaders:   allowedHeaders,
				AllowCredentials: true,
			},
		),
	)
}

// ReqHandlerWrapper creates a new Wrapper for the reqhandler endpoint.
func ReqHandlerWrapper(
	traceIDFn func(r *http.Request) string,
	panicHandlerLoggerFn utiltypes.CtxLoggerFactoryFn,
) endpointtypes.Wrapper {
	return endpoint.NewWrapper(
		ReqHandlerID,
		middleware.ReqHandler(
			maxRequestBodySize,
			maxPanicDumpPartSize,
			traceIDFn,
			panicHandlerLoggerFn,
		),
	)
}

// StackBuilder builds a middleware stack.
type StackBuilder struct {
	stack endpointtypes.Stack
}

// Build returns a new stack from the builder.
func (b *StackBuilder) Build() endpointtypes.Stack {
	newStack := b.stack.Clone()
	return newStack
}

// Clone returns a copy of the stack builder.
func (b *StackBuilder) Clone() *StackBuilder {
	return &StackBuilder{
		stack: b.stack.Clone(),
	}
}

// MustAddMiddleware adds middleware to the stack and panics if it fails.
func (b *StackBuilder) MustAddMiddleware(
	wrapper ...endpointtypes.Wrapper,
) *StackBuilder {
	for i := range wrapper {
		stack, success := b.stack.InsertAfter(
			ReqHandlerID, wrapper[i],
		)
		if !success {
			panic(fmt.Sprintf("Failed to add middleware: %s", wrapper[i].ID()))
		}
		b.stack = stack
	}
	return b
}

// DefaultStackBuilder returns a default stack builder.
func DefaultStackBuilder() *StackBuilder {
	return newStackBuilder(
		func(r *http.Request) string {
			return (&custom.UUIDGen{}).MustRandom().String()
		},
		func(ctx context.Context) utiltypes.ILogger {
			return custom.CtxLogger(ctx)
		},
	)
}

// NewStackBuilder returns a new instance.
func newStackBuilder(
	requestIDFn func(r *http.Request) string,
	panicHandlerLoggerFn utiltypes.CtxLoggerFactoryFn,
) *StackBuilder {
	return &StackBuilder{
		stack: endpoint.NewStack(ReqHandlerWrapper(
			requestIDFn, panicHandlerLoggerFn,
		)),
	}
}
