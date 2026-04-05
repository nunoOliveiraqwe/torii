package app

import "github.com/nunoOliveiraqwe/torii/internal/store"

type ServiceStore struct {
	apiKeyService              *ApiKeyService
	userService                *UserService
	systemConfigurationService *SystemConfigurationService
	acmeStore                  store.AcmeStore
}

func NewServiceStore(ds *DataStore) *ServiceStore {
	return &ServiceStore{
		apiKeyService:              NewApiKeyService(ds.ApiKeyStore),
		userService:                NewUserService(ds),
		systemConfigurationService: NewSystemConfigurationService(ds),
		acmeStore:                  ds.AcmeStore,
	}
}

func (s *ServiceStore) GetUserService() *UserService {
	return s.userService
}

func (s *ServiceStore) GetSystemConfigurationService() *SystemConfigurationService {
	return s.systemConfigurationService
}

func (s *ServiceStore) GetAcmeStore() store.AcmeStore {
	return s.acmeStore
}

func (s *ServiceStore) GetApiKeyService() *ApiKeyService {
	return s.apiKeyService
}
