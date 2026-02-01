package stream

import (
	"time"

	"github.com/gorilla/websocket"
)

func SendStart(conn *websocket.Conn, queryID string) error {
	return conn.WriteJSON(map[string]any{
		"type":      "start",
		"queryId":   queryID,
		"startedAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func SendSchema(conn *websocket.Conn, columns []map[string]any) error {
	return conn.WriteJSON(map[string]any{
		"type":    "schema",
		"columns": columns,
	})
}

func SendFields(conn *websocket.Conn, fields []string) error {
	return conn.WriteJSON(map[string]any{
		"type":   "schema",
		"fields": fields,
	})
}

func SendRows(conn *websocket.Conn, rows []any) error {
	return conn.WriteJSON(map[string]any{
		"type": "rows",
		"rows": rows,
	})
}

func SendEnd(conn *websocket.Conn, rowCount int, durationMs int64) error {
	return conn.WriteJSON(map[string]any{
		"type":       "end",
		"rowCount":   rowCount,
		"durationMs": durationMs,
	})
}

func SendError(conn *websocket.Conn, message string, errorID string) error {
	return conn.WriteJSON(map[string]any{
		"type":    "error",
		"message": message,
		"errorId": errorID,
	})
}
