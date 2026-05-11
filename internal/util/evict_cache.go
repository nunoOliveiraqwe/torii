package util

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"go.uber.org/zap"
)

const (
	defaultCacheTTL        = "1h"
	defaultCleanupInterval = "1h"
	defaultMaxEntries      = 100000
)

var ErrCacheMiss = fmt.Errorf("cache miss")

type EvictableEntry interface {
	Touch()
	GetLastReadAt() time.Time
}

type CacheOptions struct {
	Ctx                     context.Context
	Subsystem               *cacheSub.Subsystem // if defined cache is registered with subsystem
	IsUsingDefaultCacheName bool
	TrackRate               bool
	CacheName               string
	Owner                   string
	Purpose                 string
	Scope                   string
	KeyKind                 string
	ValueKind               string
	MaxEntries              int
	TTL                     time.Duration
	CleanupInterval         time.Duration
}

type Cache[T EvictableEntry] struct {
	cacheName       string
	owner           string
	purpose         string
	scope           string
	keyKind         string
	valueKind       string
	ctx             context.Context
	maxEntries      int
	ttl             time.Duration
	cleanupInterval time.Duration
	mu              sync.RWMutex
	cache           map[string]T
	tracker         Tracker
	unregister      func()
	notifyChanged   func()
	closeOnce       sync.Once
}

func ParseCacheOptions(configMap map[string]interface{}) (*CacheOptions, error) {
	cacheTtlStr := defaultCacheTTL
	cacheTtl, ok := configMap["cache-ttl"]
	if !ok {
		zap.S().Warnf("'cache-ttl' option not specified, defaulting to %s", cacheTtlStr)
	} else if str, ok := cacheTtl.(string); ok {
		cacheTtlStr = str
	}
	ttl, err := ParseTimeString(cacheTtlStr)
	if err != nil {
		zap.S().Errorf("Failed to parse TTL string %q: %v. Cannot initialize cache", cacheTtlStr, err)
		return nil, fmt.Errorf("invalid TTL string: %w", err)
	}

	cleanUpIntervalStr := defaultCleanupInterval
	cleanUpInterval, ok := configMap["cleanup-interval"]
	if !ok {
		zap.S().Warnf("'cleanup-interval' option not specified, defaulting to %s", cleanUpIntervalStr)
	} else if cleanUp, ok := cleanUpInterval.(string); ok {
		cleanUpIntervalStr = cleanUp
	}

	cleanupInterval, err2 := ParseTimeString(cleanUpIntervalStr)
	if err2 != nil {
		zap.S().Errorf("Failed to parse cleanup interval string %q: %v. Cannot initialize cache", cleanUpIntervalStr, err2)
		return nil, fmt.Errorf("invalid cleanup interval string: %w", err2)
	}

	maxCacheSize := defaultMaxEntries
	maxCache, ok := configMap["max-cache-size"]
	if !ok {
		zap.S().Warnf("'max-cache-size' option not specified, defaulting to %d", maxCacheSize)
	} else if m, ok := maxCache.(int); ok {
		maxCacheSize = m
	}

	isUsingDefaultCacheName := true
	cacheNameStr := fmt.Sprintf("cache-%d", time.Now().UnixNano())
	cacheName, ok := configMap["cache-name"]
	if ok {
		if m, ok := cacheName.(string); ok {
			cacheNameStr = m
			isUsingDefaultCacheName = false
		}
	}
	return &CacheOptions{
		IsUsingDefaultCacheName: isUsingDefaultCacheName,
		CacheName:               cacheNameStr,
		MaxEntries:              maxCacheSize,
		TTL:                     ttl,
		CleanupInterval:         cleanupInterval,
	}, nil
}

func NewCache[T EvictableEntry](options *CacheOptions) (*Cache[T], error) {
	rateTracker := NewNopRateTracker()
	if options.TrackRate {
		rateTracker = NewRateTracker()
	}
	cacheCtx := options.Ctx
	if cacheCtx == nil {
		cacheCtx = context.Background()
	}
	c := &Cache[T]{
		cacheName:       options.CacheName,
		owner:           options.Owner,
		purpose:         options.Purpose,
		scope:           options.Scope,
		keyKind:         options.KeyKind,
		valueKind:       options.ValueKind,
		ctx:             cacheCtx,
		cleanupInterval: options.CleanupInterval,
		maxEntries:      options.MaxEntries,
		mu:              sync.RWMutex{},
		ttl:             options.TTL,
		cache:           make(map[string]T),
		tracker:         rateTracker,
	}
	if options.Subsystem != nil {
		c.unregister = options.Subsystem.RegisterCache(c)
		c.notifyChanged = options.Subsystem.NotifyChanged
	}
	c.startCleanup()
	return c, nil
}

func (c *Cache[T]) CacheValue(ip string, value T) {
	c.mu.Lock()
	if len(c.cache) >= c.maxEntries {
		c.evictOldest()
	}
	value.Touch()
	c.cache[ip] = value
	c.tracker.Mark(1)
	c.mu.Unlock()
	c.notifyCacheChanged()
}

func (c *Cache[T]) GetValue(key string) (T, error) {
	c.mu.Lock()
	entry, ok := c.cache[key]
	if !ok {
		c.tracker.MarkMiss()
		c.mu.Unlock()
		c.notifyCacheChanged()
		var zero T
		return zero, ErrCacheMiss
	}
	if time.Since(entry.GetLastReadAt()) > c.ttl {
		delete(c.cache, key)
		c.tracker.MarkMiss()
		c.mu.Unlock()
		c.notifyCacheChanged()
		var zero T
		return zero, ErrCacheMiss
	}
	c.tracker.MarkHit()
	entry.Touch()
	c.mu.Unlock()
	c.notifyCacheChanged()
	return entry, nil
}

func (c *Cache[T]) Evict(key string) {
	c.mu.Lock()
	_, existed := c.cache[key]
	if existed {
		delete(c.cache, key)
	}
	c.mu.Unlock()
	if existed {
		c.notifyCacheChanged()
	}
}

// only supposed to be called after a lock is held
func (c *Cache[T]) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.cache {
		if first || entry.GetLastReadAt().Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.GetLastReadAt()
			first = false
		}
	}

	if !first {
		delete(c.cache, oldestKey)
	}
}

func (c *Cache[T]) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	go func() {
		defer ticker.Stop()
		defer c.close()
		for {
			select {
			case <-c.ctx.Done():
				zap.S().Info("Cache cleanup goroutine exiting due to context cancellation")
				return
			case <-ticker.C:
				c.sweep()
			}
		}
	}()
}

func (c *Cache[T]) close() {
	c.closeOnce.Do(func() {
		c.tracker.Stop()
		if c.unregister != nil {
			c.unregister()
		}
	})
}

func (c *Cache[T]) sweep() {
	c.mu.Lock()
	cutoff := time.Now().Add(-c.ttl)
	removed := false
	for ip, entry := range c.cache {
		if entry.GetLastReadAt().Before(cutoff) {
			delete(c.cache, ip)
			removed = true
		}
	}
	c.mu.Unlock()
	if removed {
		c.notifyCacheChanged()
	}
}

func (c *Cache[T]) notifyCacheChanged() {
	if c.notifyChanged != nil {
		c.notifyChanged()
	}
}

func (c *Cache[T]) CacheName() string {
	return c.cacheName
}

func (c *Cache[T]) GetMaxLen() int {
	return c.maxEntries
}

func (c *Cache[T]) GetCurrentNumberOfElements() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

func (c *Cache[T]) GetAllKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}

func (c *Cache[T]) CacheInsertionM1Rate() float64 {
	return c.tracker.M1Rate()
}

func (c *Cache[T]) CacheInsertionM5Rate() float64 {
	return c.tracker.M5Rate()
}

func (c *Cache[T]) CacheInsertionM15Rate() float64 {
	return c.tracker.M15Rate()
}

func (c *Cache[T]) CacheInsertionTotal() int64 {
	return c.tracker.Total()
}

func (c *Cache[T]) CacheHits() int64 {
	return c.tracker.Hits()
}

func (c *Cache[T]) CacheMisses() int64 {
	return c.tracker.Misses()
}

func (c *Cache[T]) Snapshot() cacheSub.SourceSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]cacheSub.EntrySnapshot, 0, len(keys))
	for _, key := range keys {
		entry := c.cache[key]
		entrySnapshot := cacheSub.EntrySnapshot{
			Key:         key,
			Disposition: cacheSub.EntryDispositionUnknown,
			LastSeenAt:  entry.GetLastReadAt(),
		}
		if describer, ok := any(entry).(cacheSub.EntryDescriber); ok {
			descriptor := describer.CacheEntryDescriptor()
			entrySnapshot.Disposition = descriptor.Disposition
			entrySnapshot.Summary = descriptor.Summary
			entrySnapshot.Fields = descriptor.Fields
			entrySnapshot.Tags = descriptor.Tags
			entrySnapshot.UpdatedAt = descriptor.UpdatedAt
		}
		if entrySnapshot.Disposition == "" {
			entrySnapshot.Disposition = cacheSub.EntryDispositionUnknown
		}
		entries = append(entries, entrySnapshot)
	}

	rates := cacheSub.RateSnapshot{
		M1Rate:         c.tracker.M1Rate(),
		M5Rate:         c.tracker.M5Rate(),
		M15Rate:        c.tracker.M15Rate(),
		InsertionTotal: c.tracker.Total(),
		Hits:           c.tracker.Hits(),
		Misses:         c.tracker.Misses(),
	}

	return cacheSub.SourceSnapshot{
		Name:            c.cacheName,
		Owner:           c.owner,
		Purpose:         c.purpose,
		Scope:           c.scope,
		KeyKind:         c.keyKind,
		ValueKind:       c.valueKind,
		MaxEntries:      c.maxEntries,
		CurrentEntries:  len(c.cache),
		TTL:             c.ttl.String(),
		CleanupInterval: c.cleanupInterval.String(),
		Keys:            keys,
		Entries:         entries,
		Rates:           rates,
		M1Rate:          rates.M1Rate,
		Hits:            rates.Hits,
		Misses:          rates.Misses,
		InsertionTotal:  rates.InsertionTotal,
	}
}
