package metrics

import (
	"context"
)

import (
	"sync"

	"go.uber.org/zap"
)

type ConnectionMetricsManager struct {
	connectionMetricsMap map[string]*ConnectionMetric
	metricsChan          chan *RequestMetric
	numberOfListeners    int
	context              context.Context
}

type ConnectionMetric struct {
	accumulatedMetrics *Metric
	connectionName     string
	metricsLock        sync.RWMutex
}

type MetricsReportFunc func(reqMetric *RequestMetric)

const globalMetricsConName = "global"

func NewGlobalMetricsHandler(numberOfListeners int, ctx context.Context) *ConnectionMetricsManager {
	zap.S().Debug("Creating connection metrics handler")
	h := ConnectionMetricsManager{
		connectionMetricsMap: make(map[string]*ConnectionMetric),
		metricsChan:          make(chan *RequestMetric),
		numberOfListeners:    numberOfListeners,
		context:              ctx,
	}
	zap.S().Info("Creating a new global connection metric")
	h.NewConnectionMetric(globalMetricsConName)
	return &h
}

func (h *ConnectionMetricsManager) addConnectionMetric(c *ConnectionMetric) {
	zap.S().Debugf("Adding connection metric for connection %s", c.connectionName)
	h.connectionMetricsMap[c.connectionName] = c
}

func (h *ConnectionMetricsManager) startCollectingMetrics() {
	waitG := sync.WaitGroup{}
	waitG.Add(h.numberOfListeners)
	for i := 0; i < h.numberOfListeners; i++ {
		go func() {
			h.collectGlobalMetrics()
			waitG.Done()
		}()
	}
	waitG.Wait()
	close(h.metricsChan)
}

func (h *ConnectionMetricsManager) collectGlobalMetrics() {
	for {
		select {
		case metric, ok := <-h.metricsChan:
			if !ok {
				return
			}
			h.updateConnectionMetrics(metric)
		case <-h.context.Done():
			return
		}
	}
}

func (h *ConnectionMetricsManager) updateConnectionMetrics(metric *RequestMetric) {
	zap.S().Infof("Updating connection metric for connection %s", metric.connectionName)
	conMetrics, ok := h.connectionMetricsMap[metric.connectionName]
	if !ok {
		zap.S().Warnf("Connection metric for connection %s not found", metric.connectionName)
		return
	}
	conMetrics.metricsLock.Lock()
	defer conMetrics.metricsLock.Unlock()
	conMetrics.accumulatedMetrics.AddRequestMetric(metric)
	if metric.connectionName != globalMetricsConName {
		globalConMetrics, ok2 := h.connectionMetricsMap[globalMetricsConName]
		if !ok2 {
			zap.S().Errorf("no global connection metrics found")
			return
		}
		globalConMetrics.metricsLock.Lock()
		defer globalConMetrics.metricsLock.Unlock()
		globalConMetrics.accumulatedMetrics.AddRequestMetric(metric)
	}
}

func (h *ConnectionMetricsManager) NewConnectionMetric(connectionName string) MetricsReportFunc {
	zap.S().Debugf("Creating a new connection metric for connection %s", connectionName)
	connMetric := &ConnectionMetric{
		accumulatedMetrics: &Metric{},
		metricsLock:        sync.RWMutex{},
		connectionName:     connectionName,
	}
	h.addConnectionMetric(connMetric)
	return func(metric *RequestMetric) {
		metric.connectionName = connectionName
		h.metricsChan <- metric
	}
}

func (h *ConnectionMetricsManager) GetMetricForConnection(connectionName string) *Metric {
	zap.S().Debugf("Getting connection metrics for connection %s", connectionName)
	conMetrics, ok := h.connectionMetricsMap[connectionName]
	if !ok {
		zap.S().Infof("Connection metric for connection %s not found", connectionName)
		return nil
	}
	conMetrics.metricsLock.RLock()
	defer conMetrics.metricsLock.RUnlock()
	return conMetrics.accumulatedMetrics.Copy()
}

func (h *ConnectionMetricsManager) StartCollectingMetrics() {
	go h.startCollectingMetrics()
}
