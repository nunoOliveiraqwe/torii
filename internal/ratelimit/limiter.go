package ratelimit

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/netutil"
	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Mode string

const (
	ModeGlobal    Mode = "global"
	ModePerClient Mode = "per-client"
)

type Config struct {
	Mode          Mode
	RatePerSecond float64
	Burst         int
	CacheOptions  *util.CacheOptions
	ReasonPrefix  string
}

type Decision struct {
	Allowed    bool
	RetryAfter string
	Key        string
	Reason     string
	Err        error
}

type Limiter struct {
	mode          Mode
	globalLimiter *rate.Limiter
	clientCache   *util.Cache[*clientEntry]
	ratePerSecond float64
	burst         int
	retryAfter    string
	reasonPrefix  string
}

type clientEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func New(conf Config) (*Limiter, error) {
	if conf.RatePerSecond <= 0 {
		return nil, fmt.Errorf("rate-per-second must be positive, got %f", conf.RatePerSecond)
	}
	if conf.Burst <= 0 {
		return nil, fmt.Errorf("burst must be positive, got %d", conf.Burst)
	}

	mode, err := normalizeMode(conf.Mode)
	if err != nil {
		return nil, err
	}

	reasonPrefix := conf.ReasonPrefix
	if reasonPrefix == "" {
		reasonPrefix = "rate limit"
	}

	l := &Limiter{
		mode:          mode,
		ratePerSecond: conf.RatePerSecond,
		burst:         conf.Burst,
		retryAfter:    ComputeRetryAfter(conf.RatePerSecond),
		reasonPrefix:  reasonPrefix,
	}

	if mode == ModePerClient {
		if conf.CacheOptions == nil {
			return nil, fmt.Errorf("cache options are required for per-client rate limiting")
		}
		cache, err := util.NewCache[*clientEntry](conf.CacheOptions)
		if err != nil {
			return nil, err
		}
		l.clientCache = cache
		return l, nil
	}

	l.globalLimiter = rate.NewLimiter(rate.Limit(conf.RatePerSecond), conf.Burst)
	return l, nil
}

func (l *Limiter) Allow(r *http.Request) Decision {
	if l.mode == ModeGlobal {
		return l.allowGlobal()
	}
	return l.allowClient(r)
}

func (l *Limiter) Reset(r *http.Request) {
	if l == nil || l.mode != ModePerClient || l.clientCache == nil {
		return
	}
	clientIP, err := netutil.GetClientIP(r)
	if err != nil {
		return
	}
	l.clientCache.Evict(clientIP)
}

func ComputeRetryAfter(ratePerSecond float64) string {
	if ratePerSecond <= 0 {
		return "1"
	}
	secs := int(math.Ceil(1.0 / ratePerSecond))
	if secs < 1 {
		secs = 1
	}
	return strconv.Itoa(secs)
}

func normalizeMode(mode Mode) (Mode, error) {
	if mode == "" {
		return ModeGlobal, nil
	}
	if strings.EqualFold(string(mode), string(ModeGlobal)) {
		return ModeGlobal, nil
	}
	if strings.EqualFold(string(mode), string(ModePerClient)) {
		return ModePerClient, nil
	}
	return "", fmt.Errorf("invalid rate limit mode %q", mode)
}

func (l *Limiter) allowGlobal() Decision {
	if l.globalLimiter.Allow() {
		return Decision{Allowed: true}
	}
	return Decision{
		Allowed:    false,
		RetryAfter: l.retryAfter,
		Reason:     fmt.Sprintf("global %s of %g req/s exceeded", l.reasonPrefix, float64(l.globalLimiter.Limit())),
	}
}

func (l *Limiter) allowClient(r *http.Request) Decision {
	clientIP, err := netutil.GetClientIP(r)
	if err != nil {
		zap.S().Warnf("RateLimiter: failed to extract client IP: %v", err)
		return Decision{
			Allowed:    false,
			RetryAfter: l.retryAfter,
			Reason:     fmt.Sprintf("failed to extract client IP for %s", l.reasonPrefix),
			Err:        err,
		}
	}

	entry, err := l.clientCache.GetValue(clientIP)
	if err != nil && errors.Is(err, util.ErrCacheMiss) {
		entry = &clientEntry{
			limiter: rate.NewLimiter(rate.Limit(l.ratePerSecond), l.burst),
		}
		l.clientCache.CacheValue(clientIP, entry)
	} else if err != nil {
		zap.S().Warnf("RateLimiter: failed to get client entry from cache for IP %s: %v", clientIP, err)
		return Decision{
			Allowed:    false,
			RetryAfter: l.retryAfter,
			Key:        clientIP,
			Reason:     fmt.Sprintf("cache error in %s", l.reasonPrefix),
			Err:        err,
		}
	}

	if entry.limiter.Allow() {
		return Decision{Allowed: true, Key: clientIP}
	}
	return Decision{
		Allowed:    false,
		RetryAfter: l.retryAfter,
		Key:        clientIP,
		Reason:     fmt.Sprintf("%s of %g req/s exceeded", l.reasonPrefix, float64(entry.limiter.Limit())),
	}
}

func (e *clientEntry) Touch() {
	e.lastSeen = time.Now()
}

func (e *clientEntry) GetLastReadAt() time.Time {
	return e.lastSeen
}

func (e *clientEntry) CacheEntryDescriptor() cacheSub.EntryDescriptor {
	fields := map[string]string{}
	summary := "per-client limiter"
	if e.limiter != nil {
		fields["rate_per_second"] = fmt.Sprintf("%g", float64(e.limiter.Limit()))
		fields["burst"] = strconv.Itoa(e.limiter.Burst())
		summary = fmt.Sprintf("%g req/s, burst %d", float64(e.limiter.Limit()), e.limiter.Burst())
	}
	return cacheSub.EntryDescriptor{
		Disposition: cacheSub.EntryDispositionNeutral,
		Summary:     summary,
		Fields:      fields,
		UpdatedAt:   e.lastSeen,
	}
}
