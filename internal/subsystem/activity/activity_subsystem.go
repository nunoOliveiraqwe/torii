package activity

import (
	"errors"
	"sync/atomic"

	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"go.uber.org/zap"
)

const (
	DefaultBlockLogCapacity   = 1000
	DefaultRequestLogCapacity = 1000
	DefaultErrorLogCapacity   = 1000
)

type Subsystem struct {
	IsStarted       atomic.Bool
	BlockLog        *util.RingBuffer[BlockLogEntry]
	RequestLog      *util.RingBuffer[RequestLogEntry]
	requestLogUnsub bus.UnsubscribeFunc
	ErrorLog        *util.RingBuffer[ErrorLogEntry]
	eventBus        *bus.EventBus
}

func NewActivitySubsystem(blockLogCapacity, requestLogCapacity, errorLogCapacity int, eventBus *bus.EventBus) *Subsystem {
	return &Subsystem{
		BlockLog:   NewBlockLog(resolveCapacity(blockLogCapacity, DefaultBlockLogCapacity)),
		RequestLog: NewRequestLog(resolveCapacity(requestLogCapacity, DefaultRequestLogCapacity)),
		ErrorLog:   NewErrorLog(resolveCapacity(errorLogCapacity, DefaultErrorLogCapacity)),
		eventBus:   eventBus,
	}
}

func resolveCapacity(capacity, defaultCapacity int) int {
	if capacity <= 0 {
		return defaultCapacity
	}
	return capacity
}

func NewDefaultActivitySubsystem(eventBus *bus.EventBus) *Subsystem {
	return NewActivitySubsystem(DefaultBlockLogCapacity, DefaultRequestLogCapacity, DefaultErrorLogCapacity, eventBus)
}

func (a *Subsystem) Initialize() error {
	if a.eventBus == nil {
		return errors.New("activity subsystem requires an event bus")
	}
	if a.IsStarted.Load() {
		return errors.New("activity subsystem already initialized")
	}

	zap.S().Debug("Initializing activity subsystem by subscribing to event bus")
	usub, err := bus.SubscribeTyped[bus.RequestProcessed](a.eventBus, bus.TopicRequestProcessed, a.listenForCompletedRequests)
	if err != nil {
		zap.S().Errorf("Cannot initialize activity subsystem for listening to completed requests: %v", err)
		return err
	}
	a.requestLogUnsub = usub
	a.IsStarted.Store(true)
	return nil
}

func (a *Subsystem) Shutdown() error {
	if !a.IsStarted.Load() {
		return errors.New("activity subsystem is not started")
	}
	if a.requestLogUnsub != nil {
		a.requestLogUnsub()
		a.requestLogUnsub = nil
	}
	a.IsStarted.Store(false)
	return nil
}

func (a *Subsystem) GetLogCapacities() (errorCapacity, blockCapacity, requestCapacity int) {
	if !a.IsStarted.Load() {
		zap.S().Warnf("Activity subsystem is not started, cannot retrieve log capacities")
		return 0, 0, 0
	}
	return a.ErrorLog.Capacity(), a.BlockLog.Capacity(), a.RequestLog.Capacity()
}

func (a *Subsystem) listenForCompletedRequests(evt *bus.Event, r bus.RequestProcessed) {
	if evt == nil || evt.Request == nil {
		return
	}

	statusCode := r.StatusCode
	if statusCode == 0 {
		statusCode = evt.Request.StatusCode
	}
	connectionName := r.ConnectionName
	if connectionName == "" {
		connectionName = evt.Source
	}

	if r.Blocked != nil {
		a.BlockLog.Add(BlockLogEntry{
			RemoteAddress:      evt.Request.ClientIP,
			Host:               evt.Request.Host,
			Timestamp:          evt.At,
			Method:             evt.Request.Method,
			Path:               evt.Request.Path,
			Status:             statusCode,
			ConnectionName:     connectionName,
			BlockingMiddleware: r.Blocked.Middleware,
			BlockReason:        r.Blocked.Reason,
			LatencyMs:          r.ProcessingTimeMs,
		})
	} else if statusCode >= 500 {
		a.ErrorLog.Add(ErrorLogEntry{
			Timestamp:      evt.At,
			ConnectionName: connectionName,
			RemoteAddress:  evt.Request.ClientIP,
			Host:           evt.Request.Host,
			StatusCode:     statusCode,
			Method:         evt.Request.Method,
			Path:           evt.Request.Path,
			LatencyMs:      r.ProcessingTimeMs,
		})
	} else {
		a.RequestLog.Add(RequestLogEntry{
			RemoteAddress:  evt.Request.ClientIP,
			Host:           evt.Request.Host,
			Country:        r.CountryCode,
			Timestamp:      evt.At,
			ConnectionName: connectionName,
			StatusCode:     statusCode,
			Method:         evt.Request.Method,
			Path:           evt.Request.Path,
			LatencyMs:      r.ProcessingTimeMs,
			BytesSent:      r.BytesSent,
			BytesReceived:  r.BytesReceived,
		})
	}
}
