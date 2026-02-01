package connections

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		return postgres.New(ctx, postgres.Config{
			Host:     conn.Host,
			Port:     conn.Port,
			Database: conn.Database,
			User:     conn.Username,
			Password: secret.Password,
			SSLMode:  "prefer",
		})
	case "mongodb":
		uri := conn.Host
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}
		if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
			host := conn.Host
			if conn.Port > 0 {
				host = fmt.Sprintf("%s:%d", conn.Host, conn.Port)
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
