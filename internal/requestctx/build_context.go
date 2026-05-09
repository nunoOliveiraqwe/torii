package requestctx

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"github.com/nunoOliveiraqwe/torii/metrics"
)

type BuildContext struct {
	RuntimeContext      context.Context
	MetricsManager      *metrics.ConnectionMetricsManager
	CacheInsights       *util.CacheInsightManager
	EventBus            bus.Bus
	Port                int
	ServerID            string
	Host                string
	Path                string
	OverrideMetricsName string
}

func NewBuildContext(metricsManager *metrics.ConnectionMetricsManager, cacheInsights *util.CacheInsightManager,
	eventBus bus.Bus, port int, serverID, host, path, overrideMetricsName string) BuildContext {
	return BuildContext{
		RuntimeContext:      context.Background(),
		MetricsManager:      metricsManager,
		CacheInsights:       cacheInsights,
		EventBus:            eventBus,
		Port:                port,
		ServerID:            serverID,
		Host:                host,
		Path:                path,
		OverrideMetricsName: overrideMetricsName,
	}
}

func (c BuildContext) WithRuntimeContext(ctx context.Context) BuildContext {
	c.RuntimeContext = ctx
	return c
}

func (c BuildContext) Context() context.Context {
	if c.RuntimeContext == nil {
		return context.Background()
	}
	return c.RuntimeContext
}

func (c BuildContext) WithPort(port int) BuildContext {
	c.Port = port
	return c
}

func (c BuildContext) WithServerID(serverID string) BuildContext {
	c.ServerID = serverID
	return c
}

func (c BuildContext) WithHost(host string) BuildContext {
	c.Host = host
	return c
}

func (c BuildContext) WithPath(path string) BuildContext {
	c.Path = path
	return c
}

func (c BuildContext) WithOverrideMetricsName(name string) BuildContext {
	c.OverrideMetricsName = name
	return c
}

func (c BuildContext) PortString() string {
	if c.Port == 0 {
		return ""
	}
	return strconv.Itoa(c.Port)
}

func (c BuildContext) ConnectionName() string {
	if c.Port != 0 {
		return c.BuildConnectionName("conn")
	}
	if name := ConnectionNameFromServerID(c.ServerID); name != "" {
		return name
	}
	return "conn-unknown"
}

func (c BuildContext) BuildConnectionName(prefix string) string {
	return BuildConnectionName(prefix, c.PortString(), c.Host, c.Path)
}

func BuildConnectionName(prefix, port, host, path string) string {
	base := fmt.Sprintf("%s-port-%s", prefix, port)
	if host != "" {
		base = fmt.Sprintf("%s-host-%s", base, host)
	}
	if path != "" {
		base = fmt.Sprintf("%s-path-%s", base, path)
	}
	return base
}

func ConnectionNameFromServerID(serverID string) string {
	if strings.HasPrefix(serverID, "http-") {
		port := strings.TrimPrefix(serverID, "http-")
		if port != "" {
			return BuildConnectionName("conn", port, "", "")
		}
	}
	return ""
}
