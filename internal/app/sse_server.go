package app

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nunoOliveiraqwe/torii/internal/subsystem/activity"
	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"github.com/nunoOliveiraqwe/torii/metrics"
	"go.uber.org/zap"
)

type SSEEvent struct {
	Type string
	Data []byte
}

type SSEClient struct {
	ID     string
	Events chan SSEEvent
}

type SSEBroker struct {
	mu          sync.RWMutex
	clients     map[string]*SSEClient
	nextID      int
	clientCount atomic.Int64

	mgr      *metrics.ConnectionMetricsManager
	activity *activity.Subsystem
	cache    *cacheSub.Subsystem

	listenersMu       sync.Mutex
	listeners         map[string]int
	metricsListenerID int
	errorListenerID   int
	requestListenerID int
	blockedListenerID int

	cacheUnsubscribe cacheSub.UnsubscribeFunc
}

func NewSSEBroker(mgr *metrics.ConnectionMetricsManager, activitySubsystem *activity.Subsystem, cacheSubsystem *cacheSub.Subsystem) *SSEBroker {
	b := &SSEBroker{
		clients:   make(map[string]*SSEClient),
		mgr:       mgr,
		activity:  activitySubsystem,
		cache:     cacheSubsystem,
		listeners: make(map[string]int),
	}
	// Wildcard metric listener — fires for every connection's metric updates.
	b.metricsListenerID = mgr.AddWildcardListener(func(_ string, snapshot *metrics.Metric) {
		b.broadcastJSON("metrics", snapshot)
	})
	if activitySubsystem != nil {
		b.errorListenerID = activitySubsystem.ErrorLog.AddListener(func(entry *activity.ErrorLogEntry) {
			b.broadcastJSON("proxy_error", entry)
		})
		b.requestListenerID = activitySubsystem.RequestLog.AddListener(func(entry *activity.RequestLogEntry) {
			b.broadcastJSON("proxy_request", entry)
		})
		b.blockedListenerID = activitySubsystem.BlockLog.AddListener(func(entry *activity.BlockLogEntry) {
			b.broadcastJSON("proxy_blocked", entry)
		})
	}
	if cacheSubsystem != nil {
		b.cacheUnsubscribe = cacheSubsystem.AddChangeListener(b.signalCacheUpdate)
	}
	return b
}

func (b *SSEBroker) Subscribe() *SSEClient {
	b.mu.Lock()
	id := fmt.Sprintf("sse-%d", b.nextID)
	b.nextID++
	client := &SSEClient{
		ID:     id,
		Events: make(chan SSEEvent, 512),
	}
	b.clients[id] = client
	b.clientCount.Add(1)
	b.mu.Unlock()

	allMetrics := b.mgr.GetAllMetrics()
	for _, metric := range allMetrics {
		if metric != nil {
			if data, err := json.Marshal(metric); err == nil {
				select {
				case client.Events <- SSEEvent{Type: "metrics", Data: data}:
				default:
				}
			}
		}
	}
	if b.cache != nil {
		if data, err := json.Marshal(b.cache.Snapshots()); err == nil {
			select {
			case client.Events <- SSEEvent{Type: "cache_snapshots", Data: data}:
			default:
			}
		}
	}
	zap.S().Debugf("SSEBroker: client %s subscribed", id)
	return client
}

func (b *SSEBroker) Unsubscribe(client *SSEClient) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[client.ID]; ok {
		close(client.Events)
		delete(b.clients, client.ID)
		b.clientCount.Add(-1)
		zap.S().Debugf("SSEBroker: client %s unsubscribed", client.ID)
	}
}

func (b *SSEBroker) broadcastJSON(eventType string, v any) {
	if b.clientCount.Load() == 0 {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		zap.S().Errorf("SSEBroker: failed to marshal %s event: %v", eventType, err)
		return
	}
	b.broadcastAll(SSEEvent{Type: eventType, Data: data})
}

func (b *SSEBroker) broadcastAll(event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, client := range b.clients {
		select {
		case client.Events <- event:
		default:
			zap.S().Warnf("SSEBroker: dropping event for slow client %s", client.ID)
		}
	}
}

func (b *SSEBroker) signalCacheUpdate() {
	b.broadcastCacheSnapshots()
}

func (b *SSEBroker) broadcastCacheSnapshots() {
	if b.cache == nil || b.clientCount.Load() == 0 {
		return
	}
	b.broadcastJSON("cache_snapshots", b.cache.Snapshots())
}

func (b *SSEBroker) Stop() {

	if b.cacheUnsubscribe != nil {
		b.cacheUnsubscribe()
		b.cacheUnsubscribe = nil
	}

	b.mgr.RemoveListener(b.metricsListenerID)
	if b.activity != nil {
		b.activity.ErrorLog.RemoveListener(b.errorListenerID)
		b.activity.RequestLog.RemoveListener(b.requestListenerID)
		b.activity.BlockLog.RemoveListener(b.blockedListenerID)
	}
	b.mu.Lock()
	for _, c := range b.clients {
		close(c.Events)
	}
	b.clients = make(map[string]*SSEClient)
	b.clientCount.Store(0)
	b.mu.Unlock()

	zap.S().Info("SSEBroker stopped")
}
