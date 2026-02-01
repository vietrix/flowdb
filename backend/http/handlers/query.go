package handlers

import (
	"net/http"
	"time"

	"flowdb/backend/adapters"
	"flowdb/backend/auth"
	"flowdb/backend/query"
	"flowdb/backend/store"
	"flowdb/backend/stream"
	"flowdb/backend/util"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type queryRequest struct {
	Statement  string `json:"statement"`
	ApprovalID string `json:"approvalId"`
	MaxRows    int    `json:"maxRows"`
	TimeoutMs  int    `json:"timeoutMs"`
}

type queryResponse struct {
	QueryID    string `json:"queryId"`
	Status     string `json:"status"`
	ApprovalID string `json:"approvalId,omitempty"`
}

func (h *Handler) StartQuery(w http.ResponseWriter, r *http.Request) {
	conn, err := h.parseConnection(r)
	if err != nil {
		http.Error(w, "invalid connection", http.StatusBadRequest)
		return
	}
	var req queryRequest
	if err := decodeJSON(r, &req); err != nil || req.Statement == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	env := getEnv(conn.Tags)
	action := "query:read"
	isWrite := false
	if conn.Type == "postgres" {
		isWrite = query.IsSQLWrite(req.Statement)
	} else if conn.Type == "mongodb" {
		isWrite = query.IsMongoWrite(req.Statement)
	}
	if isWrite {
		action = "query:write"
	}
	resource := "connection/" + conn.ID.String() + "/db/*"
	constraints, ok := h.authorizeWithConstraints(w, r, action, resource, env)
	if !ok {
		return
	}
	if constraints.ReadOnly && isWrite {
		http.Error(w, "read only", http.StatusForbidden)
		return
	}
	if constraints.RequireWhere && isWrite && conn.Type == "postgres" && !query.HasWhere(req.Statement) {
		http.Error(w, "where required", http.StatusForbidden)
		return
	}
	user, _ := auth.UserFromContext(r.Context())
	if user.ID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !user.IsAdmin && conn.Type == "postgres" && query.IsDangerous(req.Statement) {
		http.Error(w, "operation not allowed", http.StatusForbidden)
		return
	}
	if isWrite && isProd(env) && !h.requireStepUp(w, r) {
		return
	}
	if h.Settings.Get().FlagEnabled("enable_query_approval") && isWrite && isProd(env) {
		if req.ApprovalID == "" {
			approval, err := h.Store.CreateQueryApproval(r.Context(), store.QueryApproval{
				ConnectionID: conn.ID,
				UserID:       user.ID,
				Statement:    req.Statement,
				Status:       "pending",
				Environment:  env,
			})
			if err != nil {
				http.Error(w, "failed to create approval", http.StatusInternalServerError)
				return
			}
			_, _ = h.Store.CreateQueryHistory(r.Context(), store.QueryHistory{
				UserID:        user.ID,
				ConnectionID:  conn.ID,
				StatementHash: query.StatementHash(req.Statement),
				Status:        "pending_approval",
				StartedAt:     time.Now().UTC(),
				Action:        action,
				Resource:      resource,
				ApprovalID:    &approval.ID,
			})
			_ = h.Audit.LogEvent(r.Context(), "query_approval_requested", &user.ID, map[string]any{"approvalId": approval.ID.String()}, "")
			writeJSON(w, http.StatusAccepted, queryResponse{
				Status:     "pending_approval",
				ApprovalID: approval.ID.String(),
			})
			return
		}
		approvalID, err := uuid.Parse(req.ApprovalID)
		if err != nil {
			http.Error(w, "invalid approval id", http.StatusBadRequest)
			return
		}
		approval, err := h.Store.GetQueryApproval(r.Context(), approvalID)
		if err != nil || approval.Status != "approved" {
			http.Error(w, "approval required", http.StatusForbidden)
			return
		}
	}
	maxRows := h.Config.GlobalMaxRows
	if constraints.MaxRows > 0 && constraints.MaxRows < maxRows {
		maxRows = constraints.MaxRows
	}
	if req.MaxRows > 0 && req.MaxRows < maxRows {
		maxRows = req.MaxRows
	}
	timeoutMs := int(h.Config.StatementTimeout.Milliseconds())
	if constraints.TimeoutMs > 0 && constraints.TimeoutMs < timeoutMs {
		timeoutMs = constraints.TimeoutMs
	}
	if req.TimeoutMs > 0 && req.TimeoutMs < timeoutMs {
		timeoutMs = req.TimeoutMs
	}
	jobID := h.JobStore.Create(query.Job{
		ConnectionID: conn.ID,
		Statement:    req.Statement,
		Action:       action,
		Resource:     resource,
		UserID:       user.ID,
		ApprovalID:   parseApprovalID(req.ApprovalID),
		Options: query.Options{
			MaxRows:      maxRows,
			TimeoutMs:    timeoutMs,
			ReadOnly:     constraints.ReadOnly,
			RequireWhere: constraints.RequireWhere,
		},
	})
	writeJSON(w, http.StatusOK, queryResponse{
		QueryID: jobID,
		Status:  "ready",
	})
}

func (h *Handler) StreamQuery(w http.ResponseWriter, r *http.Request) {
	connID := chi.URLParam(r, "id")
	_ = connID
	queryID := chi.URLParam(r, "queryId")
	job, ok := h.JobStore.Get(queryID)
	if !ok {
		http.Error(w, "query not found", http.StatusNotFound)
		return
	}
	conn, err := h.Store.GetConnection(r.Context(), job.ConnectionID)
	if err != nil {
		http.Error(w, "connection not found", http.StatusNotFound)
		return
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.Stream.Register(ws)
	defer func() {
		h.Stream.Unregister(ws)
		_ = ws.Close()
		h.JobStore.Delete(queryID)
	}()
	start := time.Now()
	if err := stream.SendStart(ws, queryID); err != nil {
		return
	}
	adapter, err := h.Connections.GetAdapter(r.Context(), conn)
	if err != nil {
		_ = stream.SendError(ws, "connection failed", util.NewAppError("connection failed", err).ID)
		return
	}
	defer adapter.Close()
	opts := adapters.QueryOptions{
		MaxRows: job.Options.MaxRows,
		Timeout: time.Duration(job.Options.TimeoutMs) * time.Millisecond,
	}
	statement := job.Statement
	if conn.Type == "postgres" && job.Options.MaxRows > 0 {
		statement = query.EnforceLimit(statement, job.Options.MaxRows)
	}
	history := store.QueryHistory{
		UserID:        job.UserID,
		ConnectionID:  conn.ID,
		StatementHash: query.StatementHash(job.Statement),
		Status:        "running",
		RowCount:      0,
		DurationMs:    0,
		StartedAt:     time.Now().UTC(),
		Action:        job.Action,
		Resource:      job.Resource,
		ApprovalID:    job.ApprovalID,
	}
	history, _ = h.Store.CreateQueryHistory(r.Context(), history)
	_ = h.Audit.LogEvent(r.Context(), "query_start", &job.UserID, map[string]any{"queryId": queryID}, "")
	result, err := adapter.Query(r.Context(), statement, opts)
	if err != nil {
		_ = stream.SendError(ws, "query failed", util.NewAppError("query failed", err).ID)
		history.Status = "failed"
		history.EndedAt = timePtr(time.Now().UTC())
		_ = h.Store.UpdateQueryHistory(r.Context(), history)
		return
	}
	defer drainStream(result)
	columns := result.Columns
	if len(columns) > 0 {
		colMeta := make([]map[string]any, 0, len(columns))
		for _, c := range columns {
			colMeta = append(colMeta, map[string]any{"name": c.Name, "type": c.Type})
		}
		_ = stream.SendSchema(ws, colMeta)
	}
	rowCount := 0
	if h.Settings.Get().FlagEnabled("enable_pii_masking") {
		rules, _ := h.Store.ListPIIRules(r.Context(), conn.ID)
		colNames := make([]string, 0, len(columns))
		for _, c := range columns {
			colNames = append(colNames, c.Name)
		}
		for row := range result.Rows {
			rowCount++
			masked := query.MaskRow(job.Resource, colNames, row, rules)
			_ = stream.SendRows(ws, []any{masked})
		}
		firstDoc := true
		for doc := range result.Docs {
			rowCount++
			if firstDoc {
				fields := make([]string, 0, len(doc))
				for k := range doc {
					fields = append(fields, k)
				}
				_ = stream.SendFields(ws, fields)
				firstDoc = false
			}
			masked := query.MaskDoc(job.Resource, doc, rules)
			_ = stream.SendRows(ws, []any{masked})
		}
	} else {
		for row := range result.Rows {
			rowCount++
			_ = stream.SendRows(ws, []any{row})
		}
		firstDoc := true
		for doc := range result.Docs {
			rowCount++
			if firstDoc {
				fields := make([]string, 0, len(doc))
				for k := range doc {
					fields = append(fields, k)
				}
				_ = stream.SendFields(ws, fields)
				firstDoc = false
			}
			_ = stream.SendRows(ws, []any{doc})
		}
	}
	duration := time.Since(start).Milliseconds()
	_ = stream.SendEnd(ws, rowCount, duration)
	history.Status = "completed"
	history.RowCount = rowCount
	history.DurationMs = duration
	history.EndedAt = timePtr(time.Now().UTC())
	_ = h.Store.UpdateQueryHistory(r.Context(), history)
	_ = h.Audit.LogEvent(r.Context(), "query_end", &job.UserID, map[string]any{"queryId": queryID, "rows": rowCount}, "")
	select {
	case err := <-result.Err:
		if err != nil {
			_ = stream.SendError(ws, "stream error", util.NewAppError("stream error", err).ID)
		}
	default:
	}
}

type explainRequest struct {
	Statement string `json:"statement"`
}

func (h *Handler) ExplainQuery(w http.ResponseWriter, r *http.Request) {
	conn, adapter, ok := h.getConnectionAdapter(w, r)
	if !ok {
		return
	}
	defer adapter.Close()
	var req explainRequest
	if err := decodeJSON(r, &req); err != nil || req.Statement == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	env := getEnv(conn.Tags)
	resource := "connection/" + conn.ID.String() + "/db/*"
	if !h.authorize(w, r, "query:read", resource, env) {
		return
	}
	result, err := adapter.Explain(r.Context(), req.Statement)
	if err != nil {
		http.Error(w, "explain failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) parseConnection(r *http.Request) (store.Connection, error) {
	connID := chi.URLParam(r, "id")
	id, err := uuid.Parse(connID)
	if err != nil {
		return store.Connection{}, err
	}
	return h.Store.GetConnection(r.Context(), id)
}

func isProd(env string) bool {
	switch env {
	case "prod", "production":
		return true
	default:
		return false
	}
}

func parseApprovalID(value string) *uuid.UUID {
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return &id
}

func timePtr(t time.Time) *time.Time {
	return &t
}
