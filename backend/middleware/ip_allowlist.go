package middleware

import (
	"net"
	"net/http"
)

type IPAllowlist struct {
	Networks []*net.IPNet
}

func NewIPAllowlist(cidrs []string) *IPAllowlist {
	allow := &IPAllowlist{}
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil || network == nil {
			continue
		}
		allow.Networks = append(allow.Networks, network)
	}
	return allow
}

func (a *IPAllowlist) Contains(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, n := range a.Networks {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
}

func IPAllowlistMiddleware(allowlist *IPAllowlist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			if !allowlist.Contains(ip) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
