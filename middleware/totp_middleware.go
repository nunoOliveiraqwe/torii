package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/requestctx"
	"github.com/nunoOliveiraqwe/torii/internal/session_store"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"github.com/nunoOliveiraqwe/torii/middleware/totp"
	"go.uber.org/zap"
)

func TOTPMiddleware(ctx BuildContext, next http.HandlerFunc, conf Config) http.HandlerFunc {
	manager, err := buildTOTPManager(ctx, conf)
	if err != nil {
		zap.S().Errorf("TOTPMiddleware: failed to parse configuration: %v. Failing closed.", err)
		return func(writer http.ResponseWriter, request *http.Request) {
			http.Error(writer, "TOTPMiddleware misconfigured", http.StatusServiceUnavailable)
		}
	}

	scope := ctx.ConnectionName()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager.HasValidSession(r, scope) {
			next(w, r)
			return
		}

		if totp.IsVerifyRequest(r) {
			ok, err := manager.ValidateAndStartSession(r, r.FormValue("code"), scope)
			if err != nil {
				var rateLimitErr *totp.RateLimitError
				if errors.As(err, &rateLimitErr) {
					requestctx.CreateAndAddBlockInfoToRequestContext(r, "rate-limit", rateLimitErr.Reason, bus.TopicRateLimitTriggered)
					w.Header().Set("Retry-After", rateLimitErr.RetryAfter)
					http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
					return
				}
				zap.S().Errorf("TOTPMiddleware: failed to start session: %v", err)
				http.Error(w, "failed to start TOTP session", http.StatusInternalServerError)
				return
			}
			if ok {
				totp.RedirectAfterVerification(w, r)
				return
			}
			totp.RenderChallenge(w, r, manager.Digits(), "Invalid verification code")
			return
		}

		if r.Method == http.MethodGet && acceptsHTML(r) {
			totp.RenderChallenge(w, r, manager.Digits(), "")
			return
		}

		http.Error(w, "TOTP verification required", http.StatusUnauthorized)
	})

	return manager.WrapWithSessionMiddleware(handler).ServeHTTP
}

func buildTOTPManager(ctx BuildContext, conf Config) (*totp.Manager, error) {
	totpConf, err := parseTOTPConfig(ctx, conf)
	if err != nil {
		return nil, err
	}
	sessionConf, err := parseTOTPSessionConfig(conf)
	if err != nil {
		return nil, err
	}
	return totp.NewTOTPManager(totpConf, sessionConf)
}

func parseTOTPConfig(ctx BuildContext, conf Config) (totp.Config, error) {
	seed, err := ParseStringRequired(conf.Options, "seed")
	if err != nil {
		return totp.Config{}, err
	}
	label, err := ParseStringOpt(conf.Options, "label", "")
	if err != nil {
		return totp.Config{}, err
	}
	algorithm, err := ParseStringOpt(conf.Options, "algorithm", string(totp.AlgorithmSHA1))
	if err != nil {
		return totp.Config{}, err
	}
	period, err := parseDurationOption(conf.Options, "period", 30*time.Second)
	if err != nil {
		return totp.Config{}, err
	}
	rateLimitConf, err := parseTOTPRateLimitConfig(ctx, conf.Options)
	if err != nil {
		return totp.Config{}, err
	}

	return totp.Config{
		Label:      label,
		Seed:       seed,
		CodeWindow: ParseIntOpt(conf.Options, "code-window", 1),
		Digits:     ParseIntOpt(conf.Options, "digits", 6),
		Period:     period,
		Algorithm:  totp.Algorithm(algorithm),
		RateLimit:  rateLimitConf,
	}, nil
}

func parseTOTPRateLimitConfig(ctx BuildContext, opts map[string]interface{}) (*totp.RateLimitConfig, error) {
	defaultConf := totp.DefaultRateLimitConfig()
	enabled := ParseBoolOpt(opts, "rate-limit-enabled", true)
	if !enabled {
		return &totp.RateLimitConfig{Disabled: true}, nil
	}

	ratePerSecond := defaultConf.RatePerSecond
	burst := defaultConf.Burst
	if raw, ok := opts["limiter-req"]; ok && raw != nil {
		limiterReq, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("'limiter-req' option must be a map[string]interface{}")
		}
		var err error
		ratePerSecond, err = parseFloatOpt(limiterReq, "rate-per-second", ratePerSecond)
		if err != nil {
			return nil, err
		}
		burst, err = parseIntOptStrict(limiterReq, "burst", burst)
		if err != nil {
			return nil, err
		}
	}
	if ratePerSecond <= 0 {
		return nil, fmt.Errorf("'limiter-req.rate-per-second' must be a positive number, got %f", ratePerSecond)
	}
	if burst <= 0 {
		return nil, fmt.Errorf("'limiter-req.burst' must be a positive integer, got %d", burst)
	}

	cacheDefaults := defaultConf.CacheOptions
	cacheTTL, err := parseDurationOption(opts, "rate-limit-cache-ttl", cacheDefaults.TTL)
	if err != nil {
		return nil, err
	}
	cleanupInterval, err := parseDurationOption(opts, "rate-limit-cleanup-interval", cacheDefaults.CleanupInterval)
	if err != nil {
		return nil, err
	}
	maxClients, err := parseIntOptStrict(opts, "rate-limit-max-clients", cacheDefaults.MaxEntries)
	if err != nil {
		return nil, err
	}
	if maxClients <= 0 {
		return nil, fmt.Errorf("'rate-limit-max-clients' must be a positive integer, got %d", maxClients)
	}

	cacheName := cacheDefaults.CacheName
	if scopedName, err := ctx.ScopedName("totp-rate-limit"); err == nil {
		cacheName = scopedName
	}

	return &totp.RateLimitConfig{
		RatePerSecond: ratePerSecond,
		Burst:         burst,
		CacheOptions: &util.CacheOptions{
			Ctx:             ctx.Context(),
			Subsystem:       ctx.CacheSubsystem,
			TrackRate:       true,
			CacheName:       cacheName,
			Owner:           "TOTP",
			Purpose:         "totp-rate-limit",
			Scope:           ctx.ConnectionName(),
			KeyKind:         "client-ip",
			ValueKind:       "token-bucket",
			MaxEntries:      maxClients,
			TTL:             cacheTTL,
			CleanupInterval: cleanupInterval,
		},
	}, nil
}

func parseTOTPSessionConfig(conf Config) (session_store.Config, error) {
	sessionConf := defaultTOTPSessionConfig()

	var err error
	sessionConf.Lifetime, err = parseDurationOption(conf.Options, "session-lifetime", sessionConf.Lifetime)
	if err != nil {
		return session_store.Config{}, err
	}
	sessionConf.IdleTimeout, err = parseDurationOption(conf.Options, "session-idle-timeout", sessionConf.IdleTimeout)
	if err != nil {
		return session_store.Config{}, err
	}
	sessionConf.CleanupInterval, err = parseDurationOption(conf.Options, "session-cleanup-interval", sessionConf.CleanupInterval)
	if err != nil {
		return session_store.Config{}, err
	}

	sessionConf.CookieDomain, err = ParseStringOpt(conf.Options, "cookie-domain", sessionConf.CookieDomain)
	if err != nil {
		return session_store.Config{}, err
	}
	sessionConf.CookieSameSite, err = ParseStringOpt(conf.Options, "cookie-same-site", sessionConf.CookieSameSite)
	if err != nil {
		return session_store.Config{}, err
	}
	sessionConf.CookieSecure = ParseBoolOpt(conf.Options, "cookie-secure", sessionConf.CookieSecure)
	sessionConf.CookieHttpOnly = true
	return sessionConf, nil
}

func defaultTOTPSessionConfig() session_store.Config {
	return session_store.Config{
		Lifetime:        8 * time.Hour,
		IdleTimeout:     30 * time.Minute,
		CleanupInterval: 1 * time.Hour,
		CookieSecure:    false,
		CookieHttpOnly:  true,
		CookieSameSite:  "lax",
	}
}

func parseDurationOption(opts map[string]interface{}, key string, defaultValue time.Duration) (time.Duration, error) {
	raw, err := ParseStringOpt(opts, key, "")
	if err != nil {
		return 0, err
	}
	if raw == "" {
		return defaultValue, nil
	}
	return util.ParseTimeString(raw)
}

func parseFloatOpt(opts map[string]interface{}, key string, defaultValue float64) (float64, error) {
	raw, ok := opts[key]
	if !ok || raw == nil {
		return defaultValue, nil
	}
	switch v := raw.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("'%s' option must be a number, got %T", key, raw)
	}
}

func parseIntOptStrict(opts map[string]interface{}, key string, defaultValue int) (int, error) {
	raw, ok := opts[key]
	if !ok || raw == nil {
		return defaultValue, nil
	}
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("'%s' option must be a number, got %T", key, raw)
	}
}

func acceptsHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return true
	}
	return strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
}
