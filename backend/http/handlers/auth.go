package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"flowdb/backend/auth"

	"golang.org/x/oauth2"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	MFACode  string `json:"mfaCode"`
}

type loginResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CSRFToken string `json:"csrfToken"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, err := h.Store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	mfaEnabled := h.Settings.Get().FlagEnabled("enable_mfa") && user.MFAEnabled
	var mfaAt *time.Time
	if mfaEnabled {
		if req.MFACode == "" {
			http.Error(w, "mfa required", http.StatusUnauthorized)
			return
		}
		secret, err := h.decryptMFASecret(user.MFASecretEnc)
		if err != nil || !auth.VerifyTOTP(secret, req.MFACode) {
			http.Error(w, "invalid mfa code", http.StatusUnauthorized)
			return
		}
		now := time.Now().UTC()
		mfaAt = &now
	}
	session, err := h.Sessions.Create(r.Context(), user.ID, mfaAt)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	h.Sessions.SetCookie(w, session.ID, !h.Config.AllowInsecureCookies)
	_ = h.Audit.LogEvent(r.Context(), "login", &user.ID, map[string]any{"username": user.Username}, "")
	writeJSON(w, http.StatusOK, loginResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CSRFToken: session.CSRFToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if ok {
		_ = h.Store.DeleteSession(r.Context(), session.ID)
		_ = h.Audit.LogEvent(r.Context(), "logout", &session.UserID, map[string]any{}, "")
	}
	h.Sessions.ClearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         user.ID.String(),
		"username":   user.Username,
		"isAdmin":    user.IsAdmin,
		"mfaEnabled": user.MFAEnabled,
		"settings":   h.Settings.Get(),
	})
}

type mfaEnrollResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"otpauthUrl"`
}

func (h *Handler) EnrollMFA(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_mfa") {
		http.Error(w, "mfa disabled", http.StatusForbidden)
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	secret, err := auth.GenerateTOTP(user.Username, "FlowDB")
	if err != nil {
		http.Error(w, "failed to generate mfa", http.StatusInternalServerError)
		return
	}
	enc, err := h.encryptMFASecret(secret.Secret)
	if err != nil {
		http.Error(w, "failed to store mfa", http.StatusInternalServerError)
		return
	}
	if err := h.Store.UpdateUserMFA(r.Context(), user.ID, false, enc, nil); err != nil {
		http.Error(w, "failed to store mfa", http.StatusInternalServerError)
		return
	}
	_ = h.Audit.LogEvent(r.Context(), "mfa_enroll", &user.ID, map[string]any{}, "")
	writeJSON(w, http.StatusOK, mfaEnrollResponse{Secret: secret.Secret, URL: secret.OtpAuthURL})
}

type mfaVerifyRequest struct {
	Code string `json:"code"`
}

func (h *Handler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_mfa") {
		http.Error(w, "mfa disabled", http.StatusForbidden)
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req mfaVerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	secret, err := h.decryptMFASecret(user.MFASecretEnc)
	if err != nil || !auth.VerifyTOTP(secret, req.Code) {
		http.Error(w, "invalid code", http.StatusUnauthorized)
		return
	}
	now := time.Now().UTC()
	if err := h.Store.UpdateUserMFA(r.Context(), user.ID, true, user.MFASecretEnc, &now); err != nil {
		http.Error(w, "failed to update mfa", http.StatusInternalServerError)
		return
	}
	session, ok := auth.SessionFromContext(r.Context())
	if ok {
		_ = h.Store.UpdateSessionMFA(r.Context(), session.ID, now)
	}
	_ = h.Audit.LogEvent(r.Context(), "mfa_verify", &user.ID, map[string]any{}, "")
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_sso_oidc") {
		http.Error(w, "sso disabled", http.StatusForbidden)
		return
	}
	if h.OIDCConfig == nil || h.OIDCVerifier == nil {
		http.Error(w, "oidc not configured", http.StatusBadRequest)
		return
	}
	state, err := auth.RandomState()
	if err != nil {
		http.Error(w, "failed to create state", http.StatusInternalServerError)
		return
	}
	nonce, err := auth.RandomState()
	if err != nil {
		http.Error(w, "failed to create nonce", http.StatusInternalServerError)
		return
	}
	if err := h.saveOIDCState(r.Context(), state, nonce); err != nil {
		http.Error(w, "failed to save state", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.OIDCConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce)), http.StatusFound)
}

func (h *Handler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	if !h.Settings.Get().FlagEnabled("enable_sso_oidc") {
		http.Error(w, "sso disabled", http.StatusForbidden)
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		http.Error(w, "invalid callback", http.StatusBadRequest)
		return
	}
	nonce, err := h.consumeOIDCState(r.Context(), state)
	if err != nil {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	token, err := h.OIDCConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "exchange failed", http.StatusBadRequest)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "missing id_token", http.StatusBadRequest)
		return
	}
	idToken, err := h.OIDCVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "invalid id_token", http.StatusUnauthorized)
		return
	}
	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "invalid claims", http.StatusBadRequest)
		return
	}
	email := ""
	if val, ok := claims["email"].(string); ok {
		email = val
	}
	groups := extractGroups(claims, h.Config.OIDCGroupClaim)
	if nonce != "" {
		if idToken.Nonce != nonce {
			http.Error(w, "invalid nonce", http.StatusBadRequest)
			return
		}
	}
	subject := idToken.Subject
	user, err := h.findOrCreateExternalUser(r.Context(), "oidc", subject, email, groups)
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
	_ = h.Audit.LogEvent(r.Context(), "oidc_login", &user.ID, map[string]any{"subject": subject, "email": email}, "")
	writeJSON(w, http.StatusOK, loginResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CSRFToken: session.CSRFToken,
	})
}

func (h *Handler) encryptMFASecret(secret string) ([]byte, error) {
	return h.Cipher.Encrypt([]byte(secret))
}

func (h *Handler) decryptMFASecret(enc []byte) (string, error) {
	if len(enc) == 0 {
		return "", errors.New("missing secret")
	}
	plain, err := h.Cipher.Decrypt(enc)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (h *Handler) saveOIDCState(ctx context.Context, state string, nonce string) error {
	_, err := h.Store.DB().Exec(ctx, `
		INSERT INTO oidc_states (state, nonce, created_at) VALUES ($1,$2,now())
	`, state, nonce)
	return err
}

func (h *Handler) consumeOIDCState(ctx context.Context, state string) (string, error) {
	var nonce string
	err := h.Store.DB().QueryRow(ctx, `
		DELETE FROM oidc_states WHERE state=$1 RETURNING nonce
	`, state).Scan(&nonce)
	return nonce, err
}

func sanitizeUsername(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "@", "_")
	value = strings.ReplaceAll(value, ".", "_")
	return value
}

func (h *Handler) TokenFromHeader(r *http.Request) string {
	return r.Header.Get("Authorization")
}

func extractGroups(claims map[string]any, claimName string) []string {
	val, ok := claims[claimName]
	if !ok {
		return nil
	}
	switch t := val.(type) {
	case []any:
		out := []string{}
		for _, v := range t {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	case string:
		if t == "" {
			return nil
		}
		return strings.Split(t, ",")
	default:
		return nil
	}
}
