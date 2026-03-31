package middleware

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	defaultMaxClients = 10000
	defaultClientTTL  = 5 * time.Minute
	cleanupInterval   = 1 * time.Minute
)

type limiter interface {
	limit(r *http.Request, w http.ResponseWriter) bool
}

type globalLimiter struct {
	internalLimiter *rate.Limiter
	retryAfter      string
}

type clientEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type perClientLimiter struct {
	mu            sync.RWMutex
	clients       map[string]*clientEntry
	ratePerSecond float64
	burst         int
	retryAfter    string
	maxEntries    int
	ttl           time.Duration
}

func (g *globalLimiter) limit(_ *http.Request, w http.ResponseWriter) bool {
	if g.internalLimiter.Allow() {
		return true
	}
	w.Header().Set("Retry-After", g.retryAfter)
	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
	return false
}

func getClientIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse RemoteAddr %q: %w", r.RemoteAddr, err)
	}
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	return ip, nil
}

func (l *perClientLimiter) limit(r *http.Request, w http.ResponseWriter) bool {
	ipAddr, err := getClientIP(r)
	if err != nil {
		zap.S().Warnf("RateLimitMiddleware: failed to extract client IP: %v", err)
		w.Header().Set("Retry-After", l.retryAfter)
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return false
	}

	entry := l.getOrCreate(ipAddr)
	entry.lastSeen = time.Now()

	if entry.limiter.Allow() {
		return true
	}
	w.Header().Set("Retry-After", l.retryAfter)
	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
	return false
}

func (l *perClientLimiter) getOrCreate(ip string) *clientEntry {
	l.mu.RLock()
	entry, exists := l.clients[ip]
	l.mu.RUnlock()
	if exists {
		return entry
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	entry, exists = l.clients[ip]
	if exists {
		return entry
	}
	if len(l.clients) >= l.maxEntries {
		l.evictOldestLocked()
	}
	entry = &clientEntry{
		limiter:  rate.NewLimiter(rate.Limit(l.ratePerSecond), l.burst),
		lastSeen: time.Now(),
	}
	l.clients[ip] = entry
	return entry
}

func (l *perClientLimiter) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range l.clients {
		if first || v.lastSeen.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.lastSeen
			first = false
		}
	}
	if oldestKey != "" {
		delete(l.clients, oldestKey)
	}
}

func (l *perClientLimiter) startCleanup() {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			l.sweep()
		}
	}()
}

func (l *perClientLimiter) sweep() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-l.ttl)
	for ip, entry := range l.clients {
		if entry.lastSeen.Before(cutoff) {
			delete(l.clients, ip)
		}
	}
}

type rateLimitConf struct {
	Burst         int
	RatePerSecond float64
	Mode          string
}

func RateLimitMiddleware(_ context.Context, next http.HandlerFunc, conf Config) http.HandlerFunc {
	limitConf, err := parseConfig(conf)
	if err != nil {
		zap.S().Errorf("Failed to parse rate limit middleware configuration: %v. Failing closed.", err)
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "rate limiter misconfigured", http.StatusServiceUnavailable)
		}
	}
	l := newLimiter(limitConf)
	return func(w http.ResponseWriter, r *http.Request) {
		if l.limit(r, w) {
			next.ServeHTTP(w, r)
		}
	}
}

func computeRetryAfter(ratePerSecond float64) string {
	if ratePerSecond <= 0 {
		return "1"
	}
	secs := int(math.Ceil(1.0 / ratePerSecond))
	if secs < 1 {
		secs = 1
	}
	return strconv.Itoa(secs)
}

func newLimiter(c *rateLimitConf) limiter {
	retryAfter := computeRetryAfter(c.RatePerSecond)
	if strings.EqualFold(c.Mode, "per-client") {
		pcl := &perClientLimiter{
			clients:       make(map[string]*clientEntry),
			ratePerSecond: c.RatePerSecond,
			burst:         c.Burst,
			retryAfter:    retryAfter,
			maxEntries:    defaultMaxClients,
			ttl:           defaultClientTTL,
		}
		pcl.startCleanup()
		return pcl
	}
	return &globalLimiter{
		internalLimiter: rate.NewLimiter(rate.Limit(c.RatePerSecond), c.Burst),
		retryAfter:      retryAfter,
	}
}

func parseConfig(conf Config) (*rateLimitConf, error) {
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

	burstInt, err := parseIntOption(reqLimitMap, "burst")
	if err != nil {
		return nil, err
	}
	if burstInt <= 0 {
		return nil, fmt.Errorf("'burst' must be a positive integer, got %d", burstInt)
	}

	ratePsF, err := parseFloatOption(reqLimitMap, "rate-per-second")
	if err != nil {
		return nil, err
	}
	if ratePsF <= 0 {
		return nil, fmt.Errorf("'rate-per-second' must be a positive number, got %f", ratePsF)
	}

	mode, ok := conf.Options["mode"]
	if !ok {
		zap.S().Warn("RateLimitMiddleware: 'mode' option not specified, defaulting to 'global'")
		mode = "global"
	}
	modeStr, ok := mode.(string)
	if !ok {
		zap.S().Errorf("RateLimitMiddleware: 'mode' option must be a string")
		return nil, fmt.Errorf("'mode' option must be a string")
	}

	if !strings.EqualFold(modeStr, "global") && !strings.EqualFold(modeStr, "per-client") {
		zap.S().Errorf("RateLimitMiddleware: invalid 'mode' option value '%s', must be 'global' or 'per-client'", modeStr)
		return nil, fmt.Errorf("invalid 'mode' option value '%s', must be 'global' or 'per-client'", modeStr)
	}
	return &rateLimitConf{
		Burst:         burstInt,
		RatePerSecond: ratePsF,
		Mode:          modeStr,
	}, nil
}

func parseIntOption(m map[string]interface{}, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing required option '%s' in 'limiter-req'", key)
	}
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case int:
		return val, nil
	default:
		return 0, fmt.Errorf("'%s' option in 'limiter-req' must be a number", key)
	}
}

func parseFloatOption(m map[string]interface{}, key string) (float64, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing required option '%s' in 'limiter-req'", key)
	}
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("'%s' option in 'limiter-req' must be a number", key)
	}
}
