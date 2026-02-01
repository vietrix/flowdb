package handlers

import (
	"log/slog"

	"flowdb/backend/audit"
	"flowdb/backend/auth"
	"flowdb/backend/config"
	"flowdb/backend/connections"
	"flowdb/backend/crypto"
	"flowdb/backend/iam"
	"flowdb/backend/policies"
	"flowdb/backend/query"
	"flowdb/backend/settings"
	"flowdb/backend/store"
	"flowdb/backend/stream"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Handler struct {
	Store        *store.Store
	Settings     *settings.Store
	Sessions     *auth.SessionManager
	Cipher       *crypto.AESCipher
	Connections  *connections.Service
	Policies     *policies.Store
	Authorizer   *iam.Authorizer
	Audit        *audit.Logger
	Config       *config.Config
	Logger       *slog.Logger
	Stream       *stream.Manager
	JobStore     *query.JobStore
	OIDC         *oidc.Provider
	OIDCConfig   *oauth2.Config
	OIDCVerifier *oidc.IDTokenVerifier
}
