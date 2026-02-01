package routes

import (
	"net/http"

	"flowdb/backend/config"
	"flowdb/backend/http/handlers"
	"flowdb/backend/middleware"
	"flowdb/backend/settings"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *handlers.Handler, cfg *config.Config, settingsStore *settings.Store) http.Handler {
	r := chi.NewRouter()
	limiter := middleware.NewRateLimiter(cfg.LoginRateLimitPerMin, cfg.LoginBurst)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(h.Logger))
	r.Use(middleware.Logging(h.Logger))
	if len(cfg.CORSAllowOrigins) > 0 {
		r.Use(middleware.CORS(cfg.CORSAllowOrigins))
	}
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.SecurityControls(settingsStore, cfg.TrustedMTLSHeader))
		r.Route("/auth", func(r chi.Router) {
			r.With(limiter.Middleware).Post("/login", h.Login)
			r.With(limiter.Middleware).Get("/oidc/callback", h.OIDCCallback)
			r.Get("/oidc/login", h.OIDCLogin)
			r.Get("/saml/login", h.SAMLLogin)
			r.With(middleware.RequireAuth(h.Store, h.Sessions)).Post("/logout", h.Logout)
			r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/me", h.Me)
			r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/mfa/enroll", h.EnrollMFA)
			r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/mfa/verify", h.VerifyMFA)
		})
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/settings", h.GetSettings)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Put("/settings", h.UpdateSettings)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Put("/settings/security-mode", h.UpdateSecurityMode)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Put("/settings/flags", h.UpdateFlags)

		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections", h.ListConnections)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/connections", h.CreateConnection)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections/{id}", h.GetConnection)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Put("/connections/{id}", h.UpdateConnection)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Delete("/connections/{id}", h.DeleteConnection)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Post("/connections/{id}/test", h.TestConnection)

		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections/{id}/namespaces", h.ListNamespaces)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections/{id}/entities", h.ListEntities)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections/{id}/entities/{name}/info", h.GetEntityInfo)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections/{id}/entities/{name}/browse", h.BrowseEntity)

		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/connections/{id}/query", h.StartQuery)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/connections/{id}/query/{queryId}/stream", h.StreamQuery)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/connections/{id}/explain", h.ExplainQuery)

		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/approvals/pending", h.ListPendingApprovals)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/approvals/{id}/approve", h.Approve)
		r.With(middleware.RequireAuth(h.Store, h.Sessions), middleware.CSRF(cfg.CSRFHeaderName)).Post("/approvals/{id}/deny", h.Deny)

		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/history", h.ListHistory)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/audit", h.ListAudit)

		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/scim/Users", h.ListSCIMUsers)
		r.With(middleware.RequireAuth(h.Store, h.Sessions)).Get("/scim/Groups", h.ListSCIMGroups)
	})
	return r
}
