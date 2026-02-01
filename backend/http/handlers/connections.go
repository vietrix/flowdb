package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"flowdb/backend/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type connectionRequest struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Host     string         `json:"host"`
	Port     int            `json:"port"`
	Database string         `json:"database"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	TLS      map[string]any `json:"tls"`
	Tags     map[string]any `json:"tags"`
}

func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	conns, err := h.Store.ListConnections(r.Context())
	if err != nil {
		http.Error(w, "failed to list connections", http.StatusInternalServerError)
		return
	}
	filtered := make([]store.Connection, 0, len(conns))
	for _, conn := range conns {
		env := getEnv(conn.Tags)
		if h.authorize(w, r, "connection:read", "connection/"+conn.ID.String(), env) {
			filtered = append(filtered, conn)
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r, "connection:write", "connection/*", "") {
		return
	}
	var req connectionRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	secretBytes, err := json.Marshal(map[string]any{"password": req.Password})
	if err != nil {
		http.Error(w, "invalid secret", http.StatusBadRequest)
		return
	}
	enc, err := h.Cipher.Encrypt(secretBytes)
	if err != nil {
		http.Error(w, "failed to encrypt secret", http.StatusInternalServerError)
		return
	}
	secretID, err := h.Store.CreateConnectionSecret(r.Context(), enc)
	if err != nil {
		http.Error(w, "failed to store secret", http.StatusInternalServerError)
		return
	}
	conn, err := h.Store.CreateConnection(r.Context(), store.Connection{
		Name:      req.Name,
		Type:      req.Type,
		Host:      req.Host,
		Port:      req.Port,
		Database:  req.Database,
		Username:  req.Username,
		SecretRef: secretID,
		TLS:       req.TLS,
		Tags:      req.Tags,
	})
	if err != nil {
		http.Error(w, "failed to create connection", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "connection_create", nil, map[string]any{"id": conn.ID.String(), "name": conn.Name}, "")
	writeJSON(w, http.StatusCreated, conn)
}

func (h *Handler) GetConnection(w http.ResponseWriter, r *http.Request) {
	connID := chi.URLParam(r, "id")
	id, err := uuid.Parse(connID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	conn, err := h.Store.GetConnection(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:read", "connection/"+conn.ID.String(), env) {
		return
	}
	writeJSON(w, http.StatusOK, conn)
}

func (h *Handler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	connID := chi.URLParam(r, "id")
	id, err := uuid.Parse(connID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	conn, err := h.Store.GetConnection(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:write", "connection/"+conn.ID.String(), env) {
		return
	}
	if !h.requireStepUp(w, r) {
		return
	}
	var req connectionRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	conn.Name = req.Name
	conn.Type = req.Type
	conn.Host = req.Host
	conn.Port = req.Port
	conn.Database = req.Database
	conn.Username = req.Username
	conn.TLS = req.TLS
	conn.Tags = req.Tags
	if err := h.Store.UpdateConnection(r.Context(), conn); err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	if req.Password != "" {
		secretBytes, err := json.Marshal(map[string]any{"password": req.Password})
		if err != nil {
			http.Error(w, "invalid secret", http.StatusBadRequest)
			return
		}
		enc, err := h.Cipher.Encrypt(secretBytes)
		if err != nil {
			http.Error(w, "failed to encrypt secret", http.StatusInternalServerError)
			return
		}
		if err := h.Store.UpdateConnectionSecret(r.Context(), conn.SecretRef, enc); err != nil {
			http.Error(w, "failed to update secret", http.StatusInternalServerError)
			return
		}
	}
	_ = h.Audit.LogEvent(r.Context(), "connection_update", nil, map[string]any{"id": conn.ID.String()}, "")
	writeJSON(w, http.StatusOK, conn)
}

func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	connID := chi.URLParam(r, "id")
	id, err := uuid.Parse(connID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	conn, err := h.Store.GetConnection(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:write", "connection/"+conn.ID.String(), env) {
		return
	}
	if !h.requireStepUp(w, r) {
		return
	}
	if err := h.Store.DeleteConnection(r.Context(), id); err != nil {
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "connection_delete", nil, map[string]any{"id": conn.ID.String()}, "")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	connID := chi.URLParam(r, "id")
	id, err := uuid.Parse(connID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	conn, err := h.Store.GetConnection(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:read", "connection/"+conn.ID.String(), env) {
		return
	}
	adapter, err := h.Connections.GetAdapter(r.Context(), conn)
	if err != nil {
		http.Error(w, "failed to connect", http.StatusBadRequest)
		return
	}
	defer adapter.Close()
	_, err = adapter.ListNamespaces(r.Context())
	if err != nil {
		http.Error(w, "connection failed", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func getEnv(tags map[string]any) string {
	if tags == nil {
		return ""
	}
	if env, ok := tags["environment"].(string); ok {
		return env
	}
	if env, ok := tags["env"].(string); ok {
		return env
	}
	return ""
}

func parseInt(value string, def int) int {
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}
