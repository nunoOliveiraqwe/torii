package totp

import (
	"fmt"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/ratelimit"
	"github.com/nunoOliveiraqwe/torii/internal/util"
)

const (
	defaultRateLimitRatePerSecond = 1.0 / 12.0
	defaultRateLimitBurst         = 5
	defaultRateLimitTTL           = time.Hour
	defaultRateLimitCleanup       = 10 * time.Minute
	defaultRateLimitMaxClients    = 100000
)

type RateLimitConfig struct {
	Disabled      bool
	RatePerSecond float64
	Burst         int
	CacheOptions  *util.CacheOptions
}

type RateLimitError struct {
	RetryAfter string
	Reason     string
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RatePerSecond: defaultRateLimitRatePerSecond,
		Burst:         defaultRateLimitBurst,
		CacheOptions:  defaultRateLimitCacheOptions(),
	}
}

func (e *RateLimitError) Error() string {
	if e.Reason != "" {
		return e.Reason
	}
	return "TOTP verification rate limit exceeded"
}

func newVerificationRateLimiter(raw *RateLimitConfig) (*ratelimit.Limiter, error) {
	cfg, err := normalizeRateLimitConfig(raw)
	if err != nil {
		return nil, err
	}
	if cfg.Disabled {
		return nil, nil
	}
	return ratelimit.New(ratelimit.Config{
		Mode:          ratelimit.ModePerClient,
		RatePerSecond: cfg.RatePerSecond,
		Burst:         cfg.Burst,
		CacheOptions:  cfg.CacheOptions,
		ReasonPrefix:  "TOTP verification rate limit",
	})
}

func normalizeRateLimitConfig(raw *RateLimitConfig) (RateLimitConfig, error) {
	cfg := DefaultRateLimitConfig()
	if raw != nil {
		cfg.Disabled = raw.Disabled
		if raw.RatePerSecond != 0 {
			cfg.RatePerSecond = raw.RatePerSecond
		}
		if raw.Burst != 0 {
			cfg.Burst = raw.Burst
		}
		if raw.CacheOptions != nil {
			cfg.CacheOptions = raw.CacheOptions
		}
	}
	if cfg.Disabled {
		return cfg, nil
	}
	if cfg.RatePerSecond <= 0 {
		return cfg, fmt.Errorf("TOTP rate limit rate-per-second must be positive, got %f", cfg.RatePerSecond)
	}
	if cfg.Burst <= 0 {
		return cfg, fmt.Errorf("TOTP rate limit burst must be positive, got %d", cfg.Burst)
	}
	if cfg.CacheOptions == nil {
		cfg.CacheOptions = defaultRateLimitCacheOptions()
	}
	if cfg.CacheOptions.MaxEntries <= 0 {
		return cfg, fmt.Errorf("TOTP rate limit max cache size must be positive, got %d", cfg.CacheOptions.MaxEntries)
	}
	if cfg.CacheOptions.TTL <= 0 {
		return cfg, fmt.Errorf("TOTP rate limit cache TTL must be positive, got %s", cfg.CacheOptions.TTL)
	}
	if cfg.CacheOptions.CleanupInterval <= 0 {
		return cfg, fmt.Errorf("TOTP rate limit cleanup interval must be positive, got %s", cfg.CacheOptions.CleanupInterval)
	}
	return cfg, nil
}

func defaultRateLimitCacheOptions() *util.CacheOptions {
	return &util.CacheOptions{
		CacheName:       fmt.Sprintf("totp-rate-limit-%d", time.Now().UnixNano()),
		MaxEntries:      defaultRateLimitMaxClients,
		TTL:             defaultRateLimitTTL,
		CleanupInterval: defaultRateLimitCleanup,
	}
}
