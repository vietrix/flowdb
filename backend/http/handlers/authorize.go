package handlers

import (
	"net/http"

	"flowdb/backend/auth"
	"flowdb/backend/policies"
)

func (h *Handler) authorize(w http.ResponseWriter, r *http.Request, action string, resource string, env string) bool {
	_, ok := h.authorizeWithConstraints(w, r, action, resource, env)
	return ok
}

func (h *Handler) authorizeWithConstraints(w http.ResponseWriter, r *http.Request, action string, resource string, env string) (policies.Constraints, bool) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return policies.Constraints{}, false
	}
	decision, err := h.Authorizer.Authorize(r.Context(), user, action, resource, env)
	if err != nil {
		http.Error(w, "authorization error", http.StatusInternalServerError)
		return policies.Constraints{}, false
	}
	if !decision.Allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return policies.Constraints{}, false
	}
	return decision.Constraints, true
}
