package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CORSOffensiveTestSuite is a suite of offensive tests for the CORS middleware.
type CORSOffensiveTestSuite struct {
	suite.Suite
	nextHandler http.Handler
}

// TestCORSOffensiveTestSuite runs the offensive CORS middleware tests.
func TestCORSOffensiveTestSuite(t *testing.T) {
	suite.Run(t, new(CORSOffensiveTestSuite))
}

// SetupTest sets up the test suite.
func (s *CORSOffensiveTestSuite) SetupTest() {
	// A simple next handler that writes "ok".
	s.nextHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		assert.NoError(s.T(), err)
	})
}

// TestOffensive_VeryLongOrigin ensures that extremely long Origin headers are
// handled safely.
func (s *CORSOffensiveTestSuite) TestOffensive_VeryLongOrigin() {
	// Create a very long origin string.
	longOrigin := strings.Repeat("https://example.com/", 1000)
	opts := CORSOptions{
		AllowedOrigins:   []string{longOrigin},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"X-Test"},
		AllowCredentials: true,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	req := httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", longOrigin)
	rr := httptest.NewRecorder()

	// Should not panic and should set the header to longOrigin.
	handler.ServeHTTP(rr, req)
	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	assert.Equal(s.T(), longOrigin, allowOrigin)
}

// TestOffensive_InvalidOriginCharacters tests that origins with control
// characters or injection attempts are not accepted.
func (s *CORSOffensiveTestSuite) TestOffensive_InvalidOriginCharacters() {
	opts := CORSOptions{
		AllowedOrigins:   []string{"https://safe.com"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"X-Safe"},
		AllowCredentials: false,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	// Malicious origin with newline and script tags.
	maliciousOrigin := "https://safe.com\n<script>alert(1)</script>"
	req := httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", maliciousOrigin)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Since the allowed origins do not match exactly,
	// no Access-Control-Allow-Origin header is set.
	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	assert.Empty(s.T(), allowOrigin)
}

// TestOffensive_EmptyAllowedOrigins tests behavior when no origins are allowed.
func (s *CORSOffensiveTestSuite) TestOffensive_EmptyAllowedOrigins() {
	opts := CORSOptions{
		AllowedOrigins:   []string{}, // empty list
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{},
		AllowCredentials: false,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	req := httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", "https://anything.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Expect no Access-Control-Allow-Origin since allowed origins is empty.
	assert.Empty(s.T(), rr.Header().Get("Access-Control-Allow-Origin"))
}

// TestOffensive_WhitespaceInOrigin tests that extra whitespace in the Origin
// header causes a mismatch.
func (s *CORSOffensiveTestSuite) TestOffensive_WhitespaceInOrigin() {
	opts := CORSOptions{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{},
		AllowCredentials: false,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	// Add extra spaces to the origin.
	req := httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", "  https://example.com  ")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Without trimming, the origin does not match exactly.
	assert.Empty(s.T(), rr.Header().Get("Access-Control-Allow-Origin"))
}

// TestOffensive_MaliciousAllowedHeaderValues ensures that even if
// AllowedHeaders contains suspicious strings, the header is constructed as a
// concatenation with "Content-Type".
func (s *CORSOffensiveTestSuite) TestOffensive_MaliciousAllowedHeaderValues() {
	opts := CORSOptions{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"POST"},
		AllowedHeaders:   []string{"X-Malicious", "X-Injected: drop table"},
		AllowCredentials: false,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	req := httptest.NewRequest("POST", "http://dummy", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	expectedHeaders := "Content-Type,X-Malicious,X-Injected: drop table"
	assert.Equal(
		s.T(), expectedHeaders, rr.Header().Get("Access-Control-Allow-Headers"),
	)
}

// TestOffensive_DuplicateAllowedOrigins verifies that duplicate allowed origins
// do not cause a crash and still match.
func (s *CORSOffensiveTestSuite) TestOffensive_DuplicateAllowedOrigins() {
	opts := CORSOptions{
		AllowedOrigins:   []string{"https://dup.com", "https://dup.com"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{},
		AllowCredentials: false,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	req := httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", "https://dup.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(
		s.T(), "https://dup.com",
		rr.Header().Get("Access-Control-Allow-Origin"),
	)
}

// TestOffensive_CredentialsWithWildcard verifies that when credentials are
// enabled, the middleware only echoes the origin if it is explicitly allowed.
func (s *CORSOffensiveTestSuite) TestOffensive_CredentialsWithWildcard() {
	// Case 1: AllowedOrigins is wildcard and credentials are enabled.
	opts := CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{},
		AllowCredentials: true,
	}
	middleware := CORS(opts)
	handler := middleware(s.nextHandler)

	req := httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", "https://any.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Expect no Access-Control-Allow-Origin header since it's a wildcard.
	assert.Empty(s.T(), rr.Header().Get("Access-Control-Allow-Origin"))

	// Case 2: AllowedOrigins explicitly includes the origin.
	opts = CORSOptions{
		AllowedOrigins:   []string{"https://any.com"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{},
		AllowCredentials: true,
	}
	middleware = CORS(opts)
	handler = middleware(s.nextHandler)
	req = httptest.NewRequest("GET", "http://dummy", nil)
	req.Header.Set("Origin", "https://any.com")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Origin should be echoed.
	assert.Equal(
		s.T(),
		"https://any.com",
		rr.Header().Get("Access-Control-Allow-Origin"),
	)
}
