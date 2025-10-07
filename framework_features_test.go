package framework_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	framework "github.com/aatuh/pureapi-framework"
)

type ctxKey string

func TestContextEnricherAddsPrincipal(t *testing.T) {
	engine := framework.NewEngine(
		framework.WithContextEnrichers(
			framework.NewContextEnricher(func(ctx context.Context, r *http.Request) (context.Context, error) {
				return context.WithValue(ctx, ctxKey("principal"), "user-123"), nil
			}),
		),
	)

	type in struct{}
	type out struct {
		Principal string `json:"principal"`
	}

	endpoint := framework.Endpoint[in, out](
		engine,
		http.MethodGet,
		"/whoami",
		func(ctx context.Context, _ in) (out, error) {
			principal, _ := ctx.Value(ctxKey("principal")).(string)
			return out{Principal: principal}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload out
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Principal != "user-123" {
		t.Fatalf("expected principal user-123, got %q", payload.Principal)
	}
}

func TestErrorsIncludeRequestID(t *testing.T) {
	engine := framework.NewEngine()
	type in struct{}
	type out struct{}

	endpoint := framework.Endpoint[in, out](
		engine,
		http.MethodGet,
		"/boom",
		func(ctx context.Context, _ in) (out, error) {
			return out{}, errors.New("boom")
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	type wireError struct {
		ID      string          `json:"id"`
		Message string          `json:"message"`
		Origin  string          `json:"origin"`
		Data    json.RawMessage `json:"data"`
	}
	var payload wireError
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.ID != "internal_error" {
		t.Fatalf("expected id internal_error, got %s", payload.ID)
	}
	reqID := rec.Header().Get("X-Request-ID")
	if reqID == "" {
		t.Fatalf("expected X-Request-ID header to be set")
	}
	if payload.Origin != reqID {
		t.Fatalf("expected origin %q, got %q", reqID, payload.Origin)
	}
}

func TestAccessLoggerReceivesEntry(t *testing.T) {
	logger := &recordingAccessLogger{}
	engine := framework.NewEngine(framework.WithAccessLoggers(logger))

	type in struct{}
	type out struct{}

	endpoint := framework.Endpoint[in, out](
		engine,
		http.MethodGet,
		"/ping",
		func(ctx context.Context, _ in) (out, error) {
			return out{}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if len(logger.entries) != 1 {
		t.Fatalf("expected single access log entry, got %d", len(logger.entries))
	}
	entry := logger.entries[0]
	if entry.Status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", entry.Status)
	}
	if entry.RequestID == "" {
		t.Fatalf("expected request id in access log entry")
	}
	if entry.Err != nil {
		t.Fatalf("did not expect error in access log entry: %v", entry.Err)
	}
}

func TestAuthorizationPolicyBlocksRequest(t *testing.T) {
	engine := framework.NewEngine(
		framework.WithAuthorizationPolicies(
			framework.AuthorizationPolicyFunc(func(ctx context.Context, _ any) error {
				return framework.ErrForbidden("no access")
			}),
		),
	)

	type in struct{}
	type out struct{}

	endpoint := framework.Endpoint[in, out](
		engine,
		http.MethodGet,
		"/secure",
		func(ctx context.Context, _ in) (out, error) {
			return out{}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode auth error payload: %v", err)
	}
	if payload.ID != "forbidden" {
		t.Fatalf("expected error id forbidden, got %s", payload.ID)
	}
}

func TestEngineHonorsCanceledRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	engine := framework.NewEngine()
	type in struct{}
	type out struct{}

	calls := 0
	endpoint := framework.Endpoint[in, out](
		engine,
		http.MethodGet,
		"/cancel",
		func(ctx context.Context, _ in) (out, error) {
			calls++
			return out{}, nil
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/cancel", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if calls != 0 {
		t.Fatalf("expected handler not to run when context canceled")
	}
}

func TestPanicsAreRecoveredWithRequestID(t *testing.T) {
	engine := framework.NewEngine()
	type in struct{}
	type out struct{}

	endpoint := framework.Endpoint[in, out](
		engine,
		http.MethodGet,
		"/panic",
		func(ctx context.Context, _ in) (out, error) {
			panic("kaboom")
		},
	)

	h := framework.NewHTTPHandler(framework.NewNoopEventEmitter())
	framework.RegisterEndpoints(h, endpoint)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var payload struct {
		ID     string `json:"id"`
		Origin string `json:"origin"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode panic response: %v", err)
	}
	if payload.ID != "internal_error" {
		t.Fatalf("expected internal_error id, got %s", payload.ID)
	}
	reqID := rec.Header().Get("X-Request-ID")
	if reqID == "" {
		t.Fatalf("expected X-Request-ID header after panic")
	}
	if payload.Origin != reqID {
		t.Fatalf("expected origin %q, got %q", reqID, payload.Origin)
	}
}

type recordingAccessLogger struct {
	entries []framework.AccessLogEntry
}

func (l *recordingAccessLogger) Log(_ context.Context, entry framework.AccessLogEntry) {
	l.entries = append(l.entries, entry)
}
