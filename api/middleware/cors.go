package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/aatuh/pureapi-core/endpoint"
)

// CORSOptions encapsulates configuration options for the CORS middleware.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// CORS creates a new CORS middleware based on CORSOptions.
//
// The middleware inspects the incoming request's "Origin" header and, if the
// origin is in the allowed list, it sets the appropriate CORS headers on the
// response. When AllowCredentials is enabled, the middleware echoes the
// request’s origin when the origin matches the allowed list. It also sets the
// allowed HTTP methods and headers as configured.
//
// Parameters:
//   - opts: CORSOptions containing configuration for allowed origins, methods,
//     headers, and whether credentials are permitted.
//
// Returns:
//   - api.Middleware: A middleware function that applies the CORS
//     configuration.
func CORS(opts CORSOptions) endpoint.Middleware {
	// Check if "*" is allowed.
	isWildcard := slices.Contains(opts.AllowedOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setOriginHeader(
				w, r, opts.AllowedOrigins, isWildcard, opts.AllowCredentials,
			)
			setPreflightHeaders(w, opts.AllowedMethods, opts.AllowedHeaders)
			next.ServeHTTP(w, r)
		})
	}
}

// setOriginHeader sets the "Access-Control-Allow-Origin" header based on the
// request origin, allowed origins, and whether credentials are allowed.
func setOriginHeader(
	w http.ResponseWriter,
	r *http.Request,
	allowedOrigins []string,
	isWildcard bool,
	allowCredentials bool,
) {
	origin := r.Header.Get("Origin")
	if allowCredentials {
		if origin != "" && slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		if isWildcard {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
}

// setPreflightHeaders sets the CORS headers for preflight requests.
func setPreflightHeaders(
	w http.ResponseWriter, allowedMethods []string, allowedHeaders []string,
) {
	w.Header().Set(
		"Access-Control-Allow-Methods", strings.Join(allowedMethods, ","),
	)

	// Concatenate "Content-Type" with the allowed headers.
	allHeaders := slices.Concat([]string{"Content-Type"}, allowedHeaders)
	w.Header().Set(
		"Access-Control-Allow-Headers", strings.Join(allHeaders, ","),
	)
}
