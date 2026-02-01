package auth

import (
	"context"
	"net/http"
	"time"

	"flowdb/backend/store"
	"flowdb/backend/util"

	"github.com/google/uuid"
)

type SessionManager struct {
	store      *store.Store
	cookieName string
	ttl        time.Duration
}

func NewSessionManager(st *store.Store, cookieName string, ttl time.Duration) *SessionManager {
	return &SessionManager{store: st, cookieName: cookieName, ttl: ttl}
}

func (s *SessionManager) Create(ctx context.Context, userID uuid.UUID, mfaAt *time.Time) (store.Session, error) {
	csrf, err := util.RandomToken(32)
	if err != nil {
		return store.Session{}, err
	}
	now := time.Now().UTC()
	sess := store.Session{
		UserID:     userID,
		CSRFToken:  csrf,
		ExpiresAt:  now.Add(s.ttl),
		LastAuthAt: now,
		LastMFAAt:  mfaAt,
	}
	return s.store.CreateSession(ctx, sess)
}

func (s *SessionManager) Get(r *http.Request) (uuid.UUID, bool) {
	c, err := r.Cookie(s.cookieName)
	if err != nil {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(c.Value)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (s *SessionManager) SetCookie(w http.ResponseWriter, sessionID uuid.UUID, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *SessionManager) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
