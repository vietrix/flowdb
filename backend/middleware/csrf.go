package middleware

import (
	"net/http"
	"strings"

	"flowdb/backend/auth"
)

func CSRF(headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			sess, ok := auth.SessionFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := r.Header.Get(headerName)
			if token == "" {
				token = r.Header.Get(strings.ToLower(headerName))
			}
			if token == "" || token != sess.CSRFToken {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
