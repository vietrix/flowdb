package middleware

import (
	"net/http"
	"time"

	"flowdb/backend/auth"
	"flowdb/backend/settings"
)

func StepUp(store *settings.Store, maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current := store.Get()
			if !current.FlagEnabled("enable_step_up_auth") {
				next.ServeHTTP(w, r)
				return
			}
			session, ok := auth.SessionFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if time.Since(session.LastAuthAt) > maxAge {
				http.Error(w, "step up required", http.StatusUnauthorized)
				return
			}
			user, ok := auth.UserFromContext(r.Context())
			if ok && user.MFAEnabled {
				if session.LastMFAAt == nil || time.Since(*session.LastMFAAt) > maxAge {
					http.Error(w, "mfa step up required", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
