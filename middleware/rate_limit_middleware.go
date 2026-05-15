package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/ratelimit"
	"github.com/nunoOliveiraqwe/torii/internal/requestctx"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"go.uber.org/zap"
)

type rateLimitConf struct {
	Burst         int
	RatePerSecond float64
	Mode          string
	CacheOpt      *util.CacheOptions
}

func RateLimitMiddleware(ctx BuildContext, next http.HandlerFunc, conf Config) http.HandlerFunc {
	limitConf, err := parseRateLimitConfig(ctx, conf)
	if err != nil {
		zap.S().Errorf("Failed to parse rate limit middleware configuration: %v. Failing closed.", err)
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "rate limiter misconfigured", http.StatusServiceUnavailable)
		}
	}
	l, err := newLimiter(limitConf)
	if err != nil {
		zap.S().Errorf("Failed to initialize rate limiter: %v. Failing closed.", err)
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "rate limiter misconfigured", http.StatusServiceUnavailable)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		decision := l.Allow(r)
		if !decision.Allowed {
			if decision.Err != nil {
				GetRequestLoggerFromContext(r).Warn("RateLimitMiddleware: request denied by limiter",
					zap.String("key", decision.Key),
					zap.Error(decision.Err))
			}
			if decision.RetryAfter != "" {
				w.Header().Set("Retry-After", decision.RetryAfter)
			}
			requestctx.CreateAndAddBlockInfoToRequestContext(r, "rate-limit", decision.Reason, bus.TopicRateLimitTriggered)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func computeRetryAfter(ratePerSecond float64) string {
	return ratelimit.ComputeRetryAfter(ratePerSecond)
}

func newLimiter(c *rateLimitConf) (*ratelimit.Limiter, error) {
	return ratelimit.New(ratelimit.Config{
		Mode:          ratelimit.Mode(c.Mode),
		RatePerSecond: c.RatePerSecond,
		Burst:         c.Burst,
		CacheOptions:  c.CacheOpt,
		ReasonPrefix:  "rate limit",
	})
}

func parseRateLimitConfig(ctx BuildContext, conf Config) (*rateLimitConf, error) {
	zap.S().Debug("Parsing rate limit middleware configuration")
	reqLimit, ok := conf.Options["limiter-req"]
	if !ok {
		zap.S().Error("RateLimitMiddleware: missing required option 'limiter-req'")
		return nil, fmt.Errorf("missing required option 'limiter-req'")
	}
	reqLimitMap, ok := reqLimit.(map[string]interface{})
	if !ok {
		zap.S().Error("RateLimitMiddleware: 'limiter-req' option must be a map[string]interface{}")
		return nil, fmt.Errorf("'limiter-req' option must be a map[string]interface{}")
	}

	burstInt, err := ParseIntOptRequired(reqLimitMap, "burst")
	if err != nil {
		return nil, err
	}
	if burstInt <= 0 {
		return nil, fmt.Errorf("'burst' must be a positive integer, got %d", burstInt)
	}

	ratePsF, err := ParseFloatOptRequired(reqLimitMap, "rate-per-second")
	if err != nil {
		return nil, err
	}
	if ratePsF <= 0 {
		return nil, fmt.Errorf("'rate-per-second' must be a positive number, got %f", ratePsF)
	}

	modeStr, err := ParseStringOpt(conf.Options, "mode", "global")
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(modeStr, "global") && !strings.EqualFold(modeStr, "per-client") {
		zap.S().Errorf("RateLimitMiddleware: invalid 'mode' option value '%s', must be 'global' or 'per-client'", modeStr)
		return nil, fmt.Errorf("invalid 'mode' option value '%s', must be 'global' or 'per-client'", modeStr)
	}

	var cacheOpts *util.CacheOptions
	if strings.EqualFold(modeStr, "per-client") {
		cacheOpts, err = ParseMiddlewareCacheOptions(ctx, conf, cacheRuntimeOptions{
			Owner:      "RateLimiter",
			Purpose:    "rate-limit",
			NamePrefix: "rate-limit",
			KeyKind:    "client-ip",
			ValueKind:  "token-bucket",
			TrackRate:  true,
		})
		if err != nil {
			zap.S().Errorf("Failed to parse cache options for per-client rate limiter: %v. Failing closed.", err)
			return nil, fmt.Errorf("failed to parse cache options: %w", err)
		}
	}
	return &rateLimitConf{
		CacheOpt:      cacheOpts,
		Burst:         burstInt,
		RatePerSecond: ratePsF,
		Mode:          modeStr,
	}, nil
}
