package metrics

import "fmt"

type Metric struct {
	ConnectionName    string `json:"connection_name"`
	RequestCount      int64  `json:"request_count"`
	ErrorCount        int64  `json:"error_count"`
	TotalLatencyMs    int64  `json:"total_latency_ms"`
	AvgResponseTimeMs int64  `json:"avg_response_time_ms"`
	UpstreamTimeouts  int64  `json:"upstream_timeouts"`
	BytesSent         int64  `json:"bytes_sent"`
	BytesReceived     int64  `json:"bytes_received"`
	ActiveConnections int64  `json:"active_connections"`
	Request2xxCount   int64  `json:"request_2xx_count"`
	Request3xxCount   int64  `json:"request_3xx_count"`
	Request4xxCount   int64  `json:"request_4xx_count"`
	Request5xxCount   int64  `json:"request_5xx_count"`
	CacheHits         int64  `json:"cache_hits"`
	CacheMisses       int64  `json:"cache_misses"`
}

type RequestMetric struct {
	connectionName string
	LatencyMs      int64
	IsTimedOut     bool
	BytesSent      int64
	BytesReceived  int64
	Is2xxResponse  bool
	Is3xxResponse  bool
	Is4xxResponse  bool
	Is5xxResponse  bool
}

func (m *Metric) AddRequestMetric(metric *RequestMetric) {
	m.RequestCount++
	m.TotalLatencyMs += metric.LatencyMs
	m.BytesSent += metric.BytesSent
	m.BytesReceived += metric.BytesReceived
	if metric.Is2xxResponse {
		m.Request2xxCount++
	} else if metric.Is3xxResponse {
		m.Request3xxCount++
	} else if metric.Is4xxResponse {
		m.Request4xxCount++
	} else if metric.Is5xxResponse {
		m.Request5xxCount++
		m.ErrorCount++
	}
	if metric.IsTimedOut {
		m.UpstreamTimeouts++
	}
	m.AvgResponseTimeMs = m.TotalLatencyMs / m.RequestCount
}

func (m *Metric) Add(other *Metric) {
	m.RequestCount += other.RequestCount
	m.ErrorCount += other.ErrorCount
	m.TotalLatencyMs += other.TotalLatencyMs
	m.AvgResponseTimeMs += other.AvgResponseTimeMs
	m.UpstreamTimeouts += other.UpstreamTimeouts
	m.BytesSent += other.BytesSent
	m.BytesReceived += other.BytesReceived
	m.ActiveConnections += other.ActiveConnections
	m.Request2xxCount += other.Request2xxCount
	m.Request3xxCount += other.Request3xxCount
	m.Request4xxCount += other.Request4xxCount
	m.Request5xxCount += other.Request5xxCount
	m.CacheHits += other.CacheHits
	m.CacheMisses += other.CacheMisses
}

func (m *Metric) Copy() *Metric {
	return &Metric{
		RequestCount:      m.RequestCount,
		ErrorCount:        m.ErrorCount,
		TotalLatencyMs:    m.TotalLatencyMs,
		AvgResponseTimeMs: m.AvgResponseTimeMs,
		UpstreamTimeouts:  m.UpstreamTimeouts,
		BytesSent:         m.BytesSent,
		BytesReceived:     m.BytesReceived,
		ActiveConnections: m.ActiveConnections,
		Request2xxCount:   m.Request2xxCount,
		Request3xxCount:   m.Request3xxCount,
		Request4xxCount:   m.Request4xxCount,
		Request5xxCount:   m.Request5xxCount,
		CacheHits:         m.CacheHits,
		CacheMisses:       m.CacheMisses,
	}
}

func (m *Metric) Reset() {
	*m = Metric{}
}

func (m *Metric) String() string {
	return fmt.Sprintf("Metric{ConnectionName: %s, RequestCount: %d,"+
		" ErrorCount: %d, TotalLatencyMs: %d, AvgResponseTimeMs: %d, UpstreamTimeouts: %d,"+
		" BytesSent: %d, BytesReceived: %d, ActiveConnections: %d, Request2xxCount: %d, "+
		"Request3xxCount: %d, Request4xxCount: %d, Request5xxCount: %d, CacheHits: %d,"+
		" CacheMisses: %d",
		m.ConnectionName,
		m.RequestCount,
		m.ErrorCount,
		m.TotalLatencyMs,
		m.AvgResponseTimeMs,
		m.UpstreamTimeouts,
		m.BytesSent,
		m.BytesReceived,
		m.ActiveConnections,
		m.Request2xxCount,
		m.Request3xxCount,
		m.Request4xxCount,
		m.Request5xxCount,
		m.CacheHits,
		m.CacheMisses,
	)
}

func ProxyMetricsName(iface string, port string) string {
	if iface == "" {
		iface = "any"
	}
	return fmt.Sprintf("%s-%s", iface, port)
}
