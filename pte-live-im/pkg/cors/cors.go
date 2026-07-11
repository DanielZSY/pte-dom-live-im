package cors

import (
	"net/http"
	"net/url"
	"strings"

	"pte_live_im/pkg/setting"
)

// Middleware wraps HTTP handlers with CORS headers and OPTIONS preflight.
func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ApplyHeaders(w, r)
		if r.Method == http.MethodOptions {
			if IsRequestAllowed(r) {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}
		next(w, r)
	}
}

// CheckWebSocketOrigin validates browser WebSocket upgrade; non-browser clients without Origin are allowed.
func CheckWebSocketOrigin(r *http.Request) bool {
	if !setting.CORSSetting.Enabled {
		return true
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	return IsOriginAllowed(origin)
}

// ApplyHeaders sets CORS response headers for browser clients.
func ApplyHeaders(w http.ResponseWriter, r *http.Request) {
	cfg := setting.CORSSetting
	if !cfg.Enabled {
		return
	}

	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" && IsOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Vary", "Origin")
	}

	w.Header().Set("Access-Control-Allow-Methods", cfg.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", cfg.AllowHeaders)
	w.Header().Set("Access-Control-Expose-Headers", cfg.ExposeHeaders)
}

// IsRequestAllowed reports whether the request Origin is permitted.
func IsRequestAllowed(r *http.Request) bool {
	cfg := setting.CORSSetting
	if !cfg.Enabled {
		return true
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	return IsOriginAllowed(origin)
}

// IsOriginAllowed checks configured allowOrigins (supports "*" and "*.domain").
func IsOriginAllowed(origin string) bool {
	cfg := setting.CORSSetting
	if !cfg.Enabled {
		return false
	}
	if originListAllowsAll(cfg.AllowOrigins) {
		return true
	}
	for _, item := range cfg.AllowOrigins {
		if matchOrigin(origin, strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func originListAllowsAll(origins []string) bool {
	for _, o := range origins {
		if strings.TrimSpace(o) == "*" {
			return true
		}
	}
	return false
}

func matchOrigin(origin, pattern string) bool {
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "https" {
			return false
		}
		return strings.HasSuffix(u.Hostname(), suffix)
	}
	return origin == pattern
}
