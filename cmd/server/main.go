// cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flowdb/backend/audit"
	"flowdb/backend/auth"
	"flowdb/backend/config"
	"flowdb/backend/connections"
	"flowdb/backend/crypto"
	"flowdb/backend/http/handlers"
	"flowdb/backend/http/routes"
	"flowdb/backend/iam"
	"flowdb/backend/policies"
	"flowdb/backend/query"
	"flowdb/backend/settings"
	"flowdb/backend/store"
	"flowdb/backend/stream"
	"flowdb/backend/update"
	"flowdb/backend/util"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"golang.org/x/oauth2"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config error", "error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect error", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			logger.Error("migration db error", "error", err)
			os.Exit(1)
		}
		defer sqlDB.Close()
		if err := goose.SetDialect("postgres"); err != nil {
			logger.Error("goose dialect error", "error", err)
			os.Exit(1)
		}
		if err := goose.Up(sqlDB, "migrations"); err != nil {
			logger.Error("migration error", "error", err)
			os.Exit(1)
		}
	}

	cipher, err := crypto.NewAESCipher(cfg.MasterKey)
	if err != nil {
		logger.Error("cipher error", "error", err)
		os.Exit(1)
	}

	st := store.New(pool)
	settingsStore := settings.NewStore(pool, cfg.SettingsRefresh)
	if err := settingsStore.Start(ctx); err != nil {
		logger.Error("settings error", "error", err)
		os.Exit(1)
	}
	policyStore := policies.NewStore(st, cfg.SettingsRefresh)
	if err := policyStore.Start(ctx); err != nil {
		logger.Error("policy error", "error", err)
		os.Exit(1)
	}
	if err := ensureRoles(ctx, st); err != nil {
		logger.Error("role init error", "error", err)
		os.Exit(1)
	}
	if err := ensureAdmin(ctx, st, cfg, logger); err != nil {
		logger.Error("admin init error", "error", err)
		os.Exit(1)
	}

	sessions := auth.NewSessionManager(st, cfg.SessionCookieName, cfg.SessionTTL)
	connService := connections.NewService(st, cipher)
	authorizer := iam.NewAuthorizer(st, policyStore)
	auditLogger := audit.NewLogger(st, settingsStore)
	streamManager := stream.NewManager()
	jobStore := query.NewJobStore(10 * time.Minute)
	updateService := update.NewService(cfg.UpdateRepo, util.Version, cfg.UpdateCheckInterval, cfg.UpdateToken)

	var oidcProvider *oidc.Provider
	var oidcConfig *oauth2.Config
	var oidcVerifier *oidc.IDTokenVerifier
	if cfg.OIDCIssuerURL != "" && cfg.OIDCClientID != "" && cfg.OIDCClientSecret != "" {
		oidcProvider, err = oidc.NewProvider(ctx, cfg.OIDCIssuerURL)
		if err != nil {
			logger.Error("oidc provider error", "error", err)
		} else {
			oidcConfig = &oauth2.Config{
				ClientID:     cfg.OIDCClientID,
				ClientSecret: cfg.OIDCClientSecret,
				Endpoint:     oidcProvider.Endpoint(),
				RedirectURL:  cfg.OIDCRedirectURL,
				Scopes:       cfg.OIDCScopes,
			}
			oidcVerifier = oidcProvider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})
		}
	}

	h := &handlers.Handler{
		Store:        st,
		Settings:     settingsStore,
		Sessions:     sessions,
		Cipher:       cipher,
		Connections:  connService,
		Policies:     policyStore,
		Authorizer:   authorizer,
		Audit:        auditLogger,
		Config:       cfg,
		Logger:       logger,
		Stream:       streamManager,
		JobStore:     jobStore,
		Update:       updateService,
		OIDC:         oidcProvider,
		OIDCConfig:   oidcConfig,
		OIDCVerifier: oidcVerifier,
	}

	server := &http.Server{
		Addr:         cfg.BindAddr,
		Handler:      routes.NewRouter(h, cfg, settingsStore),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", "addr", cfg.BindAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
	case <-ctx.Done():
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
	streamManager.CloseAll()
}

func ensureRoles(ctx context.Context, st *store.Store) error {
	_, err := st.CreateRoleIfMissing(ctx, "admin", []string{"*"})
	if err != nil {
		return err
	}
	_, err = st.CreateRoleIfMissing(ctx, "editor", []string{
		"connection:read", "connection:write", "query:read", "query:write", "history:read",
	})
	if err != nil {
		return err
	}
	_, err = st.CreateRoleIfMissing(ctx, "viewer", []string{
		"connection:read", "query:read", "history:read",
	})
	return err
}

func ensureAdmin(ctx context.Context, st *store.Store, cfg *config.Config, logger *slog.Logger) error {
	count, err := st.CountUsers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := auth.HashPassword(cfg.AdminPass)
	if err != nil {
		return err
	}
	user, err := st.CreateUser(ctx, store.User{
		Username:     cfg.AdminUser,
		PasswordHash: hash,
		IsAdmin:      true,
	})
	if err != nil {
		return err
	}
	role, err := st.CreateRoleIfMissing(ctx, "admin", []string{"*"})
	if err == nil {
		_ = st.BindRoleToUser(ctx, role.ID, user.ID, "")
	}
	if cfg.AdminUser == "admin" && cfg.AdminPass == "admin" {
		logger.Warn("admin credentials defaulted")
	}
	return nil
}
