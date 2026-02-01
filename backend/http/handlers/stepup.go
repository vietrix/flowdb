package handlers

import (
	"net/http"
	"time"

	"flowdb/backend/auth"
)

func (h *Handler) requireStepUp(w http.ResponseWriter, r *http.Request) bool {
	if !h.Settings.Get().FlagEnabled("enable_step_up_auth") {
		return true
	}
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	if time.Since(session.LastAuthAt) > h.Config.StepUpMaxAge {
		http.Error(w, "step up required", http.StatusUnauthorized)
		return false
	}
	user, ok := auth.UserFromContext(r.Context())
	if ok && user.MFAEnabled {
		if session.LastMFAAt == nil || time.Since(*session.LastMFAAt) > h.Config.StepUpMaxAge {
			http.Error(w, "mfa step up required", http.StatusUnauthorized)
			return false
		}
	}
	return true
}
