package app

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/domain"
	"github.com/nunoOliveiraqwe/torii/internal/store"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"go.uber.org/zap"
)

var ErrorInvalidApiKeyRequest = fmt.Errorf("invalid API key request: no data")
var ErrorInvalidScopesApiKeyRequest = fmt.Errorf("invalid API key request: invalid scopes")
var ErrorInvalidAliasApiKeyRequest = fmt.Errorf("invalid API key request: invalid alias")
var ErrorInvalidExpiryDateApiKeyRequest = fmt.Errorf("invalid API key request: invalid expiry date")
var ErrorDuplicatedAliasApiKeyRequest = fmt.Errorf("invalid API key request: duplicated alias")

type CreateApiKeyRequest struct {
	Alias      string   `json:"alias"`
	Scopes     []string `json:"scopes"`
	ExpiryDate time.Time
}

type apiKeyCacheEntry struct {
	allowedScopes map[string]byte
	expiresAt     time.Time
	lastSeen      time.Time
}

func (e *apiKeyCacheEntry) Touch() {
	e.lastSeen = time.Now()
}

func (e *apiKeyCacheEntry) GetLastReadAt() time.Time {
	return e.lastSeen
}

type ApiKeyService struct {
	store store.ApiKeyStore
	cache *util.Cache[*apiKeyCacheEntry]
}

func NewApiKeyService(store store.ApiKeyStore) *ApiKeyService {
	zap.S().Info("Initializing API Key Service with hardcoded cache defaults (this might change in the future)")

	cacheOpts := &util.CacheOptions{
		MaxEntries:      10000,
		TTL:             1 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}

	cache, err := util.NewCache[*apiKeyCacheEntry](cacheOpts)
	if err != nil {
		zap.S().Fatalf("Failed to initialize API key cache: %v", err)
	}

	return &ApiKeyService{
		store: store,
		cache: cache,
	}
}

func (s *ApiKeyService) IsKeyValidForScope(key string, scope string) (bool, error) {
	// Check cache first
	entry, err := s.cache.GetValue(key)
	if err == nil {
		if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
			s.cache.Evict(key)
			return false, nil
		}
		_, ok := entry.allowedScopes[scope]
		return ok, nil
	}

	// Cache miss — query the store
	valid, err := s.store.IsKeyValidForScope(context.Background(), key, scope)
	if err != nil {
		return false, fmt.Errorf("failed to check key scope: %w", err)
	}

	// On a valid key, fetch the full key to populate the cache with all scopes
	if valid {
		s.warmCache(key)
	}

	return valid, nil
}

func (s *ApiKeyService) CreateApiKey(ctx context.Context, apiKeyRequest *CreateApiKeyRequest) (*domain.ApiKey, error) {
	if apiKeyRequest == nil {
		return nil, ErrorInvalidApiKeyRequest
	} else if strings.EqualFold(apiKeyRequest.Alias, "") {
		return nil, ErrorInvalidAliasApiKeyRequest
	} else if apiKeyRequest.Scopes == nil || len(apiKeyRequest.Scopes) == 0 {
		return nil, ErrorInvalidScopesApiKeyRequest
	} else if !apiKeyRequest.ExpiryDate.IsZero() && apiKeyRequest.ExpiryDate.Before(time.Now()) {
		return nil, ErrorInvalidExpiryDateApiKeyRequest
	}

	zap.S().Debugf("Creating API key with alias %s and scopes %v", apiKeyRequest.Alias, apiKeyRequest.Scopes)

	//API keys have a torii wide scope, so we need to check if any key already exists with such as alias
	k, err := s.GetApiKey(ctx, apiKeyRequest.Alias)
	if err == nil && k != nil {
		zap.S().Warnf("API key with alias %s already exists, cannot create another one with the same alias", apiKeyRequest.Alias)
		return nil, ErrorDuplicatedAliasApiKeyRequest
	}
	scopeMap := make(map[domain.Scope]byte, len(apiKeyRequest.Scopes))
	for _, scope := range apiKeyRequest.Scopes {
		s, ok := domain.AvailableScopesMap[domain.Scope(scope)]
		if !ok {
			zap.S().Warnf("Invalid scope %s provided for API key creation", scope)
			return nil, ErrorInvalidScopesApiKeyRequest
		}
		scopeMap[domain.Scope(scope)] = s
	}
	apiKey := &domain.ApiKey{
		Alias:     apiKeyRequest.Alias,
		Key:       generateApiKey(),
		Scopes:    scopeMap,
		Expires:   apiKeyRequest.ExpiryDate,
		CreatedAt: time.Now().Unix(),
	}
	err = s.store.NewApiKey(ctx, apiKey)
	if err != nil {
		zap.S().Errorf("Failed to create API key with alias %s: %v", apiKeyRequest.Alias, err)
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}
	zap.S().Debugf("API key with alias %s created successfully", apiKeyRequest.Alias)
	return apiKey, nil
}

func (s *ApiKeyService) DeleteApiKey(ctx context.Context, alias string) error {
	return s.store.DeleteApiKey(ctx, alias)
}

func (s *ApiKeyService) GetApiKey(ctx context.Context, alias string) (*domain.ApiKey, error) {
	zap.S().Debugf("Fetching API key by alias %s", alias)
	apiKey, err := s.store.GetApiKey(ctx, alias)
	if err != nil {
		zap.S().Warnf("Failed to fetch API key by alias %s: %v", alias, err)
		return nil, err
	}
	if apiKey == nil {
		return nil, nil
	}
	apiKey.Key = "" //we never return the actual API key outside the service (except creation obviously), so we clear it here to avoid any accidental leaks
	return apiKey, nil
}

func (s *ApiKeyService) GetAllApiKeys(ctx context.Context) []*domain.ApiKey {
	zap.S().Debugf("Fetching all API keys")
	apiKeys, err := s.store.GetAllApiKeys(ctx)
	if err != nil {
		zap.S().Warnf("Failed to fetch all API keys: %v", err)
		return nil
	}
	for _, apiKey := range apiKeys {
		apiKey.Key = "" //we never return the actual API key outside the service (except creation obviously), so we clear it here to avoid any accidental leaks
	}
	return apiKeys
}

// warmCache fetches the full key from the store by its raw value
// and caches all its scopes so subsequent scope checks are fast.
func (s *ApiKeyService) warmCache(key string) {
	apiKey, err := s.store.GetApiKeyByRawKey(context.Background(), key)
	if err != nil || apiKey == nil {
		zap.S().Debugf("Cache warm skipped for API key: %v", err)
		return
	}

	scopes := make(map[string]byte, len(apiKey.Scopes))
	for scope := range apiKey.Scopes {
		scopes[string(scope)] = 1
	}

	s.cache.CacheValue(key, &apiKeyCacheEntry{
		allowedScopes: scopes,
		expiresAt:     apiKey.Expires,
		lastSeen:      time.Now(),
	})
}

func generateApiKey() string {
	zap.S().Debug("Generating API key")
	return rand.Text()
}
