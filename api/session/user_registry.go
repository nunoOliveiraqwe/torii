package session

import (
	"net/http"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/nunoOliveiraqwe/torii/config"
	"github.com/nunoOliveiraqwe/torii/internal/session_store"
	"github.com/nunoOliveiraqwe/torii/internal/sqlite"
)

const cookieName = "torii-id"
const cookiePath = "/"
const keyUserId = "username"

type UserRegistry struct {
	*session_store.Registry[string]
}

func NewRegistry(db *sqlite.DB, cfg config.SessionConfig) *UserRegistry {
	reg := session_store.NewRegistryWithStore[string](
		cookieName, cookiePath, sqlite3store.NewWithCleanupInterval(db.GetDb(), cfg.CleanupInterval), toStoreConfig(cfg))
	return &UserRegistry{Registry: reg}
}

func toStoreConfig(cfg config.SessionConfig) session_store.Config {
	return session_store.Config{
		Lifetime:        cfg.Lifetime,
		IdleTimeout:     cfg.IdleTimeout,
		CleanupInterval: cfg.CleanupInterval,
		CookieDomain:    cfg.CookieDomain,
		CookieSecure:    cfg.CookieSecure,
		CookieHttpOnly:  cfg.CookieHttpOnly,
		CookieSameSite:  cfg.CookieSameSite,
	}
}

func (reg *UserRegistry) RenewSession(r *http.Request) error {
	return reg.Registry.RenewSession(r)
}

func (reg *UserRegistry) GetValueFromSession(r *http.Request) string {
	val := reg.Registry.GetValueFromSession(r, keyUserId)
	if val == nil {
		return ""
	}
	return *val
}

func (reg *UserRegistry) NewSession(r *http.Request, username string) error {
	return reg.Registry.NewSession(r, keyUserId, username)
}

func (reg *UserRegistry) HasValidSession(r *http.Request) bool {
	return reg.Registry.HasValidSession(keyUserId, r)
}

func (reg *UserRegistry) LogoutSession(r *http.Request) {
	reg.Registry.LogoutSession(keyUserId, r)
}

func (reg *UserRegistry) WrapWithSessionMiddleware(next http.Handler) http.Handler {
	return reg.Registry.WrapWithSessionMiddleware(next)
}

func (reg *UserRegistry) KillSessionsForUsername(username string) {
	reg.Registry.KillSessionsForKeyValue(keyUserId, username)
}
