package handlers

import (
	"encoding/json"
	"net/http"

	"flowdb/backend/util"
)

func decodeJSON(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeAppError(w http.ResponseWriter, status int, err error) string {
	appErr, ok := err.(*util.AppError)
	if !ok {
		appErr = util.NewAppError("internal error", err)
	}
	http.Error(w, appErr.Message, status)
	return appErr.ID
}
