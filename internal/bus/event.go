package bus

import (
	"net/http"
	"time"
)

type Event struct {
	ID      string
	Topic   []Topic
	At      time.Time
	Source  string
	Request *RequestRef
	Payload any
}

type RequestRef struct {
	ClientIP          string
	InternalRequestID string
	Method            string
	Host              string
	Path              string
	StatusCode        int
}

func NewEventFromRequest(topic []Topic, statusCode int, source, internalRequestId, clientIp string, r *http.Request,
	payload any) *Event {
	return &Event{
		ID:     generateEventID(),
		Topic:  topic,
		At:     time.Now(),
		Source: source,
		Request: &RequestRef{
			StatusCode:        statusCode,
			InternalRequestID: internalRequestId,
			ClientIP:          clientIp,
			Method:            r.Method,
			Host:              r.Host,
			Path:              r.URL.Path,
		},
		Payload: payload,
	}
}

func NewEvent(topic []Topic, statusCode int, source, internalRequestID, clientIP string, payload any) *Event {
	return &Event{
		ID:     generateEventID(),
		Topic:  topic,
		At:     time.Now(),
		Source: source,
		Request: &RequestRef{
			InternalRequestID: internalRequestID,
			ClientIP:          clientIP,
		},
		Payload: payload,
	}
}

func generateEventID() string {
	return "event-" + time.Now().Format("20060102150405") //every time I look at go's format I cry
}
