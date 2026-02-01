package middleware

import (
	"net/http"
	"time"

	"flowdb/backend/auth"
	"flowdb/backend/store"
)

func RequireAuth(st *store.Store, sessions *auth.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessID, ok := sessions.Get(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			sess, err := st.GetSession(r.Context(), sessID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if time.Now().UTC().After(sess.ExpiresAt) {
				_ = st.DeleteSession(r.Context(), sessID)
				http.Error(w, "session expired", http.StatusUnauthorized)
				return
			}
			user, err := st.GetUserByID(r.Context(), sess.UserID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := auth.WithUser(r.Context(), user)
			ctx = auth.WithSession(ctx, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
