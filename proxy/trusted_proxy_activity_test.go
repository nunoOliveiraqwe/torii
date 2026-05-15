package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nunoOliveiraqwe/torii/config"
	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/requestctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type activityCaptureBus struct {
	event *bus.Event
}

func (b *activityCaptureBus) Start() {}

func (b *activityCaptureBus) Stop() {}

func (b *activityCaptureBus) Publish(event *bus.Event) bool {
	b.event = event
	return true
}

func (b *activityCaptureBus) Subscribe(_ bus.Topic, _ bus.ListenerCallback) (bus.UnsubscribeFunc, error) {
	return func() {}, nil
}

func TestGlobalTrustedProxyResolvedClientIPReachesActivityEvent(t *testing.T) {
	eventBus := &activityCaptureBus{}
	buildCtx := requestctx.NewBuildContext(nil, nil, eventBus, 0, "", "", "", "").
		WithRuntimeContext(context.Background()).
		WithPort(8092).
		WithServerID("http-8092")

	dispatcher, err := initGlobalDispatcher(buildCtx, &config.GlobalConfig{
		TrustedProxies: &config.TrustedProxiesConfig{
			Ranges: []string{"127.0.0.1/32"},
			Header: "X-Forwarded-For",
		},
	})
	require.NoError(t, err)

	var downstreamRemoteAddr string
	routeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamRemoteAddr = r.RemoteAddr
		w.WriteHeader(http.StatusNoContent)
	})
	handler := dispatcher.registerHandler(8092, routeHandler)
	handler = requestctx.InjectContextStruct(buildCtx, handler)

	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	req.RemoteAddr = "127.0.0.1:43210"
	req.Header.Set("X-Forwarded-For", "203.0.113.42")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, "203.0.113.42:43210", downstreamRemoteAddr)
	require.NotNil(t, eventBus.event)
	require.NotNil(t, eventBus.event.Request)
	assert.Equal(t, "203.0.113.42:43210", eventBus.event.Request.ClientIP)
}
