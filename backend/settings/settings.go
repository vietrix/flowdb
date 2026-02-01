package settings

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Settings struct {
	SecurityMode string
	Flags        map[string]bool
	Config       map[string]any
	IPAllowlist  []string
	UpdatedAt    time.Time
}

func (s Settings) FlagEnabled(name string) bool {
	if s.Flags == nil {
		return false
	}
	return s.Flags[name]
}

type Store struct {
	db      *pgxpool.Pool
	refresh time.Duration
	value   atomic.Value
}

func NewStore(db *pgxpool.Pool, refresh time.Duration) *Store {
	return &Store{db: db, refresh: refresh}
}

func (s *Store) Start(ctx context.Context) error {
	if err := s.LoadOrInit(ctx); err != nil {
		return err
	}
	ticker := time.NewTicker(s.refresh)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.Refresh(ctx)
			}
		}
	}()
	return nil
}

func (s *Store) LoadOrInit(ctx context.Context) error {
	set, err := s.load(ctx)
	if err == nil {
		s.value.Store(set)
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}
	defaults := Settings{
		SecurityMode: "basic",
		Flags: map[string]bool{
			"enable_sso_oidc":         false,
			"enable_sso_saml":         false,
			"enable_mfa":              false,
			"enable_step_up_auth":     false,
			"enable_ip_allowlist":     false,
			"enable_mtls":             false,
			"enable_signed_audit_log": false,
			"enable_query_approval":   false,
			"enable_pii_masking":      false,
			"enable_scim":             false,
		},
		Config:      map[string]any{},
		IPAllowlist: []string{},
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.insert(ctx, defaults); err != nil {
		return err
	}
	s.value.Store(defaults)
	return nil
}

func (s *Store) Refresh(ctx context.Context) error {
	set, err := s.load(ctx)
	if err != nil {
		return err
	}
	s.value.Store(set)
	return nil
}

func (s *Store) Get() Settings {
	val := s.value.Load()
	if val == nil {
		return Settings{}
	}
	return val.(Settings)
}

func (s *Store) Update(ctx context.Context, updated Settings) error {
	flags, err := json.Marshal(updated.Flags)
	if err != nil {
		return err
	}
	config, err := json.Marshal(updated.Config)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		UPDATE settings
		SET security_mode=$1, flags=$2, config=$3, ip_allowlist=$4, updated_at=now()
		WHERE id=1
	`, updated.SecurityMode, flags, config, updated.IPAllowlist)
	if err != nil {
		return err
	}
	updated.UpdatedAt = time.Now().UTC()
	s.value.Store(updated)
	return nil
}

func (s *Store) UpdateMode(ctx context.Context, mode string) error {
	current := s.Get()
	current.SecurityMode = mode
	return s.Update(ctx, current)
}

func (s *Store) UpdateFlags(ctx context.Context, flags map[string]bool) error {
	current := s.Get()
	for k, v := range flags {
		current.Flags[k] = v
	}
	return s.Update(ctx, current)
}

func (s *Store) ErrNotFound() error {
	return ErrNotFound
}

var ErrNotFound = errors.New("settings not found")

func (s *Store) load(ctx context.Context) (Settings, error) {
	var mode string
	var flagsData []byte
	var configData []byte
	var allowlist []string
	var updated time.Time
	err := s.db.QueryRow(ctx, `
		SELECT security_mode, flags, config, ip_allowlist, updated_at
		FROM settings WHERE id=1
	`).Scan(&mode, &flagsData, &configData, &allowlist, &updated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settings{}, ErrNotFound
		}
		return Settings{}, err
	}
	flags := map[string]bool{}
	if len(flagsData) > 0 {
		if err := json.Unmarshal(flagsData, &flags); err != nil {
			return Settings{}, err
		}
	}
	config := map[string]any{}
	if len(configData) > 0 {
		if err := json.Unmarshal(configData, &config); err != nil {
			return Settings{}, err
		}
	}
	return Settings{
		SecurityMode: mode,
		Flags:        flags,
		Config:       config,
		IPAllowlist:  allowlist,
		UpdatedAt:    updated,
	}, nil
}

func (s *Store) insert(ctx context.Context, settings Settings) error {
	flags, err := json.Marshal(settings.Flags)
	if err != nil {
		return err
	}
	config, err := json.Marshal(settings.Config)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO settings (id, security_mode, flags, config, ip_allowlist, updated_at)
		VALUES (1, $1, $2, $3, $4, now())
	`, settings.SecurityMode, flags, config, settings.IPAllowlist)
	return err
}
