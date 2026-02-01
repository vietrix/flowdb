package handlers

import (
	"net/http"

	"flowdb/backend/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) ListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_query_approval") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !h.authorize(w, r, "query:admin", "approvals", "") {
		return
	}
	list, err := h.Store.ListPendingApprovals(r.Context())
	if err != nil {
		http.Error(w, "failed to list", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_query_approval") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !h.authorize(w, r, "query:admin", "approvals", "") {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	if err := h.Store.ApproveQuery(r.Context(), id, user.ID); err != nil {
		http.Error(w, "failed to approve", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "query_approval_approved", &user.ID, map[string]any{"approvalId": id.String()}, "")
	writeJSON(w, http.StatusOK, map[string]any{"status": "approved"})
}

func (h *Handler) Deny(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_query_approval") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !h.authorize(w, r, "query:admin", "approvals", "") {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	if err := h.Store.DenyQuery(r.Context(), id, user.ID); err != nil {
		http.Error(w, "failed to deny", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "query_approval_denied", &user.ID, map[string]any{"approvalId": id.String()}, "")
	writeJSON(w, http.StatusOK, map[string]any{"status": "denied"})
}
