package handlers

import (
	"context"
	"net/http"
	"strings"

	"flowdb/backend/auth"
	"flowdb/backend/store"

	"github.com/google/uuid"
)

func (h *Handler) SAMLLogin(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_sso_saml") {
		http.Error(w, "sso disabled", http.StatusForbidden)
		return
	}
	username := r.Header.Get("X-SAML-User")
	email := r.Header.Get("X-SAML-Email")
	groupHeader := r.Header.Get("X-SAML-Groups")
	if username == "" && email == "" {
		http.Error(w, "missing saml headers", http.StatusBadRequest)
		return
	}
	groups := []string{}
	if groupHeader != "" {
		for _, g := range strings.Split(groupHeader, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				groups = append(groups, g)
			}
		}
	}
	subject := username
	if subject == "" {
		subject = email
	}
	user, err := h.findOrCreateExternalUser(r.Context(), "saml", subject, email, groups)
	if err != nil {
		http.Error(w, "failed to map user", http.StatusInternalServerError)
		return
	}
	session, err := h.Sessions.Create(r.Context(), user.ID, nil)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	h.Sessions.SetCookie(w, session.ID, !h.Config.AllowInsecureCookies)
	_ = h.Audit.LogEvent(r.Context(), "saml_login", &user.ID, map[string]any{"subject": subject, "email": email}, "")
	writeJSON(w, http.StatusOK, loginResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CSRFToken: session.CSRFToken,
	})
}

func (h *Handler) findOrCreateExternalUser(ctx context.Context, provider string, subject string, email string, groups []string) (store.User, error) {
	identity, err := h.Store.GetExternalIdentity(ctx, provider, subject)
	if err == nil {
		return h.Store.GetUserByID(ctx, identity.UserID)
	}
	username := email
	if username == "" {
		username = subject
	}
	username = sanitizeUsername(username)
	_, err = h.Store.GetUserByUsername(ctx, username)
	if err == nil {
		username = username + "-" + uuid.NewString()[:8]
	}
	hash, err := auth.HashPassword(uuid.NewString())
	if err != nil {
		return store.User{}, err
	}
	isAdmin := false
	count, err := h.Store.CountUsers(ctx)
	if err == nil && count == 0 {
		isAdmin = true
	}
	user, err := h.Store.CreateUser(ctx, store.User{
		Username:     username,
		PasswordHash: hash,
		IsAdmin:      isAdmin,
	})
	if err != nil {
		return store.User{}, err
	}
	_, _ = h.Store.UpsertExternalIdentity(ctx, store.ExternalIdentity{
		UserID:   user.ID,
		Provider: provider,
		Subject:  subject,
		Email:    email,
		Groups:   groups,
	})
	groupIDs := []uuid.UUID{}
	for _, name := range groups {
		group, err := h.Store.UpsertGroup(ctx, name)
		if err == nil {
			groupIDs = append(groupIDs, group.ID)
		}
		if strings.EqualFold(name, h.Config.OIDCAdminGroup) {
			_ = h.Store.UpdateUserAdmin(ctx, user.ID, true)
			user.IsAdmin = true
		}
		if roleName, ok := h.Config.OIDCRoleMap[name]; ok {
			role, err := h.Store.CreateRoleIfMissing(ctx, roleName, nil)
			if err == nil {
				_ = h.Store.BindRoleToUser(ctx, role.ID, user.ID, "")
			}
		}
	}
	if len(groupIDs) > 0 {
		_ = h.Store.SetUserGroups(ctx, user.ID, groupIDs)
	}
	return user, nil
}
