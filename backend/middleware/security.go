package middleware

import (
	"net"
	"net/http"

	"flowdb/backend/settings"
)

func SecurityControls(store *settings.Store, mtlsHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current := store.Get()
			if current.FlagEnabled("enable_ip_allowlist") {
				allow := NewIPAllowlist(current.IPAllowlist)
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				if ip == "" {
					ip = r.RemoteAddr
				}
				if !allow.Contains(ip) {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
			if current.FlagEnabled("enable_mtls") {
				if r.Header.Get(mtlsHeader) != "true" {
					http.Error(w, "mTLS required", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
