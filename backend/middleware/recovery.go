package middleware

import (
	"log/slog"
	"net/http"

	"flowdb/backend/util"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					appErr := util.NewAppError("internal_error", nil)
					logger.Error("panic", "error_id", appErr.ID, "panic", rec)
					http.Error(w, "internal error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
