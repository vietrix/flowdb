package handlers

import (
	"net/http"
	"os"
	"time"

	"flowdb/backend/util"
)

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.Store.DB().Ping(ctx); err != nil {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":   util.Version,
		"commit":    util.CommitSHA,
		"buildTime": util.BuildTime,
	})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "settings:write", "settings/*", "") {
		return
	}
	ctx := r.Context()
	status, err := h.Update.Status(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"repo":            status.Repo,
			"currentVersion":  status.CurrentVersion,
			"updateAvailable": false,
			"checkedAt":       time.Now().UTC(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"repo":            status.Repo,
		"currentVersion":  status.CurrentVersion,
		"latest":          status.Latest,
		"updateAvailable": status.UpdateAvailable,
		"checkedAt":       time.Now().UTC(),
	})
}

type updateApplyRequest struct {
	Tag string `json:"tag"`
}

func (h *Handler) UpdateApply(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "settings:write", "settings/*", "") {
		return
	}
	var req updateApplyRequest
	_ = decodeJSON(r, &req)
	result, err := h.Update.ApplyUpdate(r.Context(), req.Tag)
	if err != nil {
		http.Error(w, "update failed", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"applied":        result.Applied,
		"release":        result.Info,
		"willAutoRestart": h.Config.UpdateAutoRestart,
	})
	if h.Config.UpdateAutoRestart {
		go func() {
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}()
	}
}
