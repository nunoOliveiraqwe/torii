package subsystem

import (
	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/subsystem/activity"
	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"go.uber.org/zap"
)

type Subsystem interface {
	Initialize() error
	Shutdown() error
}

type Manager struct {
	eventBus          *bus.EventBus
	activitySubsystem *activity.Subsystem
	cacheSubsystem    *cacheSub.Subsystem
}

func NewSubsystemManager(eventBus *bus.EventBus, cacheSubsystem *cacheSub.Subsystem) *Manager {
	zap.S().Debug("Creating SubsystemManager")
	if cacheSubsystem == nil {
		cacheSubsystem = cacheSub.NewSubsystem()
	}
	activitySubsystem := activity.NewDefaultActivitySubsystem(eventBus)
	return &Manager{
		eventBus:          eventBus,
		activitySubsystem: activitySubsystem,
		cacheSubsystem:    cacheSubsystem,
	}
}

func (m *Manager) Initialize() error {
	zap.S().Debug("Initializing SubsystemManager")
	if err := m.cacheSubsystem.Initialize(); err != nil {
		return err
	}
	if err := m.activitySubsystem.Initialize(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Shutdown() error {
	zap.S().Info("Shutting down SubsystemManager")
	if err := m.activitySubsystem.Shutdown(); err != nil {
		return err
	}
	if err := m.cacheSubsystem.Shutdown(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) GetActivitySubsystem() *activity.Subsystem {
	return m.activitySubsystem
}

func (m *Manager) GetCacheSubsystem() *cacheSub.Subsystem {
	return m.cacheSubsystem
}
