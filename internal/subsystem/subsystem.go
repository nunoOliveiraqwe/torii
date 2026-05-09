package subsystem

import (
	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/subsystem/activity"
	"go.uber.org/zap"
)

type Subsystem interface {
	Initialize() error
	Shutdown() error
}

type Manager struct {
	eventBus          *bus.EventBus
	activitySubsystem *activity.Subsystem
}

func NewSubsystemManager(eventBus *bus.EventBus) *Manager {
	zap.S().Debug("Creating SubsystemManager")
	activitySubsystem := activity.NewDefaultActivitySubsystem(eventBus)
	return &Manager{
		eventBus:          eventBus,
		activitySubsystem: activitySubsystem,
	}
}

func (m *Manager) Initialize() error {
	zap.S().Debug("Initializing SubsystemManager")
	if err := m.activitySubsystem.Initialize(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Shutdown() error {
	zap.S().Debug("Shutting down SubsystemManager")
	if err := m.activitySubsystem.Shutdown(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) GetActivitySubsystem() *activity.Subsystem {
	return m.activitySubsystem
}
