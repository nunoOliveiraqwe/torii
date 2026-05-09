package bus

import (
	"go.uber.org/zap"
)

type handlerJob struct {
	event   *Event
	handler ListenerCallback
}

func dispatchMessages(eb *EventBus) {
	for {
		select {
		case msg := <-eb.busChannel:
			dispatchEvent(eb, msg)
		case <-eb.ctx.Done():
			zap.S().Infof("Worker received shutdown signal, exiting")
			return
		}
	}
}

func dispatchEvent(eb *EventBus, event *Event) {
	eb.listenerLock.RLock()
	handlers := make([]ListenerCallback, 0)

	for _, t := range event.Topic {
		handlersByID, exists := eb.listeners[t]
		if !exists {
			continue
		}

		for _, handler := range handlersByID {
			handlers = append(handlers, handler)
		}
	}
	eb.listenerLock.RUnlock()

	for _, handler := range handlers {
		dispatchHandler(eb, handler, event)
	}
}

func dispatchHandler(eb *EventBus, handler ListenerCallback, event *Event) {
	select {
	case eb.handlerQueue <- handlerJob{event: event, handler: handler}:
	case <-eb.ctx.Done():
	}
}

func handleMessages(eb *EventBus) {
	for {
		select {
		case job := <-eb.handlerQueue:
			safeHandleEvent(job.handler, job.event)
		case <-eb.ctx.Done():
			zap.S().Infof("Handler worker received shutdown signal, exiting")
			return
		}
	}
}

func safeHandleEvent(handler ListenerCallback, event *Event) {
	defer func() {
		if r := recover(); r != nil {
			zap.S().Errorf("Event handler panicked for topic %q: %v", event.Topic, r)
		}
	}()

	handler(event)
}
