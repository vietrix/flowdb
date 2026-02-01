package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowOrigins []string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, origin := range allowOrigins {
		allowed[origin] = struct{}{}
	}
	allowAll := len(allowOrigins) == 1 && allowOrigins[0] == "*"
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowAll || originAllowed(origin, allowed)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed map[string]struct{}) bool {
	_, ok := allowed[origin]
	if ok {
		return true
	}
	for o := range allowed {
		if strings.HasPrefix(o, "*.") {
			suffix := strings.TrimPrefix(o, "*")
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
	}
	return false
}
