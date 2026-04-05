package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/domain"
	"github.com/nunoOliveiraqwe/torii/internal/store"
)

// Ensure service implements interface.
var _ store.ApiKeyStore = (*ApiKeyStore)(nil)

type ApiKeyStore struct {
	db *DB
}

func NewApiKeyStore(db *DB) store.ApiKeyStore {
	return &ApiKeyStore{db: db}
}

func (s *ApiKeyStore) NewApiKey(ctx context.Context, key *domain.ApiKey) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	scopes := marshalScopes(key.Scopes)

	var expiresAt *string
	if !key.Expires.IsZero() {
		s := key.Expires.UTC().Format(time.RFC3339)
		expiresAt = &s
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO api_key (alias, key, scopes, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		key.Alias,
		key.Key,
		scopes,
		expiresAt,
		key.CreatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ApiKeyStore) DeleteApiKey(ctx context.Context, alias string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM api_key WHERE alias = ?`, alias)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ApiKeyStore) IsKeyValidForScope(ctx context.Context, key string, scope string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var scopes string
	var expiresAt sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT scopes, expires_at FROM api_key WHERE key = ?`, key,
	).Scan(&scopes, &expiresAt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if expiresAt.Valid {
		t, parseErr := time.Parse(time.RFC3339, expiresAt.String)
		if parseErr == nil && time.Now().After(t) {
			return false, nil
		}
	}

	scopeMap := unmarshalScopes(scopes)
	_, ok := scopeMap[domain.Scope(scope)]
	return ok, nil
}

func (s *ApiKeyStore) GetApiKey(ctx context.Context, alias string) (*domain.ApiKey, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var apiKey domain.ApiKey
	var scopes string
	var expiresAt sql.NullString
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT id, alias, key, scopes, expires_at, created_at
		FROM api_key
		WHERE alias = ?`, alias,
	).Scan(
		&apiKey.ID,
		&apiKey.Alias,
		&apiKey.Key,
		&scopes,
		&expiresAt,
		(*NullTime)(&createdAt),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	apiKey.CreatedAt = createdAt.Unix()
	if expiresAt.Valid {
		apiKey.Expires, _ = time.Parse(time.RFC3339, expiresAt.String)
	}

	apiKey.Scopes = unmarshalScopes(scopes)
	return &apiKey, nil
}

func (s *ApiKeyStore) GetApiKeyByRawKey(ctx context.Context, key string) (*domain.ApiKey, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var apiKey domain.ApiKey
	var scopes string
	var expiresAt sql.NullString
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT id, alias, key, scopes, expires_at, created_at
		FROM api_key
		WHERE key = ?`, key,
	).Scan(
		&apiKey.ID,
		&apiKey.Alias,
		&apiKey.Key,
		&scopes,
		&expiresAt,
		(*NullTime)(&createdAt),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	apiKey.CreatedAt = createdAt.Unix()
	if expiresAt.Valid {
		apiKey.Expires, _ = time.Parse(time.RFC3339, expiresAt.String)
	}

	apiKey.Scopes = unmarshalScopes(scopes)
	return &apiKey, nil
}

func (s *ApiKeyStore) GetAllApiKeys(ctx context.Context) ([]*domain.ApiKey, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, alias, key, scopes, expires_at, created_at
		FROM api_key
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.ApiKey
	for rows.Next() {
		var apiKey domain.ApiKey
		var scopes string
		var expiresAt sql.NullString
		var createdAt time.Time
		if err := rows.Scan(
			&apiKey.ID,
			&apiKey.Alias,
			&apiKey.Key,
			&scopes,
			&expiresAt,
			(*NullTime)(&createdAt),
		); err != nil {
			return nil, err
		}
		apiKey.CreatedAt = createdAt.Unix()
		if expiresAt.Valid {
			apiKey.Expires, _ = time.Parse(time.RFC3339, expiresAt.String)
		}
		apiKey.Scopes = unmarshalScopes(scopes)
		keys = append(keys, &apiKey)
	}
	return keys, rows.Err()
}

// marshalScopes converts a scope map to a comma-separated string for DB storage.
func marshalScopes(scopes map[domain.Scope]byte) string {
	if len(scopes) == 0 {
		return ""
	}
	parts := make([]string, 0, len(scopes))
	for s := range scopes {
		parts = append(parts, string(s))
	}
	return strings.Join(parts, ",")
}

// unmarshalScopes converts a comma-separated string from the DB back to a scope map.
func unmarshalScopes(csv string) map[domain.Scope]byte {
	m := make(map[domain.Scope]byte)
	if csv == "" {
		return m
	}
	for _, s := range strings.Split(csv, ",") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			m[domain.Scope(trimmed)] = 1
		}
	}
	return m
}
