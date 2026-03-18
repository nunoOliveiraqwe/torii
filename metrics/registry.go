package metrics

var (
	globalManager *ConnectionMetricsManager
)

func RegisterGlobalMetricsManager(m *ConnectionMetricsManager) {
	globalManager = m
}

func GlobalMetricsManager() *ConnectionMetricsManager {
	return globalManager
}
