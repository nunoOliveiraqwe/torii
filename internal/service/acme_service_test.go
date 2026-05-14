package service

import (
	"testing"

	"github.com/nunoOliveiraqwe/torii/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAcmeStore struct {
	conf *domain.AcmeConfiguration
}

func (f *fakeAcmeStore) GetConfiguration() (*domain.AcmeConfiguration, error) {
	return f.conf, nil
}

func (f *fakeAcmeStore) SaveConfiguration(conf *domain.AcmeConfiguration) error {
	f.conf = conf
	return nil
}

func (f *fakeAcmeStore) GetAccount(string) (*domain.AcmeAccount, error) {
	return nil, nil
}

func (f *fakeAcmeStore) SaveAccount(*domain.AcmeAccount) error {
	return nil
}

func (f *fakeAcmeStore) GetCertificate(string) (*domain.AcmeCertificate, error) {
	return nil, nil
}

func (f *fakeAcmeStore) SaveCertificate(*domain.AcmeCertificate) error {
	return nil
}

func (f *fakeAcmeStore) ListCertificates() ([]*domain.AcmeCertificate, error) {
	return nil, nil
}

func (f *fakeAcmeStore) ResetAll() error {
	f.conf = nil
	return nil
}

func TestCollectAllDomains_ConfiguredDomainsTakePrecedence(t *testing.T) {
	svc := &AcmeService{
		store: &fakeAcmeStore{
			conf: &domain.AcmeConfiguration{
				Domains: []string{"*.example.com"},
			},
		},
	}
	svc.RegisterProxy(&AcmeRegisteredProxy{
		DomainSupplier: func() []string {
			t.Fatal("route domain supplier should not be called when ACME domains are configured")
			return nil
		},
	})

	got := svc.collectAllDomains()

	assert.Equal(t, []string{"*.example.com"}, got)
}

func TestCollectAllDomains_UsesRouteDomainsWhenNoConfiguredDomains(t *testing.T) {
	svc := &AcmeService{
		store: &fakeAcmeStore{conf: &domain.AcmeConfiguration{}},
	}
	svc.RegisterProxy(&AcmeRegisteredProxy{
		DomainSupplier: func() []string {
			return []string{"app.example.com", "api.example.com"}
		},
	})

	got := svc.collectAllDomains()

	require.Len(t, got, 2)
	assert.ElementsMatch(t, []string{"app.example.com", "api.example.com"}, got)
}
