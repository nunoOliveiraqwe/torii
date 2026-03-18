package model

const system_config_id = 1 // there is only one system configuration
type SystemConfiguration struct {
	ID                        int
	IsFirstTimeSetupConcluded bool
}

type SystemConfigurationDal interface {
	GetSystemConfiguration() (*SystemConfiguration, error)
	UpdateSystemConfiguration(config *SystemConfiguration) error
}
