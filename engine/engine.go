package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-framework/binder"
	frameworkerrors "github.com/aatuh/pureapi-framework/errors"
	"github.com/aatuh/pureapi-framework/hooks"
	"github.com/aatuh/pureapi-framework/obs/accesslog"
	codecjson "github.com/aatuh/pureapi-framework/renderer/json"
	"github.com/aatuh/pureapi-framework/renderer/registry"
)

// HandlerFunc is the generic endpoint handler signature.
type HandlerFunc[TIn any, TOut any] func(ctx context.Context, input TIn) (TOut, error)

// Engine wires declarative endpoints to pureapi-core primitives.
type Engine struct {
	binder                binder.Binder
	renderRegistry        *registry.Registry
	errorMapper           *frameworkerrors.ErrorMapper
	catalog               *frameworkerrors.ErrorCatalog
	globalMiddlewares     []endpoint.Middleware
	contextEnrichers      []hooks.ContextEnricher
	authorizationPolicies []hooks.AuthorizationPolicy
	accessLoggers         []accesslog.AccessLogger
	inputHooks            []hooks.InputHook
	outputHooks           []hooks.OutputHook
}

// EngineOption configures a new Engine.
type EngineOption func(*Engine)

type rendererRegistration struct {
	contentType string
	fn          registry.RenderFunc
}

// WithBinder overrides the default binder implementation.
func WithBinder(b binder.Binder) EngineOption {
	return func(e *Engine) {
		if b != nil {
			e.binder = b
		}
	}
}

// WithRenderer overrides the default renderer implementation.
func WithRenderer(contentType string, fn registry.RenderFunc) EngineOption {
	return func(e *Engine) {
		if fn == nil {
			return
		}
		if e.renderRegistry == nil {
			e.renderRegistry = registry.New(contentType, fn)
			return
		}
		e.renderRegistry.Register(contentType, fn)
	}
}

// WithErrorMapper overrides the default error mapper.
func WithErrorMapper(mapper *frameworkerrors.ErrorMapper) EngineOption {
	return func(e *Engine) {
		if mapper != nil {
			e.errorMapper = mapper
		}
	}
}

// WithGlobalMiddlewares adds middlewares applied to every endpoint.
func WithGlobalMiddlewares(mw ...endpoint.Middleware) EngineOption {
	return func(e *Engine) {
		if len(mw) == 0 {
			return
		}
		combined := make([]endpoint.Middleware, 0, len(e.globalMiddlewares)+len(mw))
		combined = append(combined, e.globalMiddlewares...)
		combined = append(combined, mw...)
		e.globalMiddlewares = combined
	}
}

// WithContextEnrichers registers enrichers that run on every endpoint before binding.
func WithContextEnrichers(enrichers ...hooks.ContextEnricher) EngineOption {
	return func(e *Engine) {
		for _, enricher := range enrichers {
			if enricher == nil {
				continue
			}
			e.contextEnrichers = append(e.contextEnrichers, enricher)
		}
	}
}

// WithAuthorizationPolicies registers authorization policies evaluated after input hooks.
func WithAuthorizationPolicies(policies ...hooks.AuthorizationPolicy) EngineOption {
	return func(e *Engine) {
		for _, policy := range policies {
			if policy == nil {
				continue
			}
			e.authorizationPolicies = append(e.authorizationPolicies, policy)
		}
	}
}

// WithAccessLoggers registers structured access loggers.
func WithAccessLoggers(loggers ...accesslog.AccessLogger) EngineOption {
	return func(e *Engine) {
		for _, logger := range loggers {
			if logger == nil {
				continue
			}
			e.accessLoggers = append(e.accessLoggers, logger)
		}
	}
}

// WithInputHooks registers hooks that run on every endpoint before handler execution.
func WithInputHooks(hooks ...hooks.InputHook) EngineOption {
	return func(e *Engine) {
		for _, hook := range hooks {
			if hook == nil {
				continue
			}
			e.inputHooks = append(e.inputHooks, hook)
		}
	}
}

// WithOutputHooks registers hooks that run on every endpoint before rendering.
func WithOutputHooks(hooks ...hooks.OutputHook) EngineOption {
	return func(e *Engine) {
		for _, hook := range hooks {
			if hook == nil {
				continue
			}
			e.outputHooks = append(e.outputHooks, hook)
		}
	}
}

// NewEngine builds an Engine using framework defaults.
func NewEngine(opts ...EngineOption) *Engine {
	catalog := frameworkerrors.DefaultErrorCatalog()
	mapper, _ := frameworkerrors.NewErrorMapper(catalog, "internal_error")

	jsonRenderer := codecjson.Renderer{}
	engine := &Engine{
		binder:            binder.NewDefaultBinder(),
		renderRegistry:    registry.New("application/json", jsonRenderer.RenderFunc()),
		errorMapper:       mapper,
		catalog:           catalog,
		globalMiddlewares: []endpoint.Middleware{endpoint.RequestIDMiddleware()},
	}
	for _, opt := range opts {
		opt(engine)
	}
	engine.ensureDefaults()
	return engine
}

func (e *Engine) ensureDefaults() {
	if e.binder == nil {
		e.binder = binder.NewDefaultBinder()
	}
	if e.renderRegistry == nil {
		jsonRenderer := codecjson.Renderer{}
		e.renderRegistry = registry.New("application/json", jsonRenderer.RenderFunc())
	}
	if e.catalog == nil {
		e.catalog = frameworkerrors.DefaultErrorCatalog()
	}
	if e.errorMapper == nil {
		mapper, _ := frameworkerrors.NewErrorMapper(e.catalog, "internal_error")
		e.errorMapper = mapper
	}
	_ = e.errorMapper.RegisterType((*binder.BindError)(nil), "invalid_request")
}

// EndpointOption configures a declarative endpoint.
type EndpointOption[TIn any, TOut any] func(*DeclarativeEndpoint[TIn, TOut])

// EndpointMeta carries optional documentation metadata.
type EndpointMeta struct {
	Summary     string
	Description string
	OperationID string
	Tags        []string
	Extras      map[string]any
}

// WithMeta sets the endpoint metadata.
func WithMeta[TIn any, TOut any](meta EndpointMeta) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		ep.Meta = meta
	}
}

// WithEndpointMiddlewares attaches per-endpoint middlewares.
func WithEndpointMiddlewares[TIn any, TOut any](mw ...endpoint.Middleware) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		if len(mw) == 0 {
			return
		}
		combined := make([]endpoint.Middleware, 0, len(ep.middlewares)+len(mw))
		combined = append(combined, ep.middlewares...)
		combined = append(combined, mw...)
		ep.middlewares = combined
	}
}

// WithSuccessStatus overrides the default success status code.
func WithSuccessStatus[TIn any, TOut any](status int) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		ep.successStatus = status
	}
}

// WithEndpointBinder overrides the binder for this endpoint only.
func WithEndpointBinder[TIn any, TOut any](binder binder.Binder) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		if binder != nil {
			ep.binder = binder
		}
	}
}

// WithEndpointRenderer overrides the renderer for this endpoint only.
func WithEndpointRenderer[TIn any, TOut any](contentType string, fn registry.RenderFunc) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		if fn == nil {
			return
		}
		ep.renderers = append(ep.renderers, rendererRegistration{contentType: contentType, fn: fn})
	}
}

// WithEndpointErrorMapper overrides the error mapper for this endpoint.
func WithEndpointErrorMapper[TIn any, TOut any](mapper *frameworkerrors.ErrorMapper) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		if mapper != nil {
			ep.errorMapper = mapper
		}
	}
}

// WithEndpointContextEnrichers attaches per-endpoint context enrichers executed before binding.
func WithEndpointContextEnrichers[TIn any, TOut any](enrichers ...hooks.ContextEnricher) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		for _, enricher := range enrichers {
			if enricher == nil {
				continue
			}
			ep.contextEnrichers = append(ep.contextEnrichers, enricher)
		}
	}
}

// WithEndpointAuthorizationPolicies attaches per-endpoint authorization policies.
func WithEndpointAuthorizationPolicies[TIn any, TOut any](policies ...hooks.AuthorizationPolicy) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		for _, policy := range policies {
			if policy == nil {
				continue
			}
			ep.authorizationPolicies = append(ep.authorizationPolicies, policy)
		}
	}
}

// WithEndpointAccessLoggers attaches per-endpoint access loggers.
func WithEndpointAccessLoggers[TIn any, TOut any](loggers ...accesslog.AccessLogger) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		for _, logger := range loggers {
			if logger == nil {
				continue
			}
			ep.accessLoggers = append(ep.accessLoggers, logger)
		}
	}
}

// WithEndpointInputHooks attaches per-endpoint hooks that operate on the bound input.
func WithEndpointInputHooks[TIn any, TOut any](hooks ...hooks.InputHook) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		for _, hook := range hooks {
			if hook == nil {
				continue
			}
			ep.inputHooks = append(ep.inputHooks, hook)
		}
	}
}

// WithEndpointOutputHooks attaches per-endpoint hooks that operate on the handler output.
func WithEndpointOutputHooks[TIn any, TOut any](hooks ...hooks.OutputHook) EndpointOption[TIn, TOut] {
	return func(ep *DeclarativeEndpoint[TIn, TOut]) {
		for _, hook := range hooks {
			if hook == nil {
				continue
			}
			ep.outputHooks = append(ep.outputHooks, hook)
		}
	}
}

// Endpoint creates a declarative endpoint definition bound to engine.
func Endpoint[TIn any, TOut any](engine *Engine, method, path string, handler HandlerFunc[TIn, TOut], opts ...EndpointOption[TIn, TOut]) *DeclarativeEndpoint[TIn, TOut] {
	if engine == nil {
		panic("framework Endpoint: engine must not be nil")
	}
	declarative := &DeclarativeEndpoint[TIn, TOut]{
		engine:        engine,
		Method:        strings.ToUpper(strings.TrimSpace(method)),
		Path:          path,
		handler:       handler,
		successStatus: 0,
	}
	for _, opt := range opts {
		opt(declarative)
	}
	return declarative
}

// DeclarativeEndpoint describes an endpoint with binding and rendering details.
type DeclarativeEndpoint[TIn any, TOut any] struct {
	engine *Engine
	Method string
	Path   string
	Meta   EndpointMeta

	handler HandlerFunc[TIn, TOut]

	binder                binder.Binder
	errorMapper           *frameworkerrors.ErrorMapper
	middlewares           []endpoint.Middleware
	contextEnrichers      []hooks.ContextEnricher
	authorizationPolicies []hooks.AuthorizationPolicy
	accessLoggers         []accesslog.AccessLogger
	renderers             []rendererRegistration
	inputHooks            []hooks.InputHook
	outputHooks           []hooks.OutputHook
	successStatus         int
}

var _ endpoint.EndpointSpec = (*DeclarativeEndpoint[any, any])(nil)

// ToEndpoint converts the declarative endpoint into a pureapi-core endpoint.
func (d *DeclarativeEndpoint[TIn, TOut]) ToEndpoint() endpoint.Endpoint {
	binder := d.binder
	if binder == nil {
		binder = d.engine.binder
	}
	mapper := d.errorMapper
	if mapper == nil {
		mapper = d.engine.errorMapper
	}
	combined := make([]endpoint.Middleware, 0, len(d.engine.globalMiddlewares)+len(d.middlewares))
	combined = append(combined, d.engine.globalMiddlewares...)
	combined = append(combined, d.middlewares...)

	contextEnrichers := append([]hooks.ContextEnricher{}, d.engine.contextEnrichers...)
	contextEnrichers = append(contextEnrichers, d.contextEnrichers...)
	authorizationPolicies := append([]hooks.AuthorizationPolicy{}, d.engine.authorizationPolicies...)
	authorizationPolicies = append(authorizationPolicies, d.authorizationPolicies...)
	accessLoggers := append([]accesslog.AccessLogger{}, d.engine.accessLoggers...)
	accessLoggers = append(accessLoggers, d.accessLoggers...)
	renderReg := d.engine.renderRegistry.Clone()
	if renderReg == nil {
		jsonRenderer := codecjson.Renderer{}
		renderReg = registry.New("application/json", jsonRenderer.RenderFunc())
	}
	for _, rr := range d.renderers {
		renderReg.Register(rr.contentType, rr.fn)
	}
	inputHooks := append([]hooks.InputHook{}, d.engine.inputHooks...)
	inputHooks = append(inputHooks, d.inputHooks...)
	outputHooks := append([]hooks.OutputHook{}, d.engine.outputHooks...)
	outputHooks = append(outputHooks, d.outputHooks...)

	handler := d.wrapHandler(binder, renderReg, mapper, contextEnrichers, authorizationPolicies, accessLoggers, inputHooks, outputHooks)
	var core endpoint.Endpoint = endpoint.NewEndpoint(d.Path, d.Method)
	if len(combined) > 0 {
		core = core.WithMiddlewares(endpoint.NewMiddlewares(combined...))
	}
	core = core.WithHandler(handler)
	return core
}

func (d *DeclarativeEndpoint[TIn, TOut]) wrapHandler(
	binder binder.Binder,
	renderRegistry *registry.Registry,
	mapper *frameworkerrors.ErrorMapper,
	contextEnrichers []hooks.ContextEnricher,
	authorizationPolicies []hooks.AuthorizationPolicy,
	accessLoggers []accesslog.AccessLogger,
	inputHooks []hooks.InputHook,
	outputHooks []hooks.OutputHook,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lw := newLoggingResponseWriter(w)
		w = lw
		start := time.Now()
		var handlerErr error

		defer func() {
			entry := accesslog.Entry{
				Method:       r.Method,
				Path:         r.URL.Path,
				Status:       lw.Status(),
				Duration:     time.Since(start),
				RequestID:    endpoint.RequestIDFromContext(ctx),
				RemoteAddr:   r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				ResponseSize: lw.BytesWritten(),
				Err:          handlerErr,
			}
			for _, logger := range accessLoggers {
				if logger == nil {
					continue
				}
				logger.Log(ctx, entry)
			}
		}()

		defer func() {
			if rec := recover(); rec != nil {
				var panicErr error
				switch v := rec.(type) {
				case error:
					panicErr = fmt.Errorf("panic: %w", v)
				default:
					panicErr = fmt.Errorf("panic: %v", v)
				}
				handlerErr = panicErr
				d.writeError(ctx, lw, renderRegistry, r, mapper, panicErr)
			}
		}()

		var err error
		if ctx, err = executeContextEnrichers(ctx, r, contextEnrichers); err != nil {
			handlerErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			d.writeError(ctx, lw, renderRegistry, r, mapper, err)
			return
		}
		r = r.WithContext(ctx)

		var input TIn
		if binder != nil {
			if err = binder.Bind(ctx, r, &input); err != nil {
				handlerErr = err
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				d.writeError(ctx, lw, renderRegistry, r, mapper, err)
				return
			}
		}
		if err = executeInputHooks(ctx, &input, inputHooks); err != nil {
			handlerErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			d.writeError(ctx, lw, renderRegistry, r, mapper, err)
			return
		}
		if err = executeAuthorizationPolicies(ctx, &input, authorizationPolicies); err != nil {
			handlerErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			d.writeError(ctx, lw, renderRegistry, r, mapper, err)
			return
		}

		var output TOut
		if output, err = d.handler(ctx, input); err != nil {
			handlerErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			d.writeError(ctx, lw, renderRegistry, r, mapper, err)
			return
		}
		if err = executeOutputHooks(ctx, &output, outputHooks); err != nil {
			handlerErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			d.writeError(ctx, lw, renderRegistry, r, mapper, err)
			return
		}
		status := d.successStatus
		if status == 0 {
			status = defaultSuccessStatus(d.Method)
		}
		if err = renderRegistry.Render(ctx, lw, r, status, output); err != nil {
			handlerErr = err
			http.Error(lw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (d *DeclarativeEndpoint[TIn, TOut]) writeError(
	ctx context.Context,
	w http.ResponseWriter,
	renderRegistry *registry.Registry,
	req *http.Request,
	mapper *frameworkerrors.ErrorMapper,
	err error,
) {
	mapped := mapper.Map(err)
	payload := frameworkerrors.RenderError(mapped)
	if requestID := endpoint.RequestIDFromContext(ctx); requestID != "" {
		payload = payload.WithOrigin(requestID)
	}
	if renderRegistry != nil {
		if renderErr := renderRegistry.Render(ctx, w, req, mapped.Entry.Status, payload); renderErr == nil {
			return
		}
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func executeAuthorizationPolicies(ctx context.Context, input any, policies []hooks.AuthorizationPolicy) error {
	if len(policies) == 0 {
		return nil
	}
	for _, policy := range policies {
		if policy == nil {
			continue
		}
		if err := policy.Authorize(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func executeContextEnrichers(ctx context.Context, r *http.Request, enrichers []hooks.ContextEnricher) (context.Context, error) {
	if len(enrichers) == 0 {
		return ctx, nil
	}
	current := ctx
	for _, enricher := range enrichers {
		if enricher == nil {
			continue
		}
		updated, err := enricher.Enrich(current, r)
		if err != nil {
			return current, err
		}
		if updated != nil {
			current = updated
		}
	}
	return current, nil
}

func executeInputHooks(ctx context.Context, value any, hooks []hooks.InputHook) error {
	if len(hooks) == 0 {
		return nil
	}
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := hook.Process(ctx, value); err != nil {
			return err
		}
	}
	return nil
}

func executeOutputHooks(ctx context.Context, value any, hooks []hooks.OutputHook) error {
	if len(hooks) == 0 {
		return nil
	}
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := hook.Process(ctx, value); err != nil {
			return err
		}
	}
	return nil
}

func defaultSuccessStatus(method string) int {
	switch strings.ToUpper(method) {
	case http.MethodPost:
		return http.StatusCreated
	case http.MethodDelete:
		return http.StatusNoContent
	default:
		return http.StatusOK
	}
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (lw *loggingResponseWriter) WriteHeader(status int) {
	lw.status = status
	lw.ResponseWriter.WriteHeader(status)
}

func (lw *loggingResponseWriter) Write(p []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	n, err := lw.ResponseWriter.Write(p)
	lw.bytes += n
	return n, err
}

func (lw *loggingResponseWriter) Status() int {
	if lw.status == 0 {
		return http.StatusOK
	}
	return lw.status
}

func (lw *loggingResponseWriter) BytesWritten() int {
	return lw.bytes
}
