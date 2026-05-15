package totp

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2/memstore"
	"github.com/nunoOliveiraqwe/torii/internal/session_store"
)

const cookieName = "torii-otp-id"
const cookiePath = "/"
const keyRegID = "torii-totp"

type Registry struct {
	*session_store.Registry[SessionData]
}

type SessionData struct {
	VerifiedAt time.Time
	LastSeenAt time.Time
	IP         string
	UserAgent  string
	SeedLabel  string
	Scope      string
}

func NewRegistry(cfg session_store.Config) *Registry {
	inMemorySessionStore := memstore.New() //for now only in memory
	gob.Register(SessionData{})
	reg := session_store.NewRegistryWithStore[SessionData](
		cookieName, cookiePath, inMemorySessionStore, cfg)
	return &Registry{Registry: reg}
}

func (reg *Registry) RenewSession(r *http.Request) error {
	return reg.Registry.RenewSession(r)
}

func (reg *Registry) GetValueFromSession(r *http.Request) *SessionData {
	return reg.Registry.GetValueFromSession(r, keyRegID)
}

func (reg *Registry) NewSession(r *http.Request, seedLabel, scope string) error {
	return reg.Registry.NewSession(r, keyRegID, createSessionDataFromRequest(r, seedLabel, scope))
}

func (reg *Registry) HasValidSession(r *http.Request) bool {
	sessionData := reg.Registry.GetValueFromSession(r, keyRegID)
	if sessionData == nil {
		return false
	}
	if sessionData.IP != r.RemoteAddr || sessionData.UserAgent != r.UserAgent() {
		return false
	}
	return true
}

func (reg *Registry) LogoutSession(r *http.Request) {
	reg.Registry.LogoutSession(keyRegID, r)
}

func (reg *Registry) WrapWithSessionMiddleware(next http.Handler) http.Handler {
	return reg.Registry.WrapWithSessionMiddleware(next)
}

func createSessionDataFromRequest(r *http.Request, seedLabel, scope string) SessionData {
	now := time.Now().UTC()
	return SessionData{
		VerifiedAt: now,
		LastSeenAt: now,
		IP:         r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		SeedLabel:  seedLabel,
		Scope:      scope,
	}
}
