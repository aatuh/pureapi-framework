package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CORSTestSuite is a suite of tests for the CORS middleware.
type CORSTestSuite struct {
	suite.Suite
}

// TestCORSTestSuite runs the test suite.
func TestCORSTestSuite(t *testing.T) {
	suite.Run(t, new(CORSTestSuite))
}

// TestCORSMiddleware tests the CORS middleware.
func (s *CORSTestSuite) TestCORSMiddleware() {
	type corsTestCase struct {
		name                     string
		opts                     CORSOptions
		requestOrigin            string
		expectedAllowOrigin      string
		expectedAllowMethods     string
		expectedAllowHeaders     string
		expectedAllowCredentials string
	}

	testCases := []corsTestCase{
		{
			name: "Credentials allowed with matching origin",
			opts: CORSOptions{
				AllowedOrigins:   []string{"https://example.com"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"X-Custom"},
				AllowCredentials: true,
			},
			requestOrigin:            "https://example.com",
			expectedAllowOrigin:      "https://example.com",
			expectedAllowMethods:     "GET,POST",
			expectedAllowHeaders:     "Content-Type,X-Custom",
			expectedAllowCredentials: "true",
		},
		{
			name: "Wildcard without credentials",
			opts: CORSOptions{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{},
				AllowCredentials: false,
			},
			requestOrigin:            "https://malicious.com",
			expectedAllowOrigin:      "*",
			expectedAllowMethods:     "GET",
			expectedAllowHeaders:     "Content-Type",
			expectedAllowCredentials: "",
		},
		{
			name: "Multiple origins, no credentials, matching origin",
			opts: CORSOptions{
				AllowedOrigins:   []string{"https://foo.com", "https://bar.com"},
				AllowedMethods:   []string{"PUT"},
				AllowedHeaders:   []string{"X-Requested-With"},
				AllowCredentials: false,
			},
			requestOrigin:            "https://bar.com",
			expectedAllowOrigin:      "https://bar.com",
			expectedAllowMethods:     "PUT",
			expectedAllowHeaders:     "Content-Type,X-Requested-With",
			expectedAllowCredentials: "",
		},
		{
			name: "Wildcard with credentials but origin not matching",
			opts: CORSOptions{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "OPTIONS"},
				AllowedHeaders:   []string{"X-Test"},
				AllowCredentials: true,
			},
			requestOrigin:            "https://any.com",
			expectedAllowOrigin:      "",
			expectedAllowMethods:     "GET,OPTIONS",
			expectedAllowHeaders:     "Content-Type,X-Test",
			expectedAllowCredentials: "true",
		},
		{
			name: "Empty Origin no match",
			opts: CORSOptions{
				AllowedOrigins:   []string{"https://example.com"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{},
				AllowCredentials: false,
			},
			requestOrigin:            "",
			expectedAllowOrigin:      "",
			expectedAllowMethods:     "GET",
			expectedAllowHeaders:     "Content-Type",
			expectedAllowCredentials: "",
		},
	}

	// Create a dummy next handler that simply writes "ok".
	nextHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("ok"))
			assert.NoError(s.T(), err)
		},
	)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Wrap the next handler with our CORS middleware.
			middleware := CORS(tc.opts)
			handler := middleware(nextHandler)

			// Create a new HTTP request and set the Origin header if provided.
			req := httptest.NewRequest("GET", "http://dummy", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}

			// Create a ResponseRecorder.
			rr := httptest.NewRecorder()

			// Serve the request.
			handler.ServeHTTP(rr, req)

			// Check the Access-Control-Allow-Origin header.
			allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			assert.Equal(s.T(), tc.expectedAllowOrigin, allowOrigin,
				"Access-Control-Allow-Origin mismatch",
			)

			// Check the Access-Control-Allow-Methods header.
			allowMethods := rr.Header().Get("Access-Control-Allow-Methods")
			assert.Equal(s.T(), tc.expectedAllowMethods, allowMethods,
				"Access-Control-Allow-Methods mismatch",
			)

			// Check the Access-Control-Allow-Headers header.
			allowHeaders := rr.Header().Get("Access-Control-Allow-Headers")
			// Since order might matter, we can compare after splitting.
			assert.Equal(
				s.T(),
				strings.Split(tc.expectedAllowHeaders, ","),
				strings.Split(allowHeaders, ","),
				"Access-Control-Allow-Headers mismatch",
			)

			// Check the Access-Control-Allow-Credentials header.
			allowCred := rr.Header().Get("Access-Control-Allow-Credentials")
			assert.Equal(s.T(), tc.expectedAllowCredentials, allowCred,
				"Access-Control-Allow-Credentials mismatch",
			)
		})
	}
}
