package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BindAddr             string
	DatabaseURL          string
	MongoURI             string
	MasterKey            []byte
	SessionTTL           time.Duration
	SessionCookieName    string
	CSRFHeaderName       string
	SettingsRefresh      time.Duration
	LoginRateLimitPerMin int
	LoginBurst           int
	RequestTimeout       time.Duration
	CORSAllowOrigins     []string
	TrustedProxyCIDR     []string
	OIDCIssuerURL        string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCRedirectURL      string
	OIDCScopes           []string
	OIDCGroupClaim       string
	OIDCAdminGroup       string
	OIDCRoleMap          map[string]string
	AdminUser            string
	AdminPass            string
	AutoMigrate          bool
	StepUpMaxAge         time.Duration
	GlobalMaxRows        int
	StatementTimeout     time.Duration
	AllowInsecureCookies bool
	TrustedMTLSHeader    string
	UpdateRepo           string
	UpdateCheckInterval  time.Duration
	UpdateToken          string
	UpdateAutoRestart    bool
}

func Load() (*Config, error) {
	cfg := &Config{
		BindAddr:             envOrDefault("BIND_ADDR", "127.0.0.1:8080"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		MongoURI:             envOrDefault("MONGO_URI", "mongodb://mongo:27017"),
		SessionTTL:           envDuration("SESSION_TTL", 24*time.Hour),
		SessionCookieName:    envOrDefault("SESSION_COOKIE_NAME", "flowdb_session"),
		CSRFHeaderName:       envOrDefault("CSRF_HEADER_NAME", "X-CSRF-Token"),
		SettingsRefresh:      envDuration("SETTINGS_REFRESH", 10*time.Second),
		LoginRateLimitPerMin: envInt("LOGIN_RATE_LIMIT_PER_MIN", 10),
		LoginBurst:           envInt("LOGIN_RATE_LIMIT_BURST", 5),
		RequestTimeout:       envDuration("REQUEST_TIMEOUT", 30*time.Second),
		OIDCIssuerURL:        os.Getenv("OIDC_ISSUER_URL"),
		OIDCClientID:         os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:     os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:      os.Getenv("OIDC_REDIRECT_URL"),
		OIDCScopes:           splitCSV(envOrDefault("OIDC_SCOPES", "openid,profile,email,groups")),
		OIDCGroupClaim:       envOrDefault("OIDC_GROUP_CLAIM", "groups"),
		OIDCAdminGroup:       envOrDefault("OIDC_ADMIN_GROUP", "flowdb-admin"),
		AdminUser:            envOrDefault("ADMIN_USER", "admin"),
		AdminPass:            envOrDefault("ADMIN_PASS", "admin"),
		AutoMigrate:          envBool("AUTO_MIGRATE", true),
		StepUpMaxAge:         envDuration("STEP_UP_MAX_AGE", 10*time.Minute),
		GlobalMaxRows:        envInt("GLOBAL_MAX_ROWS", 1000),
		StatementTimeout:     envDuration("STATEMENT_TIMEOUT", 30*time.Second),
		AllowInsecureCookies: envBool("ALLOW_INSECURE_COOKIES", false),
		TrustedMTLSHeader:    envOrDefault("MTLS_TRUSTED_HEADER", "X-Client-Cert-Verified"),
		UpdateRepo:           envOrDefault("UPDATE_REPO", "vietrix/flowdb"),
		UpdateCheckInterval:  envDuration("UPDATE_CHECK_INTERVAL", 5*time.Minute),
		UpdateToken:          os.Getenv("UPDATE_GITHUB_TOKEN"),
		UpdateAutoRestart:    envBool("UPDATE_AUTO_RESTART", true),
	}

	if cors := os.Getenv("CORS_ALLOW_ORIGINS"); cors != "" {
		cfg.CORSAllowOrigins = splitCSV(cors)
	}
	if trust := os.Getenv("TRUSTED_PROXY_CIDR"); trust != "" {
		cfg.TrustedProxyCIDR = splitCSV(trust)
	}
	if roleMap := os.Getenv("OIDC_ROLE_MAP"); roleMap != "" {
		if err := json.Unmarshal([]byte(roleMap), &cfg.OIDCRoleMap); err != nil {
			return nil, err
		}
	}
	masterKey := os.Getenv("MASTER_KEY")
	if masterKey == "" {
		return nil, errors.New("MASTER_KEY required")
	}
	keyBytes, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, err
	}
	if len(keyBytes) != 32 {
		return nil, errors.New("MASTER_KEY must be 32 bytes base64")
	}
	cfg.MasterKey = keyBytes
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL required")
	}
	return cfg, nil
}

func envOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func envInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return parsed
}

func envBool(key string, def bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}
	return parsed
}

func envDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return def
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
