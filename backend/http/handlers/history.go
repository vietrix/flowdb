package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
)

func (h *Handler) ListHistory(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "query:read", "history", "") {
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 100)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	list, err := h.Store.ListHistory(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "failed to list history", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "audit:read", "audit", "") {
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 100)
	offset := parseInt(r.URL.Query().Get("offset"), 0)
	entries, err := h.Store.ListAudit(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "failed to list audit", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("ndjson") == "true" {
		w.Header().Set("Content-Type", "application/x-ndjson")
		writer := bufio.NewWriter(w)
		for _, entry := range entries {
			line, _ := json.Marshal(entry)
			_, _ = writer.Write(line)
			_, _ = writer.Write([]byte("\n"))
		}
		_ = writer.Flush()
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
