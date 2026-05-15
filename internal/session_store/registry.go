package session_store

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/nunoOliveiraqwe/torii/internal/sqlite"
	"go.uber.org/zap"
)

type Config struct {
	Lifetime        time.Duration
	IdleTimeout     time.Duration
	CleanupInterval time.Duration
	CookieDomain    string
	CookieSecure    bool
	CookieHttpOnly  bool
	CookieSameSite  string
}

func (c Config) SameSiteMode() http.SameSite {
	switch strings.ToLower(c.CookieSameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

type Registry[T any] struct {
	manager *scs.SessionManager
}

func NewRegistry[T any](cookieName, cookiePath string, db *sqlite.DB, cfg Config) *Registry[T] {
	return NewRegistryWithStore[T](cookieName, cookiePath,
		sqlite3store.NewWithCleanupInterval(db.GetDb(), cfg.CleanupInterval), cfg)
}

func NewRegistryWithStore[T any](cookieName, cookiePath string, store scs.Store, cfg Config) *Registry[T] {
	sessionManager := scs.New()
	sessionManager.Store = store
	sessionManager.Lifetime = cfg.Lifetime
	sessionManager.IdleTimeout = cfg.IdleTimeout
	sessionManager.Cookie.Name = cookieName
	sessionManager.Cookie.Domain = cfg.CookieDomain
	sessionManager.Cookie.HttpOnly = cfg.CookieHttpOnly
	sessionManager.Cookie.Path = cookiePath
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = cfg.SameSiteMode()
	sessionManager.Cookie.Secure = cfg.CookieSecure

	return &Registry[T]{manager: sessionManager}
}

func (reg *Registry[T]) RenewSession(r *http.Request) error {
	err := reg.manager.RenewToken(r.Context())
	if err != nil {
		zap.L().Error("failed to renew session", zap.Error(err))
		return err
	}
	return nil
}

func (reg *Registry[T]) GetValueFromSession(r *http.Request, key string) *T {
	zap.S().Debugf("Fetching value with key %s from session", key)
	val := reg.manager.Get(r.Context(), key)
	if val == nil {
		zap.S().Debugf("No value found for key %s in session", key)
		return nil
	}
	castedVal, ok := val.(T)
	if !ok {
		var zero T
		zap.S().Errorf("failed to cast session value for key %s from %T to %T", key, val, zero)
		return nil
	}
	return &castedVal
}

func (reg *Registry[T]) NewSession(r *http.Request, key string, value T) error {
	zap.S().Infof("Creating new session for key %s", key)
	err := reg.RenewSession(r)
	if err != nil {
		return err
	}
	reg.manager.Put(r.Context(), key, value)
	return nil
}

func (reg *Registry[T]) HasValidSession(key string, r *http.Request) bool {
	val := reg.manager.Get(r.Context(), key)
	if val == nil {
		return false
	}
	_, ok := val.(T)
	return ok
}

func (reg *Registry[T]) LogoutSession(key string, r *http.Request) {
	reg.manager.Remove(r.Context(), key)
	err := reg.manager.Destroy(r.Context())
	if err != nil {
		zap.S().Errorf("Cannot logout session, %v", err)
	}
}

func (reg *Registry[T]) WrapWithSessionMiddleware(next http.Handler) http.Handler {
	return reg.manager.LoadAndSave(next)
}

func (reg *Registry[T]) KillSessionsForKeyValue(key string, value T) {
	contexts := []context.Context{}
	err := reg.manager.Iterate(context.Background(), func(ctx context.Context) error {
		val := reg.manager.Get(ctx, key)
		castedVal, ok := val.(T)
		if !ok {
			var zero T
			zap.S().Errorf("failed to cast session value for key %s from %T to %T", key, val, zero)
			return nil
		}
		if reflect.DeepEqual(castedVal, value) {
			zap.S().Debugf("Appending context to terminate list")
			contexts = append(contexts, ctx)
		}
		return nil
	})
	if err != nil {
		zap.S().Errorf("Cannot kill sessions for k:%s v:%v, %v", key, value, err)
		return
	}
	for _, ctx := range contexts {
		err := reg.manager.Destroy(ctx)
		if err != nil {
			zap.S().Errorf("Cannot logout session, %v", err)
		}
	}
}
