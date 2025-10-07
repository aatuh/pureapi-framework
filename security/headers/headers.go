package headers

import "net/http"

// Config controls which security headers are applied.
type Config struct {
	NoSniff        bool
	FrameOptions   string
	XSSProtection  string
	ReferrerPolicy string
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		NoSniff:        true,
		FrameOptions:   "DENY",
		XSSProtection:  "0",
		ReferrerPolicy: "no-referrer",
	}
}

// Middleware applies the configured security headers.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.NoSniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}
			if cfg.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}
			if cfg.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Default returns a middleware populated with DefaultConfig.
func Default() func(http.Handler) http.Handler {
	return Middleware(DefaultConfig())
}
