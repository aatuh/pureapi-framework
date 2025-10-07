package framework

import (
	"context"
	"net/http"

	coreendpoint "github.com/aatuh/pureapi-core/endpoint"
	coreevent "github.com/aatuh/pureapi-core/event"
	coreserver "github.com/aatuh/pureapi-core/server"

	"github.com/aatuh/pureapi-framework/binder"
	"github.com/aatuh/pureapi-framework/engine"
	"github.com/aatuh/pureapi-framework/errors"
	"github.com/aatuh/pureapi-framework/hooks"
	"github.com/aatuh/pureapi-framework/obs/accesslog"
	codecjson "github.com/aatuh/pureapi-framework/renderer/json"
	"github.com/aatuh/pureapi-framework/renderer/registry"
	"github.com/aatuh/pureapi-framework/security/cors"
	securityheaders "github.com/aatuh/pureapi-framework/security/headers"
)

// Facade type aliases keep consumers on the root package.
type (
	// EndpointSpec mirrors the pureapi-core endpoint specification contract.
	EndpointSpec = coreendpoint.EndpointSpec
	// RuntimeEndpoint represents a registered HTTP endpoint.
	RuntimeEndpoint = coreendpoint.Endpoint
	// Middleware is a request/response decorator applied around handlers.
	Middleware = coreendpoint.Middleware
	// Middlewares is a middleware collection that can be chained onto handlers.
	Middlewares = coreendpoint.Middlewares
	// EventEmitter emits framework/server lifecycle events.
	EventEmitter = coreevent.EventEmitter
	// Event represents an emitted event payload.
	Event = coreevent.Event
	// EventType categorises emitted events.
	EventType = coreevent.EventType
	// HTTPHandler drives request routing and middleware execution.
	HTTPHandler = coreserver.Handler
	// HTTPHandlerOption configures the HTTP handler.
	HTTPHandlerOption = coreserver.HandlerOption
)

// NewHTTPHandler constructs an HTTP handler with the provided emitter and options.
func NewHTTPHandler(emitter EventEmitter, opts ...HTTPHandlerOption) *HTTPHandler {
	return coreserver.NewHandler(emitter, opts...)
}

// RegisterEndpoints converts the declarative specs into runtime endpoints and
// registers them on the handler.
func RegisterEndpoints(handler *HTTPHandler, specs ...EndpointSpec) {
	if handler == nil || len(specs) == 0 {
		return
	}
	handler.Register(coreendpoint.ToEndpoints(specs...))
}

// NewNoopEventEmitter returns a no-op emitter suitable for tests.
func NewNoopEventEmitter() EventEmitter {
	return coreevent.NewNoopEventEmitter()
}

// NewMiddlewares constructs an immutable middleware collection.
func NewMiddlewares(mw ...Middleware) Middlewares {
	return coreendpoint.NewMiddlewares(mw...)
}

// RequestIDMiddleware exposes the pureapi-core request ID middleware.
func RequestIDMiddleware() Middleware {
	return coreendpoint.RequestIDMiddleware()
}

// Re-export types from subpackages
type (
	// Binder converts HTTP requests into typed inputs.
	Binder = binder.Binder
	// BodyDecoder decodes a raw body into dest.
	BodyDecoder = binder.BodyDecoder
	// JSONBodyDecoder implements BodyDecoder using encoding/json.
	JSONBodyDecoder = binder.JSONBodyDecoder
	// DefaultBinder is the framework binder implementation.
	DefaultBinder = binder.DefaultBinder
	// FieldSource identifies where a value originated.
	FieldSource = binder.FieldSource
	// FieldError describes a single binding failure.
	FieldError = binder.FieldError
	// BindError aggregates binding failures.
	BindError = binder.BindError

	// RenderFunc renders payloads as bytes and content type.
	RenderFunc = registry.RenderFunc
	// JSONRenderer renders JSON payloads.
	JSONRenderer = codecjson.Renderer

	// ErrorCatalog keeps registered catalog entries keyed by ID.
	ErrorCatalog = errors.ErrorCatalog
	// ErrorMapper maps Go errors to catalog entries.
	ErrorMapper = errors.ErrorMapper
	// CatalogEntry describes a wire error returned by the framework.
	CatalogEntry = errors.CatalogEntry

	// InputHook processes bound input before handler execution.
	InputHook = hooks.InputHook
	// OutputHook processes handler output before rendering.
	OutputHook = hooks.OutputHook
	// ContextEnricher attaches values to the context ahead of handler execution.
	ContextEnricher = hooks.ContextEnricher
	// AuthorizationPolicy authorizes access to a resource.
	AuthorizationPolicy = hooks.AuthorizationPolicy
	// AuthorizationPolicyFunc lifts a function into an AuthorizationPolicy.
	AuthorizationPolicyFunc = hooks.AuthorizationPolicyFunc
	// AccessLogger receives structured access log entries.
	AccessLogger = accesslog.AccessLogger
	// AccessLogEntry holds structured access log data.
	AccessLogEntry = accesslog.Entry
	// CORSConfig controls the provided CORS middleware.
	CORSConfig = cors.Config
	// SecurityHeadersConfig controls the security header middleware.
	SecurityHeadersConfig = securityheaders.Config

	// Engine wires declarative endpoints to pureapi-core primitives.
	Engine = engine.Engine
	// EngineOption configures a new Engine.
	EngineOption = engine.EngineOption
	// HandlerFunc is the generic endpoint handler signature.
	HandlerFunc[TIn any, TOut any] = engine.HandlerFunc[TIn, TOut]
	// DeclarativeEndpoint describes an endpoint with binding and rendering details.
	DeclarativeEndpoint[TIn any, TOut any] = engine.DeclarativeEndpoint[TIn, TOut]
	// EndpointOption configures a declarative endpoint.
	EndpointOption[TIn any, TOut any] = engine.EndpointOption[TIn, TOut]
	// EndpointMeta carries optional documentation metadata.
	EndpointMeta = engine.EndpointMeta
)

// Re-export functions from subpackages
var (
	// Binder functions
	NewDefaultBinder             = binder.NewDefaultBinder
	NewFieldError                = binder.NewFieldError
	NewBindError                 = binder.NewBindError
	NewRendererRegistry          = registry.New
	DefaultSecurityHeadersConfig = securityheaders.DefaultConfig

	// Error functions
	NewErrorCatalog     = errors.NewErrorCatalog
	DefaultErrorCatalog = errors.DefaultErrorCatalog
	NewErrorMapper      = errors.NewErrorMapper
	RenderError         = errors.RenderError

	// Authorization helpers
	NewAuthorizationError = hooks.NewAuthorizationError
	ErrUnauthorized       = hooks.ErrUnauthorized
	ErrForbidden          = hooks.ErrForbidden

	// Access log helpers
	NewStdAccessLogger = accesslog.NewStdLogger

	// Hook functions - generic functions cannot be re-exported directly
	// Use hooks.NewInputHook[T] and hooks.NewOutputHook[T] directly

	// Engine functions - generic functions cannot be re-exported directly
	// Use engine.NewEngine, engine.Endpoint[TIn, TOut], etc. directly
)

// Re-export constants from subpackages
const (
	SourcePath   = binder.SourcePath
	SourceQuery  = binder.SourceQuery
	SourceHeader = binder.SourceHeader
	SourceCookie = binder.SourceCookie
	SourceBody   = binder.SourceBody
)

// Wrapper functions for generic types that can be re-exported

// NewEngine builds an Engine using framework defaults.
func NewEngine(opts ...EngineOption) *Engine {
	return engine.NewEngine(opts...)
}

// Non-generic engine options
var (
	WithBinder                = engine.WithBinder
	WithErrorMapper           = engine.WithErrorMapper
	WithGlobalMiddlewares     = engine.WithGlobalMiddlewares
	WithContextEnrichers      = engine.WithContextEnrichers
	WithAuthorizationPolicies = engine.WithAuthorizationPolicies
	WithAccessLoggers         = engine.WithAccessLoggers
	WithInputHooks            = engine.WithInputHooks
	WithOutputHooks           = engine.WithOutputHooks
)

func NewInputHook[T any](fn func(ctx context.Context, value *T) error) InputHook {
	return hooks.NewInputHook(fn)
}

func NewContextEnricher(fn func(ctx context.Context, r *http.Request) (context.Context, error)) ContextEnricher {
	return hooks.NewContextEnricher(fn)
}

func NewOutputHook[T any](fn func(ctx context.Context, value *T) error) OutputHook {
	return hooks.NewOutputHook(fn)
}

func NewCORSMiddleware(cfg CORSConfig) Middleware {
	return cors.Middleware(cfg)
}

func NewSecurityHeadersMiddleware(cfg SecurityHeadersConfig) Middleware {
	return securityheaders.Middleware(cfg)
}

func NewDefaultSecurityHeadersMiddleware() Middleware {
	return securityheaders.Default()
}

func Endpoint[TIn any, TOut any](eng *Engine, method, path string, handler HandlerFunc[TIn, TOut], opts ...EndpointOption[TIn, TOut]) *DeclarativeEndpoint[TIn, TOut] {
	return engine.Endpoint(eng, method, path, handler, opts...)
}

func WithRenderer(contentType string, fn RenderFunc) EngineOption {
	return engine.WithRenderer(contentType, fn)
}

func WithMeta[TIn any, TOut any](meta EndpointMeta) EndpointOption[TIn, TOut] {
	return engine.WithMeta[TIn, TOut](meta)
}

func WithEndpointMiddlewares[TIn any, TOut any](mw ...Middleware) EndpointOption[TIn, TOut] {
	return engine.WithEndpointMiddlewares[TIn, TOut](mw...)
}

func WithSuccessStatus[TIn any, TOut any](status int) EndpointOption[TIn, TOut] {
	return engine.WithSuccessStatus[TIn, TOut](status)
}

func WithEndpointBinder[TIn any, TOut any](binder Binder) EndpointOption[TIn, TOut] {
	return engine.WithEndpointBinder[TIn, TOut](binder)
}

func WithEndpointRenderer[TIn any, TOut any](contentType string, fn RenderFunc) EndpointOption[TIn, TOut] {
	return engine.WithEndpointRenderer[TIn, TOut](contentType, fn)
}

func WithEndpointContextEnrichers[TIn any, TOut any](enrichers ...ContextEnricher) EndpointOption[TIn, TOut] {
	return engine.WithEndpointContextEnrichers[TIn, TOut](enrichers...)
}

func WithEndpointAuthorizationPolicies[TIn any, TOut any](policies ...AuthorizationPolicy) EndpointOption[TIn, TOut] {
	return engine.WithEndpointAuthorizationPolicies[TIn, TOut](policies...)
}

func WithEndpointAccessLoggers[TIn any, TOut any](loggers ...AccessLogger) EndpointOption[TIn, TOut] {
	return engine.WithEndpointAccessLoggers[TIn, TOut](loggers...)
}

func WithEndpointErrorMapper[TIn any, TOut any](mapper *ErrorMapper) EndpointOption[TIn, TOut] {
	return engine.WithEndpointErrorMapper[TIn, TOut](mapper)
}

func WithEndpointInputHooks[TIn any, TOut any](hooks ...InputHook) EndpointOption[TIn, TOut] {
	return engine.WithEndpointInputHooks[TIn, TOut](hooks...)
}

func WithEndpointOutputHooks[TIn any, TOut any](hooks ...OutputHook) EndpointOption[TIn, TOut] {
	return engine.WithEndpointOutputHooks[TIn, TOut](hooks...)
}
