package handlers

import (
	"net/http"
)

func (h *Handler) ListSCIMUsers(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_scim") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !h.authorize(w, r, "settings:write", "scim", "") {
		return
	}
	limit := parseInt(r.URL.Query().Get("count"), 100)
	start := parseInt(r.URL.Query().Get("startIndex"), 1)
	if start < 1 {
		start = 1
	}
	offset := start - 1
	users, err := h.Store.ListUsers(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}
	resources := make([]map[string]any, 0, len(users))
	for _, u := range users {
		resources = append(resources, map[string]any{
			"id":       u.ID.String(),
			"userName": u.Username,
			"active":   true,
			"schemas":  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": len(resources),
		"startIndex":   start,
		"itemsPerPage": len(resources),
		"Resources":    resources,
	})
}

func (h *Handler) ListSCIMGroups(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_scim") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !h.authorize(w, r, "settings:write", "scim", "") {
		return
	}
	limit := parseInt(r.URL.Query().Get("count"), 100)
	start := parseInt(r.URL.Query().Get("startIndex"), 1)
	if start < 1 {
		start = 1
	}
	offset := start - 1
	groups, err := h.Store.ListGroups(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "failed to list groups", http.StatusInternalServerError)
		return
	}
	resources := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		resources = append(resources, map[string]any{
			"id":          g.ID.String(),
			"displayName": g.Name,
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": len(resources),
		"startIndex":   start,
		"itemsPerPage": len(resources),
		"Resources":    resources,
	})
}
