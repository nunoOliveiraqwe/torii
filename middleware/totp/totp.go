package totp

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/ratelimit"
	"github.com/nunoOliveiraqwe/torii/internal/resolve"
	"github.com/nunoOliveiraqwe/torii/internal/session_store"
	"go.uber.org/zap"
)

const (
	defaultCodeWindow = 1
	defaultDigits     = 6
	defaultPeriod     = 30 * time.Second
)

type Manager struct {
	sessionRegistry     *Registry
	seed                []byte
	label               string
	codeWindow          int
	digits              int
	period              time.Duration
	algorithm           Algorithm
	verificationLimiter *ratelimit.Limiter
}

type Config struct {
	Label      string
	Seed       string
	CodeWindow int
	Digits     int
	Period     time.Duration
	Algorithm  Algorithm
	RateLimit  *RateLimitConfig
}

func NewTOTPManager(totpConf Config, sessionConf session_store.Config) (*Manager, error) {
	zap.S().Info("Creating TOTP manager with seed and session config")

	seed, err := resolveSeed(totpConf.Seed)
	if err != nil {
		return nil, err
	}

	codeWindow := totpConf.CodeWindow
	if codeWindow < 0 {
		return nil, fmt.Errorf("code window cannot be negative")
	}
	if codeWindow == 0 {
		codeWindow = defaultCodeWindow
	}

	digits := totpConf.Digits
	if digits == 0 {
		digits = defaultDigits
	}
	if digits < 6 || digits > 8 {
		return nil, fmt.Errorf("digits must be between 6 and 8")
	}

	period := totpConf.Period
	if period == 0 {
		period = defaultPeriod
	}
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive")
	}

	algorithm, err := normalizeAlgorithm(totpConf.Algorithm)
	if err != nil {
		return nil, err
	}

	verificationLimiter, err := newVerificationRateLimiter(totpConf.RateLimit)
	if err != nil {
		return nil, err
	}

	return &Manager{
		sessionRegistry:     NewRegistry(sessionConf),
		seed:                seed,
		label:               totpConf.Label,
		codeWindow:          codeWindow,
		digits:              digits,
		period:              period,
		algorithm:           algorithm,
		verificationLimiter: verificationLimiter,
	}, nil
}

func resolveSeed(seed string) ([]byte, error) {
	rawSecret, err := resolve.ResolveValue(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TOTP seed %w", err)
	}
	decoded, err := decodeSecret(rawSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TOTP seed %w", err)
	}
	return decoded, nil
}

func (m *Manager) HasValidSession(r *http.Request, scope string) bool {
	data := m.sessionRegistry.GetValueFromSession(r)
	if data == nil {
		return false
	}
	return data.Scope == scope
}

func (m *Manager) WrapWithSessionMiddleware(next http.Handler) http.Handler {
	return m.sessionRegistry.WrapWithSessionMiddleware(next)
}

func (m *Manager) Digits() int {
	return m.digits
}

func (m *Manager) ValidateAndStartSession(r *http.Request, code, scope string) (bool, error) {
	if m.verificationLimiter != nil {
		decision := m.verificationLimiter.Allow(r)
		if !decision.Allowed {
			return false, &RateLimitError{
				RetryAfter: decision.RetryAfter,
				Reason:     decision.Reason,
			}
		}
	}

	seedLabel, ok := m.ValidateCode(code, time.Now())
	if !ok {
		return false, nil
	}
	if m.verificationLimiter != nil {
		m.verificationLimiter.Reset(r)
	}
	if err := m.sessionRegistry.NewSession(r, seedLabel, scope); err != nil {
		return false, err
	}
	return true, nil
}

func (m *Manager) ValidateCode(code string, now time.Time) (string, bool) {
	normalizedCode := strings.TrimSpace(code)
	if len(normalizedCode) != m.digits {
		return "", false
	}
	step := now.Unix() / int64(m.period.Seconds())
	for offset := -m.codeWindow; offset <= m.codeWindow; offset++ {
		candidate, err := generateCode(m.seed, step+int64(offset), m.digits, m.algorithm)
		if err != nil {
			zap.S().Errorf("Error validating TOTP code: %v", err)
			return "", false
		}
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(normalizedCode)) == 1 {
			return m.label, true
		}
	}
	return "", false
}
