package handlers

import (
	"net/http"

	"flowdb/backend/adapters"
	"flowdb/backend/query"
	"flowdb/backend/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	conn, adapter, ok := h.getConnectionAdapter(w, r)
	if !ok {
		return
	}
	defer adapter.Close()
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:read", "connection/"+conn.ID.String(), env) {
		return
	}
	list, err := adapter.ListNamespaces(r.Context())
	if err != nil {
		http.Error(w, "failed to list namespaces", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	conn, adapter, ok := h.getConnectionAdapter(w, r)
	if !ok {
		return
	}
	defer adapter.Close()
	ns := r.URL.Query().Get("ns")
	if ns == "" {
		http.Error(w, "ns required", http.StatusBadRequest)
		return
	}
	resource := "connection/" + conn.ID.String() + "/db/" + ns + "/entity/*"
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:read", resource, env) {
		return
	}
	list, err := adapter.ListEntities(r.Context(), ns)
	if err != nil {
		http.Error(w, "failed to list entities", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) GetEntityInfo(w http.ResponseWriter, r *http.Request) {
	conn, adapter, ok := h.getConnectionAdapter(w, r)
	if !ok {
		return
	}
	defer adapter.Close()
	ns := r.URL.Query().Get("ns")
	name := chi.URLParam(r, "name")
	if ns == "" || name == "" {
		http.Error(w, "ns and name required", http.StatusBadRequest)
		return
	}
	resource := "connection/" + conn.ID.String() + "/db/" + ns + "/entity/" + name
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "connection:read", resource, env) {
		return
	}
	info, err := adapter.GetEntityInfo(r.Context(), ns, name)
	if err != nil {
		http.Error(w, "failed to get entity info", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *Handler) BrowseEntity(w http.ResponseWriter, r *http.Request) {
	conn, adapter, ok := h.getConnectionAdapter(w, r)
	if !ok {
		return
	}
	defer adapter.Close()
	ns := r.URL.Query().Get("ns")
	name := chi.URLParam(r, "name")
	if ns == "" || name == "" {
		http.Error(w, "ns and name required", http.StatusBadRequest)
		return
	}
	resource := "connection/" + conn.ID.String() + "/db/" + ns + "/entity/" + name
	env := getEnv(conn.Tags)
	if !h.authorize(w, r, "query:read", resource, env) {
		return
	}
	opts := adapters.BrowseOptions{
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("pageSize"), 100),
		Sort:     r.URL.Query().Get("sort"),
		Filter:   r.URL.Query().Get("filter"),
	}
	stream, err := adapter.Browse(r.Context(), ns, name, opts)
	if err != nil {
		http.Error(w, "browse failed", http.StatusInternalServerError)
		return
	}
	defer drainStream(stream)
	response := map[string]any{
		"columns": stream.Columns,
		"rows":    []any{},
		"docs":    []any{},
	}
	if h.Settings.Get().FlagEnabled("enable_pii_masking") {
		rules, _ := h.Store.ListPIIRules(r.Context(), conn.ID)
		columns := make([]string, 0, len(stream.Columns))
		for _, c := range stream.Columns {
			columns = append(columns, c.Name)
		}
		for row := range stream.Rows {
			masked := query.MaskRow(resource, columns, row, rules)
			response["rows"] = append(response["rows"].([]any), masked)
		}
		for doc := range stream.Docs {
			masked := query.MaskDoc(resource, doc, rules)
			response["docs"] = append(response["docs"].([]any), masked)
		}
	} else {
		for row := range stream.Rows {
			response["rows"] = append(response["rows"].([]any), row)
		}
		for doc := range stream.Docs {
			response["docs"] = append(response["docs"].([]any), doc)
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) getConnectionAdapter(w http.ResponseWriter, r *http.Request) (store.Connection, adapters.Adapter, bool) {
	connID := chi.URLParam(r, "id")
	id, err := uuid.Parse(connID)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return store.Connection{}, nil, false
	}
	conn, err := h.Store.GetConnection(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return store.Connection{}, nil, false
	}
	adapter, err := h.Connections.GetAdapter(r.Context(), conn)
	if err != nil {
		http.Error(w, "failed to connect", http.StatusBadRequest)
		return store.Connection{}, nil, false
	}
	return conn, adapter, true
}

func drainStream(stream *adapters.ResultStream) {
	for range stream.Rows {
	}
	for range stream.Docs {
	}
}
