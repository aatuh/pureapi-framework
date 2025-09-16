package middleware

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// dummyReqWrap implements the ReqWrap interface.
type dummyReqWrap struct {
	req  *http.Request
	body []byte
}

// implement ReqWrap interface
func (d *dummyReqWrap) GetRequest() *http.Request {
	return d.req
}
func (d *dummyReqWrap) GetBody() []byte {
	return d.body
}

// dummyResWrap implements the ResWrap interface.
type dummyResWrap struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// implement ResWrap interface
func (d *dummyResWrap) WriteHeader(code int) {
	d.statusCode = code
	d.ResponseWriter.WriteHeader(code)
}
func (d *dummyResWrap) StatusCode() int {
	return d.statusCode
}
func (d *dummyResWrap) Body() []byte {
	return d.body
}
func (d *dummyResWrap) Write(data []byte) (int, error) {
	d.body = append(d.body, data...)
	return d.ResponseWriter.Write(data)
}
func (d *dummyResWrap) Flush() {
	if fl, ok := d.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}
func (d *dummyResWrap) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := d.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("hijack not supported")
}

// dummyPanicHandler records panics.
type dummyPanicHandler struct {
	called   bool
	panicVal any
}

func (d *dummyPanicHandler) HandlePanic(
	w http.ResponseWriter, r *http.Request, err any,
) {
	d.called = true
	d.panicVal = err
	http.Error(w, "panic occurred", http.StatusInternalServerError)
}

// dummyLogger collects log messages.
type dummyLogger struct {
	messages []string
}

// Implement the Logger interface.
func (d *dummyLogger) Debug(args ...any) {
	d.messages = append(d.messages, fmt.Sprint(args...))
}
func (d *dummyLogger) Debugf(fmtStr string, args ...any) {
	d.messages = append(d.messages, fmt.Sprintf(fmtStr, args...))
}
func (d *dummyLogger) Trace(args ...any) {
	d.messages = append(d.messages, fmt.Sprint(args...))
}
func (d *dummyLogger) Tracef(fmtStr string, args ...any) {
	d.messages = append(d.messages, fmt.Sprintf(fmtStr, args...))
}
func (d *dummyLogger) Info(args ...any) {
	d.messages = append(d.messages, fmt.Sprint(args...))
}
func (d *dummyLogger) Infof(fmtStr string, args ...any) {
	d.messages = append(d.messages, fmt.Sprintf(fmtStr, args...))
}
func (d *dummyLogger) Warn(args ...any) {
	d.messages = append(d.messages, fmt.Sprint(args...))
}
func (d *dummyLogger) Warnf(fmtStr string, args ...any) {
	d.messages = append(d.messages, fmt.Sprintf(fmtStr, args...))
}
func (d *dummyLogger) Error(args ...any) {
	d.messages = append(d.messages, fmt.Sprint(args...))
}
func (d *dummyLogger) Errorf(fmtStr string, args ...any) {
	d.messages = append(d.messages, fmt.Sprintf(fmtStr, args...))
}
func (d *dummyLogger) Fatal(args ...any) {
	d.messages = append(d.messages, fmt.Sprint(args...))
}
func (d *dummyLogger) Fatalf(fmtStr string, args ...any) {
	d.messages = append(d.messages, fmt.Sprintf(fmtStr, args...))
}

// ReqHandlerTestSuite tests the ReqHandler middleware.
type ReqHandlerTestSuite struct {
	suite.Suite
	logger *dummyLogger
}

// TestReqHandlerTestSuite runs the test suite.
func TestReqHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ReqHandlerTestSuite))
}

// SetupTest sets up the test suite.
func (s *ReqHandlerTestSuite) SetupTest() {
	s.logger = &dummyLogger{}
}

// TestReqHandlerWithWrappers tests the middleware when both request and
// response wrappers are provided.
func (s *ReqHandlerTestSuite) TestReqHandlerWithWrappers() {
	var capturedReq *http.Request

	// Next handler captures the request and writes "OK".
	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			capturedReq = r
			_, err := w.Write([]byte("OK"))
			assert.NoError(s.T(), err)
		},
	)

	panicHandler := &dummyPanicHandler{}

	opts := &ReqHandlerOptions{
		TraceIDFactoryFn: func(r *http.Request) string {
			return "trace-123"
		},
		SpanIDFactoryFn: func(r *http.Request) string {
			return "span-456"
		},
		CtxLoggerFactoryFn: func(ctx context.Context) logging.ILogger {
			return s.logger
		},
		PanicHandler: panicHandler,
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return &dummyReqWrap{req: r, body: []byte("wrapped")}, nil
		},
		ResWrapFactoryFn: func(w http.ResponseWriter) ResWrap {
			return &dummyResWrap{ResponseWriter: w}
		},
	}

	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	s.Equal("OK", rr.Body.String(), "response body should be OK")

	// Verify that request metadata was set.
	meta := GetRequestMetadata(capturedReq.Context())
	s.NotNil(meta, "Request metadata should be set")
	s.Equal("trace-123", meta.TraceID, "TraceID should match")
	s.Equal("span-456", meta.SpanID, "SpanID should match")
	s.Equal("GET", meta.HTTPMethod, "HTTP method should match")
	s.Equal("example.com/test", meta.URL, "URL should match")

	// Verify that a response wrapper is stored in context.
	resWrap := GetResponseWrapper(capturedReq)
	s.NotNil(resWrap, "Response wrapper should be set")

	// Check that the logger recorded "Request completed".
	var found bool
	for _, msg := range s.logger.messages {
		if msg == "Request completed" {
			found = true
			break
		}
	}
	s.True(found, "Logger should log 'Request completed'")
}

// TestReqHandlerWithoutResponseWrapper tests the middleware when only a
// request wrapper is provided (i.e. no response wrapper).
func (s *ReqHandlerTestSuite) TestReqHandlerWithoutResponseWrapper() {
	var capturedReq *http.Request

	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			capturedReq = r
			_, err := w.Write([]byte("OK"))
			assert.NoError(s.T(), err)
		},
	)

	opts := &ReqHandlerOptions{
		TraceIDFactoryFn: func(r *http.Request) string {
			return "trace-abc"
		},
		SpanIDFactoryFn: func(r *http.Request) string {
			return "span-def"
		},
		CtxLoggerFactoryFn: func(ctx context.Context) logging.ILogger {
			return s.logger
		},
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return &dummyReqWrap{req: r, body: []byte("wrapped")}, nil
		},
		// ResWrapFactoryFn is not provided.
	}

	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("POST", "http://example.com/submit", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	s.Equal("OK", rr.Body.String(), "response body should be OK")
	meta := GetRequestMetadata(capturedReq.Context())
	s.NotNil(meta, "Request metadata should be set")
	s.Equal("trace-abc", meta.TraceID, "TraceID should match")
	s.Equal("span-def", meta.SpanID, "SpanID should match")
	s.Equal("POST", meta.HTTPMethod, "HTTP method should match")
	s.Equal("example.com/submit", meta.URL, "URL should match")

	// No response wrapper is provided, so GetResponseWrapper should return nil.
	resWrap := GetResponseWrapper(capturedReq)
	s.Nil(resWrap, "Response wrapper should be nil")
}

// TestPanicRecovery verifies that when the next handler panics the
// middleware recovers and invokes the PanicHandler.
func (s *ReqHandlerTestSuite) TestPanicRecovery() {
	panicHandler := &dummyPanicHandler{}

	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		},
	)

	opts := &ReqHandlerOptions{
		CtxLoggerFactoryFn: func(ctx context.Context) logging.ILogger {
			return s.logger
		},
		PanicHandler: panicHandler,
		ReqWrapFactoryFn: func(r *http.Request) (ReqWrap, error) {
			return &dummyReqWrap{req: r, body: []byte("wrapped")}, nil
		},
		// ResWrapFactoryFn can be nil.
	}

	handler := ReqHandler(opts)(nextHandler)
	req := httptest.NewRequest("GET", "http://example.com/fail", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()

	// The panic should be recovered and handled.
	handler.ServeHTTP(rr, req)

	s.True(panicHandler.called, "PanicHandler should be called")
	s.Equal("test panic", panicHandler.panicVal, "Panic value should match")
	s.Equal(http.StatusInternalServerError, rr.Code,
		"Response code should be 500")
	s.Contains(rr.Body.String(), "panic occurred",
		"Response body should indicate panic")
}
