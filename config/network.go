package config

import (
	"time"

	"github.com/nunoOliveiraqwe/torii/middleware"
)

type IpFlag byte

const (
	Ipv4Flag IpFlag = 1 << iota
	Ipv6Flag IpFlag = 2
	BothFlag IpFlag = Ipv4Flag | Ipv6Flag
)

type NetworkConfig struct {
	HTTPListeners []HTTPListener `yaml:"http" json:"http,omitempty"`
	TCPListeners  []TCPListener  `yaml:"tcp" json:"tcp,omitempty"`
	Global        *GlobalConfig  `yaml:"global" json:"global,omitempty"`
}

type HTTPListener struct {
	Port              int           `yaml:"port" json:"port"`
	Bind              IpFlag        `yaml:"bind" json:"bind"`
	Interface         string        `yaml:"interface" json:"interface,omitempty"`
	TLS               *TLSConfig    `yaml:"tls" json:"tls,omitempty"`
	ReadTimeout       time.Duration `yaml:"read-timeout" json:"read_timeout,omitempty"`
	ReadHeaderTimeout time.Duration `yaml:"read-header-timeout" json:"read_header_timeout,omitempty"`
	WriteTimeout      time.Duration `yaml:"write-timeout" json:"write_timeout,omitempty"`
	IdleTimeout       time.Duration `yaml:"idle-timeout" json:"idle_timeout,omitempty"`
	Routes            []Route       `yaml:"routes" json:"routes,omitempty"`
	Default           *RouteTarget  `yaml:"default" json:"default,omitempty"`
}

type GlobalConfig struct {
	Middlewares []middleware.Config `yaml:"middlewares" json:"middlewares,omitempty"`
}

type TCPListener struct {
	Port        int                 `yaml:"port" json:"port"`
	Bind        string              `yaml:"bind" json:"bind,omitempty"`
	Interface   string              `yaml:"interface" json:"interface,omitempty"`
	Backend     string              `yaml:"backend" json:"backend"`
	Middlewares []middleware.Config `yaml:"middlewares" json:"middlewares,omitempty"`
}

type Route struct {
	Host   string      `yaml:"host" json:"host"`
	Target RouteTarget `yaml:"target" json:"target"`
}

type PathRule struct {
	Pattern     string              `yaml:"pattern" json:"pattern"`
	Backend     string              `yaml:"backend" json:"backend,omitempty"`
	DropQuery   *bool               `yaml:"drop-query" json:"drop_query,omitempty"`
	Middlewares []middleware.Config `yaml:"middlewares" json:"middlewares,omitempty"`
}

type RouteTarget struct {
	Backend     string              `yaml:"backend" json:"backend"`
	Middlewares []middleware.Config `yaml:"middlewares" json:"middlewares,omitempty"`
	Paths       []PathRule          `yaml:"paths" json:"paths,omitempty"`
}

type TLSConfig struct {
	UseAcme bool   `yaml:"use-acme" json:"use_acme"`
	Cert    string `yaml:"cert" json:"cert,omitempty"`
	Key     string `yaml:"key" json:"key,omitempty"`
}
