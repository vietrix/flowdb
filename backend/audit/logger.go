package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"time"

	"flowdb/backend/settings"
	"flowdb/backend/store"
	"flowdb/backend/util"

	"github.com/google/uuid"
)

type Logger struct {
	store    *store.Store
	settings *settings.Store
}

func NewLogger(st *store.Store, settings *settings.Store) *Logger {
	return &Logger{store: st, settings: settings}
}

func (l *Logger) LogEvent(ctx context.Context, eventType string, userID *uuid.UUID, details map[string]any, errorID string) error {
	payload, err := util.CanonicalJSON(details)
	if err != nil {
		return err
	}
	useHash := l.settings.Get().FlagEnabled("enable_signed_audit_log")
	prevHash := ""
	currentHash := ""
	if useHash {
		prevHash, err = l.lastHash(ctx)
		if err != nil {
			return err
		}
		currentHash = computeHash(prevHash, payload)
	}
	_, err = l.store.DB().Exec(ctx, `
		INSERT INTO audit_log (id, event_type, actor_user_id, details, created_at, prev_hash, hash, error_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, uuid.New(), eventType, userID, payload, time.Now().UTC(), prevHash, currentHash, errorID)
	return err
}

func (l *Logger) lastHash(ctx context.Context) (string, error) {
	var hashValue string
	err := l.store.DB().QueryRow(ctx, `
		SELECT hash FROM audit_log ORDER BY created_at DESC LIMIT 1
	`).Scan(&hashValue)
	if err != nil {
		return "", nil
	}
	return hashValue, nil
}

func computeHash(prevHash string, payload []byte) string {
	var h hash.Hash = sha256.New()
	h.Write([]byte(prevHash))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
