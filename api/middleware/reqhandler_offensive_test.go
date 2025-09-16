package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReqHandlerOffensiveTestSuite struct {
	suite.Suite
	logger *dummyLogger
}

func TestReqHandlerOffensiveTestSuite(t *testing.T) {
	suite.Run(t, new(ReqHandlerOffensiveTestSuite))
}

func (s *ReqHandlerOffensiveTestSuite) SetupTest() {
	s.logger = &dummyLogger{}
}

// TestOffensive_ReqWrapFactoryError verifies that when the request wrapper
// factory returns an error, the middleware writes a 500 response and logs the
// error.
func (s *ReqHandlerOffensiveTestSuite) TestOffensive_ReqWrapFactoryError() {
	opts := &ReqHandlerOptions{
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return nil, fmt.Errorf("simulated reqwrap error")
		},
		CtxLoggerFactoryFn: func(ctx context.Context) logging.ILogger {
			return s.logger
		},
		PanicHandler: &dummyPanicHandler{},
	}
	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Should not be reached.
			_, err := w.Write([]byte("unexpected"))
			assert.NoError(s.T(), err)
		},
	)
	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	s.Equal(
		http.StatusInternalServerError, rr.Code,
		"Expected HTTP 500 when ReqWrapFactoryFn fails",
	)
	s.Contains(
		rr.Body.String(), http.StatusText(http.StatusInternalServerError),
	)
	s.NotEmpty(s.logger.messages, "Logger should record the wrapper error")
}

// TestOffensive_PanicNoHandler verifies that when no PanicHandler is provided,
// a panic in the next handler propagates.
func (s *ReqHandlerOffensiveTestSuite) TestOffensive_PanicNoHandler() {
	opts := &ReqHandlerOptions{
		// No PanicHandler provided.
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return &dummyReqWrap{req: r, body: []byte("wrapped")}, nil
		},
		// CtxLoggerFactoryFn provided.
		CtxLoggerFactoryFn: func(ctx context.Context) logging.ILogger {
			return s.logger
		},
	}
	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("unhandled panic")
		},
	)
	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("GET", "http://example.com/panic", nil)
	rr := httptest.NewRecorder()

	s.Panics(
		func() { handler.ServeHTTP(rr, req) },
		"Expected panic to propagate when PanicHandler is nil",
	)
}

// TestOffensive_InvalidRequest verifies that even when the request has an
// empty host, the middleware computes metadata without panicking.
func (s *ReqHandlerOffensiveTestSuite) TestOffensive_InvalidRequest() {
	opts := &ReqHandlerOptions{
		TraceIDFactoryFn: func(r *http.Request) string {
			return "trace-invalid"
		},
		SpanIDFactoryFn: func(r *http.Request) string {
			return "span-invalid"
		},
		CtxLoggerFactoryFn: func(ctx context.Context) logging.ILogger {
			return s.logger
		},
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return &dummyReqWrap{req: r, body: []byte("wrapped")}, nil
		},
	}
	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			capturedReq = r
			_, err := w.Write([]byte("OK"))
			assert.NoError(s.T(), err)
		},
	)
	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("GET", "http://dummy/invalid", nil)
	// Force empty Host.
	req.Host = ""
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	s.Equal("OK", rr.Body.String())

	// Check metadata on the request as received by the next handler.
	meta := GetRequestMetadata(capturedReq.Context())
	s.NotNil(meta, "Request metadata should be set even for an invalid host")
	// With an empty Host, the URL is computed as just the URL path.
	s.Equal(
		req.URL.Path, meta.URL,
		"Metadata URL should be just the path when host is empty",
	)
}

// TestOffensive_NilLogger verifies that a nil CtxLoggerFactoryFn does not cause
// panics.
func (s *ReqHandlerOffensiveTestSuite) TestOffensive_NilLogger() {
	var capturedReq *http.Request

	opts := &ReqHandlerOptions{
		TraceIDFactoryFn: func(r *http.Request) string { return "trace-nil" },
		SpanIDFactoryFn:  func(r *http.Request) string { return "span-nil" },
		// LoggerFn is nil.
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return &dummyReqWrap{req: r, body: []byte("wrapped")}, nil
		},
	}

	// Capture the request received by the next handler.
	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			capturedReq = r
			_, err := w.Write([]byte("OK"))
			assert.NoError(s.T(), err)
		},
	)
	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("GET", "http://example.com/nil", nil)
	rr := httptest.NewRecorder()

	// Execute the middleware chain.
	handler.ServeHTTP(rr, req)
	s.Equal("OK", rr.Body.String(), "response body should be OK")

	// Use the request as seen by the next handler.
	meta := GetRequestMetadata(capturedReq.Context())
	s.NotNil(meta, "Request metadata should be set when LoggerFn is nil")
	s.Equal("trace-nil", meta.TraceID)
	s.Equal("span-nil", meta.SpanID)
}
