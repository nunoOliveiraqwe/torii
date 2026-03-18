package manager

import (
	"context"
	"fmt"

	"github.com/nunoOliveiraqwe/micro-proxy/configuration"
	"github.com/nunoOliveiraqwe/micro-proxy/internal/sqlite"
	"github.com/nunoOliveiraqwe/micro-proxy/metrics"
	"github.com/nunoOliveiraqwe/micro-proxy/proxy"
	"github.com/nunoOliveiraqwe/micro-proxy/server/session"
	"go.uber.org/zap"
)

type SystemManager interface {
	Start() error
	Stop() error
	SessionRegistry() *session.SessionRegistry
	GetSystemHandler() SystemInterfaceHandler
}

type SystemInterfaceHandler interface {
	IsFirstTimeSetupCompleted() bool
}

type systemManager struct {
	micro          *proxy.MicroProxy
	db             *sqlite.DB
	sessions       *session.SessionRegistry
	MetricsManager *metrics.ConnectionMetricsManager
}

func (sm *systemManager) GetSystemHandler() SystemInterfaceHandler {
	return sm
}

func NewSystemManager(conf configuration.ApplicationConfiguration) (SystemManager, error) {
	zap.S().Info("Initializing system manager")

	metricsManager := metrics.NewGlobalMetricsHandler(2, context.Background())
	metrics.RegisterGlobalMetricsManager(metricsManager)

	m, err := proxy.NewMicroProxy(conf.NetConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create micro proxy: %w", err)
	}

	db := sqlite.NewDB("micro-proxy.db")
	if err := db.Open(); err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	sessions := session.NewSessionRegistry(db, conf.Session)

	return &systemManager{
		micro:          m,
		db:             db,
		sessions:       sessions,
		MetricsManager: metricsManager,
	}, nil
}

func (sm *systemManager) SessionRegistry() *session.SessionRegistry {
	return sm.sessions
}

func (sm *systemManager) Start() error {
	zap.S().Info("Starting system manager")
	if err := sm.micro.Start(); err != nil {
		return fmt.Errorf("failed to start micro proxy: %w", err)
	}
	sm.MetricsManager.StartCollectingMetrics()
	zap.S().Info("System manager started successfully")
	return nil
}

func (sm *systemManager) Stop() error {
	zap.S().Info("Stopping system manager")
	if err := sm.micro.Stop(); err != nil {
		return fmt.Errorf("failed to stop micro proxy: %w", err)
	}
	zap.S().Info("System manager stopped successfully")
	return nil
}

func (sm *systemManager) GetRunningProxies() {
	zap.S().Infof("Getting running proxies")
}

func (sm *systemManager) IsFirstTimeSetupCompleted() bool {
	return true
}
