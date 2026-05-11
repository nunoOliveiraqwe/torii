package cache

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type EntryDisposition string

const (
	EntryDispositionUnknown EntryDisposition = "unknown"
	EntryDispositionAllowed EntryDisposition = "allowed"
	EntryDispositionBlocked EntryDisposition = "blocked"
	EntryDispositionNeutral EntryDisposition = "neutral"
)

type EntryDescriptor struct {
	Disposition EntryDisposition  `json:"disposition"`
	Summary     string            `json:"summary,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}

type EntryDescriber interface {
	CacheEntryDescriptor() EntryDescriptor
}

type EntrySnapshot struct {
	Key         string            `json:"key"`
	Disposition EntryDisposition  `json:"disposition"`
	Summary     string            `json:"summary,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	LastSeenAt  time.Time         `json:"last_seen_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}

type RateSnapshot struct {
	M1Rate         float64 `json:"m1_rate"`
	M5Rate         float64 `json:"m5_rate"`
	M15Rate        float64 `json:"m15_rate"`
	InsertionTotal int64   `json:"insertion_total"`
	Hits           int64   `json:"hits"`
	Misses         int64   `json:"misses"`
}

type SourceSnapshot struct {
	Name            string          `json:"name"`
	Owner           string          `json:"owner,omitempty"`
	Purpose         string          `json:"purpose,omitempty"`
	Scope           string          `json:"scope,omitempty"`
	KeyKind         string          `json:"key_kind,omitempty"`
	ValueKind       string          `json:"value_kind,omitempty"`
	MaxEntries      int             `json:"max_entries"`
	CurrentEntries  int             `json:"current_entries"`
	TTL             string          `json:"ttl,omitempty"`
	CleanupInterval string          `json:"cleanup_interval,omitempty"`
	Keys            []string        `json:"keys"`
	Entries         []EntrySnapshot `json:"entries"`
	Rates           RateSnapshot    `json:"rates"`

	M1Rate         float64 `json:"m1_rate"`
	Hits           int64   `json:"hits"`
	Misses         int64   `json:"misses"`
	InsertionTotal int64   `json:"insertion_total"`
}

type Source interface {
	CacheName() string
	Snapshot() SourceSnapshot
}

type ChangeListener func()

type UnsubscribeFunc func()

type Subsystem struct {
	mu             sync.RWMutex
	nextID         uint64
	nextListenerID uint64
	caches         map[string]Source
	order          []string
	listeners      map[uint64]ChangeListener
	started        bool
}

func NewSubsystem() *Subsystem {
	return &Subsystem{
		caches:    make(map[string]Source),
		listeners: make(map[uint64]ChangeListener),
	}
}

func (s *Subsystem) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = true
	return nil
}

func (s *Subsystem) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = false
	s.caches = make(map[string]Source)
	s.order = nil
	s.listeners = make(map[uint64]ChangeListener)
	return nil
}

func (s *Subsystem) RegisterCache(source Source) func() {
	if s == nil || source == nil {
		return func() {}
	}

	s.mu.Lock()
	s.nextID++
	id := fmt.Sprintf("%s#%d", source.CacheName(), s.nextID)
	s.caches[id] = source
	s.order = append(s.order, id)
	s.mu.Unlock()

	s.NotifyChanged()

	return func() {
		s.unregister(id)
	}
}

func (s *Subsystem) RegisterCacheSource(source Source) func() {
	return s.RegisterCache(source)
}

func (s *Subsystem) unregister(id string) {
	s.mu.Lock()

	if _, ok := s.caches[id]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.caches, id)
	for i, existingID := range s.order {
		if existingID == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	s.mu.Unlock()

	s.NotifyChanged()
}

func (s *Subsystem) GetCaches() []Source {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	sources := make([]Source, 0, len(s.order))
	for _, id := range s.order {
		if source, ok := s.caches[id]; ok {
			sources = append(sources, source)
		}
	}
	return sources
}

func (s *Subsystem) Snapshots() []SourceSnapshot {
	sources := s.GetCaches()
	snapshots := make([]SourceSnapshot, 0, len(sources))
	for _, source := range sources {
		snapshots = append(snapshots, source.Snapshot())
	}
	sort.SliceStable(snapshots, func(i, j int) bool {
		return snapshots[i].Name < snapshots[j].Name
	})
	return snapshots
}

func (s *Subsystem) AddChangeListener(fn ChangeListener) UnsubscribeFunc {
	if s == nil || fn == nil {
		return func() {}
	}

	s.mu.Lock()
	s.nextListenerID++
	id := s.nextListenerID
	s.listeners[id] = fn
	s.mu.Unlock()

	return func() {
		s.removeChangeListener(id)
	}
}

func (s *Subsystem) removeChangeListener(id uint64) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.listeners, id)
}

func (s *Subsystem) NotifyChanged() {
	if s == nil {
		return
	}

	s.mu.RLock()
	listeners := make([]ChangeListener, 0, len(s.listeners))
	for _, fn := range s.listeners {
		listeners = append(listeners, fn)
	}
	s.mu.RUnlock()

	for _, fn := range listeners {
		fn()
	}
}
