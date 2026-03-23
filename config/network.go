package config

import (
	"time"

	"github.com/nunoOliveiraqwe/micro-proxy/middleware"
)

type IpFlag byte

const (
	Ipv4Flag IpFlag = 1 << iota
	Ipv6Flag IpFlag = 2
	BothFlag IpFlag = Ipv4Flag | Ipv6Flag
)

type ACMEConfig struct {
	Email      string `yaml:"email"`
	Cache      string `yaml:"cache"`
	OpenPort80 bool   `yaml:"open-port-80"`
}

type NetworkConfig struct {
	HTTPListeners []HTTPListener `yaml:"http"`
	ACMEConfig    *ACMEConfig    `yaml:"acme"`
	TCPListeners  []TCPListener  `yaml:"tcp"`
}

type HTTPListener struct {
	Port              int           `yaml:"port"`
	Bind              IpFlag        `yaml:"bind"`
	Interface         string        `yaml:"interface"`
	TLS               *TLSConfig    `yaml:"tls"`
	ReadTimeout       time.Duration `yaml:"read-timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read-header-timeout"`
	WriteTimeout      time.Duration `yaml:"write-timeout"`
	IdleTimeout       time.Duration `yaml:"idle-timeout"`
	Routes            []Route       `yaml:"routes"`
	Default           *RouteTarget  `yaml:"default"`
}

type TCPListener struct {
	Port        int                 `yaml:"port"`
	Bind        string              `yaml:"bind"`
	Interface   string              `yaml:"interface"`
	Backend     string              `yaml:"backend"`
	Middlewares []middleware.Config `yaml:"middlewares"`
}

type Route struct {
	Host        string              `yaml:"host"`
	Backend     string              `yaml:"backend"`
	Middlewares []middleware.Config `yaml:"middlewares"`
}

// RouteTarget is the shared backend+middlewares block
// used by both Route and Listener.Default
type RouteTarget struct {
	Backend     string              `yaml:"backend"`
	Middlewares []middleware.Config `yaml:"middlewares"`
}

type TLSConfig struct {
	UseAcme bool   `yaml:"use-acme"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}
