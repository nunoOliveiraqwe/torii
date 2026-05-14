package service

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nunoOliveiraqwe/torii/config"
	"github.com/nunoOliveiraqwe/torii/internal/domain"
	"github.com/nunoOliveiraqwe/torii/internal/service/acme"
	"github.com/nunoOliveiraqwe/torii/internal/store"
	"go.uber.org/zap"
)

var (
	ErrAcmeAlreadyConfigured     = fmt.Errorf("ACME is already configured; reset to reconfigure")
	ErrAcmeNotConfigured         = fmt.Errorf("no ACME configuration exists")
	ErrLegoManagerNotInitialized = fmt.Errorf("ACME manager is not initialized")
	ErrEmailRequired             = fmt.Errorf("email is required")
	ErrDNSProviderRequired       = fmt.Errorf("DNS provider is required")
	ErrInvalidDNSProvider        = fmt.Errorf("invalid DNS provider")
	ErrInvalidDNSProviderCfg     = fmt.Errorf("invalid DNS provider configuration")
	ErrInvalidRenewalFmt         = fmt.Errorf("invalid renewal interval format (use Go duration, e.g. 12h, 6h30m)")
	ErrRenewalTooShort           = fmt.Errorf("renewal interval must be at least 1h")
)

type AcmeConfigResult struct {
	Email                string
	DNSProvider          string
	CADirURL             string
	RenewalCheckInterval string
	Enabled              bool
	Configured           bool
	DNSResolvers         []string
}

// AcmeCertResult is returned by ListCertificates.
type AcmeCertResult struct {
	Domain    string
	ExpiresAt time.Time
	CreatedAt time.Time
	Active    bool
}

type SaveAcmeConfigRequest struct {
	Email                string
	CADirURL             string
	RenewalCheckInterval string
	Enabled              bool
	DNSProvider          string
	CredentialMap        map[string]string
	Domains              []string
	DNSResolvers         []string
}

type AcmeRegisteredProxy struct {
	DomainSupplier func() []string
}

type AcmeService struct {
	registeredProxy []*AcmeRegisteredProxy
	store           store.AcmeStore

	legoManager atomic.Pointer[acme.LegoAcmeManager]
}

func NewAcmeService(acmeStore store.AcmeStore, conf *config.AcmeConfig) *AcmeService {
	mgr, err := getLegoManagerViaBootstrap(acmeStore, conf)
	if err != nil {
		zap.S().Errorf("Failed to initialize ACME manager: %v. HTTPS will not work properly if ACME is specified ", err)
	}
	svc := &AcmeService{
		store: acmeStore,
	}
	svc.legoManager.Store(mgr)
	return svc
}

func (s *AcmeService) RegisterProxy(p *AcmeRegisteredProxy) {
	s.registeredProxy = append(s.registeredProxy, p)
}

func (s *AcmeService) NotifyDomainsChanged() {
	currentManager := s.legoManager.Load()
	if currentManager == nil || !currentManager.IsStarted() {
		return
	}

	go func() {
		if err := currentManager.EnsureCertificates(); err != nil {
			zap.S().Errorf("acme: ensure certs after domain change: %v", err)
		}
	}()
}

func (s *AcmeService) GetConfiguration() (*domain.AcmeConfiguration, error) {
	conf, err := s.store.GetConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ACME configuration: %w", err)
	}
	if conf == nil {
		return nil, nil
	}
	return conf, nil
}

func (s *AcmeService) SaveConfiguration(req *SaveAcmeConfigRequest) error {
	existing, err := s.store.GetConfiguration()
	if err != nil {
		return fmt.Errorf("failed to check existing configuration: %w", err)
	}
	if existing != nil {
		return ErrAcmeAlreadyConfigured
	}

	if req.Email == "" {
		return ErrEmailRequired
	}
	if req.DNSProvider == "" {
		return ErrDNSProviderRequired
	}

	provider, err := acme.GetDNSProvider(req.DNSProvider)
	if err != nil {
		return ErrInvalidDNSProvider
	}

	renewalInterval := 12 * time.Hour
	if req.RenewalCheckInterval != "" {
		parsed, pErr := time.ParseDuration(req.RenewalCheckInterval)
		if pErr != nil {
			return ErrInvalidRenewalFmt
		}
		if parsed < 1*time.Hour {
			return ErrRenewalTooShort
		}
		renewalInterval = parsed
	}

	if err := provider.IsValidMap(req.CredentialMap); err != nil {
		return ErrInvalidDNSProviderCfg
	}
	sf, err := provider.Serialize(req.CredentialMap)
	if err != nil {
		return fmt.Errorf("failed to serialize DNS provider configuration: %w", err)
	}

	conf := &domain.AcmeConfiguration{
		Email:                req.Email,
		DNSProvider:          provider.Name(),
		CADirURL:             req.CADirURL,
		RenewalCheckInterval: renewalInterval,
		Enabled:              req.Enabled,
		SerializedFields:     sf,
		Domains:              req.Domains,
		DNSResolvers:         acme.NormalizeDNSResolvers(req.DNSResolvers),
	}

	if err := s.store.SaveConfiguration(conf); err != nil {
		return fmt.Errorf("failed to save ACME configuration: %w", err)
	}

	if err := s.initializeLegoAcmeManager(); err != nil {
		return fmt.Errorf("configuration saved but failed to initialize ACME manager: %w", err)
	}

	zap.S().Infow("ACME configuration saved and applied",
		"email", req.Email,
		"dnsProvider", provider.Name(),
		"enabled", req.Enabled,
	)
	return nil
}

func (s *AcmeService) ToggleEnabled(enabled bool) error {
	conf, err := s.store.GetConfiguration()
	if err != nil {
		return fmt.Errorf("failed to read ACME configuration: %w", err)
	}
	if conf == nil {
		return ErrAcmeNotConfigured
	}
	currentManager := s.legoManager.Load()
	if currentManager == nil {
		return ErrLegoManagerNotInitialized
	}

	conf.Enabled = enabled
	if err := s.store.SaveConfiguration(conf); err != nil {
		return fmt.Errorf("failed to update ACME enabled state: %w", err)
	}

	if enabled {
		currentManager.Start()
	} else {
		currentManager.Stop()
	}

	state := "disabled"
	if enabled {
		state = "enabled"
	}
	zap.S().Infof("ACME %s", state)
	return nil
}

func (s *AcmeService) ListCertificates() ([]AcmeCertResult, error) {
	certs, err := s.store.ListCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ACME certificates: %w", err)
	}

	activeDomains := make(map[string]bool)
	for _, d := range s.collectAllDomains() {
		activeDomains[strings.ToLower(d)] = true
	}

	results := make([]AcmeCertResult, 0, len(certs))
	for _, c := range certs {
		results = append(results, AcmeCertResult{
			Domain:    c.Domain,
			ExpiresAt: c.ExpiresAt,
			CreatedAt: c.CreatedAt,
			Active:    isCertDomainActive(strings.ToLower(c.Domain), activeDomains),
		})
	}
	return results, nil
}

func (s *AcmeService) ResetAll() error {
	currentManager := s.legoManager.Swap(nil)

	if currentManager != nil {
		zap.S().Info("Resetting ACME data: stopping renewal loop, revoking certificates and clearing configuration")
		currentManager.Stop()
		if err := currentManager.ResetAll(); err != nil {
			zap.S().Warnf("ACME manager reset had errors (continuing): %v", err)
		}
		zap.S().Info("ACME data reset successfully")
		return nil
	}

	zap.S().Info("No active ACME manager; resetting store only")
	if err := s.store.ResetAll(); err != nil {
		return fmt.Errorf("failed to reset ACME data: %w", err)
	}
	zap.S().Info("ACME data reset successfully")
	return nil
}

func (s *AcmeService) GetAcmeTLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: s.getCertificate,
		MinVersion:     tls.VersionTLS12,
	}
}

func (s *AcmeService) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	currentManager := s.legoManager.Load()
	if currentManager == nil {
		return nil, fmt.Errorf("ACME is not configured")
	}

	return currentManager.GetCertificate(hello)
}

func (s *AcmeService) Start() {
	currentManager := s.legoManager.Load()
	if currentManager != nil {
		currentManager.SetDomainSupplier(s.collectAllDomains)
		currentManager.Start()
	}
}

func (s *AcmeService) Stop() {
	currentManager := s.legoManager.Load()
	if currentManager != nil {
		currentManager.Stop()
	}
}

func (s *AcmeService) UpdateDomains(domains []string) error {
	conf, err := s.store.GetConfiguration()
	if err != nil {
		return fmt.Errorf("failed to read ACME configuration: %w", err)
	}
	if conf == nil {
		return ErrAcmeNotConfigured
	}

	conf.Domains = domains
	if err := s.store.SaveConfiguration(conf); err != nil {
		return fmt.Errorf("failed to update ACME domains: %w", err)
	}

	s.NotifyDomainsChanged()
	zap.S().Infow("ACME domains updated", "domains", domains)
	return nil
}

func (s *AcmeService) collectAllDomains() []string {
	var domains []string

	// Include domains stored in the ACME configuration
	conf, err := s.store.GetConfiguration()
	if err == nil && conf != nil {
		if len(conf.Domains) > 0 {
			domains = append(domains, conf.Domains...)
			return domains
		}
	}

	for _, p := range s.registeredProxy {
		if p.DomainSupplier != nil {
			domains = append(domains, p.DomainSupplier()...)
		}
	}
	return domains
}

// getLegoManagerViaBootstrap initializes the ACME manager using the bootstrap method, which reads configuration from the store and sets up the manager accordingly.
// This is used during service initialization to ensure that if there is existing configuration, it is loaded properly. I can return nil
// if no configuration exists, and the caller can handle that case by leaving the manager uninitialized until configuration is saved.
func getLegoManagerViaBootstrap(acmeStore store.AcmeStore,
	conf *config.AcmeConfig) (*acme.LegoAcmeManager, error) {
	acmeMgr, err := acme.Bootstrap(acmeStore, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap ACME: %w", err)
	}
	return acmeMgr, nil
}

// initializeLegoAcmeManager initializes the ACME manager based on the current configuration in the store.
// It checks if the manager is already initialized to prevent reinitialization, then reads the configuration, creates a new manager,
// and sets it up with the domain supplier. This is called after saving new configuration to apply changes.
// this is used when the config is supplied via API and we want to initialize the manager immediately after saving the config.
func (s *AcmeService) initializeLegoAcmeManager() error {
	if s.legoManager.Load() != nil {
		return fmt.Errorf("ACME manager is already initialized")
	}

	conf, err := s.store.GetConfiguration()
	if err != nil {
		return fmt.Errorf("failed to read ACME configuration: %w", err)
	}

	newMgr, err := acme.NewLegoAcmeManager(conf, s.store)
	if err != nil {
		return fmt.Errorf("failed to create ACME manager: %w", err)
	}
	newMgr.SetDomainSupplier(s.collectAllDomains)

	if !s.legoManager.CompareAndSwap(nil, newMgr) {
		return fmt.Errorf("ACME manager is already initialized")
	}
	if conf.Enabled {
		newMgr.Start()
	}
	return nil
}

func (s *AcmeService) reloadLegoAcmeManager() error {
	conf, err := s.store.GetConfiguration()
	if err != nil {
		return fmt.Errorf("failed to read ACME configuration: %w", err)
	}

	var newManager *acme.LegoAcmeManager
	if conf != nil && conf.Enabled {
		newManager, err = acme.NewLegoAcmeManager(conf, s.store)
		if err != nil {
			return fmt.Errorf("failed to create ACME manager: %w", err)
		}
		newManager.SetDomainSupplier(s.collectAllDomains)
	}

	oldManager := s.legoManager.Swap(newManager)
	if oldManager != nil {
		oldManager.Stop()
	}
	if newManager != nil {
		newManager.Start()
	}
	return nil
}

func isCertDomainActive(certDomain string, activeDomains map[string]bool) bool {
	if activeDomains[certDomain] {
		return true
	}

	if strings.HasPrefix(certDomain, "*.") {
		parent := certDomain[2:] // e.g. "example.com"
		for host := range activeDomains {
			if idx := strings.Index(host, "."); idx > 0 && host[idx+1:] == parent {
				return true
			}
		}
	} else {
		if idx := strings.Index(certDomain, "."); idx > 0 {
			wildcard := "*." + certDomain[idx+1:]
			if activeDomains[wildcard] {
				return true
			}
		}
	}

	return false
}
