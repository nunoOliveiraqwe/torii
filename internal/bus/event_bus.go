package bus

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

type UnsubscribeFunc func()

type ListenerCallback func(*Event)

type Bus interface {
	Start()
	Stop()
	Publish(*Event) bool
	Subscribe(topic Topic, handler ListenerCallback) (UnsubscribeFunc, error)
}

const eventQueueSize = 1000
const handlerQueueSize = 5000
const numberOfWorkers = 10

type EventBus struct {
	ebMu          sync.Mutex
	started       atomic.Bool
	busChannel    chan *Event
	handlerQueue  chan handlerJob
	closeFunc     context.CancelFunc
	ctx           context.Context
	clientCounter atomic.Uint32
	listenerLock  sync.RWMutex
	listeners     map[Topic]map[uint32]ListenerCallback
}

func NewEventBus() *EventBus {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &EventBus{
		ebMu:         sync.Mutex{},
		listeners:    make(map[Topic]map[uint32]ListenerCallback),
		listenerLock: sync.RWMutex{},
		busChannel:   make(chan *Event, eventQueueSize),
		handlerQueue: make(chan handlerJob, handlerQueueSize),
		started:      atomic.Bool{},
		ctx:          ctx,
		closeFunc:    cancelFunc,
	}
}

func (eb *EventBus) Start() {
	eb.ebMu.Lock()
	defer eb.ebMu.Unlock()

	if !eb.started.CompareAndSwap(false, true) {
		zap.S().Warnf("Superfluous call to start event bus, already started") //annoying, just like header re-write from http package
		return
	}

	for i := 0; i < numberOfWorkers; i++ {
		go dispatchMessages(eb)
		go handleMessages(eb)
	}
}

func (eb *EventBus) Stop() {
	eb.ebMu.Lock()
	defer eb.ebMu.Unlock()

	if !eb.started.CompareAndSwap(true, false) {
		zap.S().Warnf("Superfluous call to stop event bus, already stopped")
		return
	}

	eb.closeFunc()
}

func (eb *EventBus) Publish(evt *Event) bool {
	if !eb.started.Load() {
		zap.S().Warnf("Attempted to publish event while bus is not started, dropping event: %v", evt)
		return false
	}
	select {
	case eb.busChannel <- evt:
		zap.S().Debugf("Published event: %v", evt)
	default:
		zap.S().Warnf("Event bus channel is full, dropping event: %v", evt)
		return false
	}
	return true
}

func (eb *EventBus) Subscribe(topic Topic, handler ListenerCallback) (UnsubscribeFunc, error) {
	if !eb.started.Load() {
		zap.S().Warnf("Attempted to subscribe event while bus is not started")
		return nil, fmt.Errorf("event bus is not started")
	}

	if topic == "" {
		return nil, errors.New("topic cannot be empty")
	} else if handler == nil {
		return nil, errors.New("handler cannot be nil")
	}

	eb.listenerLock.Lock()
	defer eb.listenerLock.Unlock()

	subscriptionMap := eb.listeners[topic]
	if subscriptionMap == nil {
		subscriptionMap = make(map[uint32]ListenerCallback)
		eb.listeners[topic] = subscriptionMap
	}

	clientID := eb.clientCounter.Add(1)
	subscriptionMap[clientID] = handler

	return func() {
		eb.listenerLock.Lock()
		defer eb.listenerLock.Unlock()
		delete(subscriptionMap, clientID)
		if len(subscriptionMap) == 0 {
			delete(eb.listeners, topic)
		}
	}, nil
}

func SubscribeTyped[T any](eb Bus, topic Topic, handler func(*Event, T)) (UnsubscribeFunc, error) {
	return eb.Subscribe(topic, func(evt *Event) {
		payload, ok := evt.Payload.(T)
		if !ok {
			zap.S().Warnf("Received event with invalid payload type for topic %q: expected %T, got %T", topic, payload, evt.Payload)
			return
		}
		handler(evt, payload)
	})
}
