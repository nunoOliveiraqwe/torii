package util

import (
	"context"
	"testing"
	"time"

	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCacheEntry struct {
	lastSeen time.Time
	allowed  bool
	country  string
}

func (e *testCacheEntry) Touch() {
	e.lastSeen = time.Now()
}

func (e *testCacheEntry) GetLastReadAt() time.Time {
	return e.lastSeen
}

func (e *testCacheEntry) CacheEntryDescriptor() cacheSub.EntryDescriptor {
	disposition := cacheSub.EntryDispositionBlocked
	if e.allowed {
		disposition = cacheSub.EntryDispositionAllowed
	}
	return cacheSub.EntryDescriptor{
		Disposition: disposition,
		Summary:     e.country,
		Fields: map[string]string{
			"country": e.country,
		},
		UpdatedAt: e.lastSeen,
	}
}

func TestParseCacheOptionsUsesConfiguredCacheName(t *testing.T) {
	opts, err := ParseCacheOptions(map[string]interface{}{
		"cache-name":       "geo-cache",
		"cache-ttl":        "1h",
		"cleanup-interval": "1h",
		"max-cache-size":   100,
	})

	require.NoError(t, err)
	assert.Equal(t, "geo-cache", opts.CacheName)
	assert.False(t, opts.IsUsingDefaultCacheName)
}

func TestCacheSnapshotIncludesEntryDescriptors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager := cacheSub.NewSubsystem()
	cache, err := NewCache[*testCacheEntry](&CacheOptions{
		Ctx:             ctx,
		Subsystem:       manager,
		TrackRate:       true,
		CacheName:       "country-block-port-8080",
		Owner:           "CountryBlock",
		Purpose:         "geoip-decision",
		Scope:           "conn-port-8080",
		KeyKind:         "client-ip",
		ValueKind:       "country-decision",
		MaxEntries:      100,
		TTL:             time.Hour,
		CleanupInterval: time.Hour,
	})
	require.NoError(t, err)

	cache.CacheValue("203.0.113.10", &testCacheEntry{allowed: true, country: "PT"})

	snapshots := manager.Snapshots()
	require.Len(t, snapshots, 1)
	snapshot := snapshots[0]
	assert.Equal(t, "country-block-port-8080", snapshot.Name)
	assert.Equal(t, "CountryBlock", snapshot.Owner)
	assert.Equal(t, "geoip-decision", snapshot.Purpose)
	assert.Equal(t, []string{"203.0.113.10"}, snapshot.Keys)
	require.Len(t, snapshot.Entries, 1)
	assert.Equal(t, cacheSub.EntryDispositionAllowed, snapshot.Entries[0].Disposition)
	assert.Equal(t, "PT", snapshot.Entries[0].Fields["country"])
}

func TestCacheUnregistersWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	manager := cacheSub.NewSubsystem()
	_, err := NewCache[*testCacheEntry](&CacheOptions{
		Ctx:             ctx,
		Subsystem:       manager,
		CacheName:       "temporary-cache",
		MaxEntries:      10,
		TTL:             time.Hour,
		CleanupInterval: time.Hour,
	})
	require.NoError(t, err)
	require.Len(t, manager.Snapshots(), 1)

	cancel()

	require.Eventually(t, func() bool {
		return len(manager.Snapshots()) == 0
	}, time.Second, 10*time.Millisecond)
}
