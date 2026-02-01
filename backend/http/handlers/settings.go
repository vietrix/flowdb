package handlers

import (
	"net/http"
)

type settingsUpdateRequest struct {
	SecurityMode string          `json:"securityMode"`
	Flags        map[string]bool `json:"flags"`
	Config       map[string]any  `json:"config"`
	IPAllowlist  []string        `json:"ipAllowlist"`
}

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "settings:write", "settings", "") {
		return
	}
	if !h.requireStepUp(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, h.Settings.Get())
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "settings:write", "settings", "") {
		return
	}
	if !h.requireStepUp(w, r) {
		return
	}
	var req settingsUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	current := h.Settings.Get()
	if req.SecurityMode != "" {
		current.SecurityMode = req.SecurityMode
	}
	if req.Flags != nil {
		for k, v := range req.Flags {
			current.Flags[k] = v
		}
	}
	if req.Config != nil {
		current.Config = req.Config
	}
	if req.IPAllowlist != nil {
		current.IPAllowlist = req.IPAllowlist
	}
	if err := h.Settings.Update(r.Context(), current); err != nil {
		http.Error(w, "failed to update settings", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "settings_update", nil, map[string]any{"mode": current.SecurityMode}, "")
	writeJSON(w, http.StatusOK, current)
}

type modeUpdateRequest struct {
	Mode string `json:"mode"`
}

func (h *Handler) UpdateSecurityMode(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "settings:write", "settings", "") {
		return
	}
	if !h.requireStepUp(w, r) {
		return
	}
	var req modeUpdateRequest
	if err := decodeJSON(r, &req); err != nil || req.Mode == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.Settings.UpdateMode(r.Context(), req.Mode); err != nil {
		http.Error(w, "failed to update mode", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "settings_mode_update", nil, map[string]any{"mode": req.Mode}, "")
	writeJSON(w, http.StatusOK, h.Settings.Get())
}

type flagsUpdateRequest struct {
	Flags map[string]bool `json:"flags"`
}

func (h *Handler) UpdateFlags(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "settings:write", "settings", "") {
		return
	}
	var req flagsUpdateRequest
	if err := decodeJSON(r, &req); err != nil || req.Flags == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.Settings.UpdateFlags(r.Context(), req.Flags); err != nil {
		http.Error(w, "failed to update flags", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "settings_flags_update", nil, map[string]any{"flags": req.Flags}, "")
	writeJSON(w, http.StatusOK, h.Settings.Get())
}
