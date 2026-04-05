package store

import (
	"context"

	"github.com/nunoOliveiraqwe/torii/internal/domain"
)

type ApiKeyStore interface {
	NewApiKey(ctx context.Context, key *domain.ApiKey) error
	DeleteApiKey(ctx context.Context, alias string) error
	IsKeyValidForScope(ctx context.Context, key string, scope string) (bool, error)
	GetApiKey(ctx context.Context, alias string) (*domain.ApiKey, error)
	GetApiKeyByRawKey(ctx context.Context, key string) (*domain.ApiKey, error)
	GetAllApiKeys(ctx context.Context) ([]*domain.ApiKey, error)
}
