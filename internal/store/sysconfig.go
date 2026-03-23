package store

import (
	"github.com/nunoOliveiraqwe/micro-proxy/internal/domain"
)

type SystemConfigStore interface {
	GetSystemConfiguration() (*domain.SystemConfiguration, error)
	UpdateSystemConfiguration(config *domain.SystemConfiguration) error
}
