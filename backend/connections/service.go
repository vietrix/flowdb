package connections

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"flowdb/backend/adapters"
	"flowdb/backend/adapters/mongodb"
	"flowdb/backend/adapters/postgres"
	"flowdb/backend/crypto"
	"flowdb/backend/store"
)

type Secret struct {
	Password string `json:"password"`
}

type Service struct {
	store  *store.Store
	cipher *crypto.AESCipher
}

func NewService(st *store.Store, cipher *crypto.AESCipher) *Service {
	return &Service{store: st, cipher: cipher}
}

func (s *Service) GetAdapter(ctx context.Context, conn store.Connection) (adapters.Adapter, error) {
	secretData, err := s.store.GetConnectionSecret(ctx, conn.SecretRef)
	if err != nil {
		return nil, err
	}
	plain, err := s.cipher.Decrypt(secretData)
	if err != nil {
		return nil, err
	}
	var secret Secret
	if err := json.Unmarshal(plain, &secret); err != nil {
		return nil, err
	}
	switch conn.Type {
	case "postgres":
		host := normalizeHost(conn.Host)
		return postgres.New(ctx, postgres.Config{
			Host:     host,
			Port:     conn.Port,
			Database: conn.Database,
			User:     conn.Username,
			Password: secret.Password,
			SSLMode:  "prefer",
		})
	case "mongodb":
		uri := conn.Host
		if uri == "" {
			uri = "mongodb://" + normalizeHost("localhost") + ":27017"
		}
		if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
			host := normalizeHost(conn.Host)
			if conn.Port > 0 {
				host = fmt.Sprintf("%s:%d", host, conn.Port)
			}
			uri = "mongodb://" + host
		}
		return mongodb.New(ctx, mongodb.Config{
			URI:      uri,
			Database: conn.Database,
		})
	default:
		return nil, errors.New("unsupported connection type")
	}
}

func normalizeHost(host string) string {
	lower := strings.ToLower(strings.TrimSpace(host))
	if lower == "" {
		return host
	}
	if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" {
		if gw := os.Getenv("DOCKER_HOST_GATEWAY"); gw != "" {
			return gw
		}
		if os.Getenv("DOCKER_MODE") == "true" {
			return "host.docker.internal"
		}
	}
	return host
}
