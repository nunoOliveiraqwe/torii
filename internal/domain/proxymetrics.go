package domain

import (
	"time"
)

const globalProxyMetricsID = 1 // there is only one global proxy metrics row

type GlobalProxyMetrics struct {
	ID                int
	RequestCount      int64
	ErrorCount        int64
	TotalLatencyMs    int64
	AvgResponseTimeMs int64
	UpstreamTimeouts  int64
	BytesSent         int64
	BytesReceived     int64
	ActiveConnections int64
	Request2xxCount   int64
	Request3xxCount   int64
	Request4xxCount   int64
	Request5xxCount   int64
	CacheHits         int64
	CacheMisses       int64
	SslHandshakes     int64
	BackendErrors     int64
	UpdatedAt         time.Time
}
