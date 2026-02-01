package middleware

import "net/http"

func MTLS(headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(headerName) != "true" {
				http.Error(w, "mTLS required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
