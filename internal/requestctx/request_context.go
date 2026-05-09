package requestctx

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/bus"
	"github.com/nunoOliveiraqwe/torii/internal/ctxkeys"
	"github.com/nunoOliveiraqwe/torii/internal/netutil"
	"github.com/nunoOliveiraqwe/torii/metrics"
	"go.uber.org/zap"
)

var requestIDCounter uint64

type RequestContextStruct struct {
	BlockInfo         *BlockInfo
	CountryCode       string
	ContinentCode     string
	InternalRequestId string //used internally to match requests. requestId is an optional middleware that might not be present
	RequestId         string
	Logger            *zap.Logger
	ReqMetrics        *metrics.RequestMetric
}

type responseWriterWithStatus struct {
	http.ResponseWriter
	statusCode  int
	bytesSent   int64
	wroteHeader bool
}

func (w *responseWriterWithStatus) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWithStatus) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesSent += int64(n)
	return n, err
}

func (w *responseWriterWithStatus) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriterWithStatus) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func InjectContextStruct(bCtx BuildContext, next http.HandlerFunc) http.HandlerFunc {
	return InjectContextStructWithBuildContextAndLog(&bCtx, next)
}

func InjectContextStructWithBuildContextAndLog(bCtx *BuildContext, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		incomingTime := time.Now()
		ctxStuct := r.Context().Value(ctxkeys.ContextStruct)
		if ctxStuct != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		c := RequestContextStruct{
			InternalRequestId: generateInternalRequestId(),
		}
		ctx = context.WithValue(ctx, ctxkeys.ContextStruct, &c)
		r = r.WithContext(ctx)
		ww := &responseWriterWithStatus{
			ResponseWriter: w,
		}
		startTime := time.Now()
		next.ServeHTTP(ww, r)
		elapsedTime := time.Since(startTime).Milliseconds()
		statusCode := ww.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		if bCtx != nil {
			raiseProcessedRequestEvent(incomingTime, elapsedTime, bCtx.ServerID, bCtx.ConnectionName(), c.InternalRequestId,
				c.CountryCode, c.ContinentCode, r.UserAgent(),
				statusCode, ww.bytesSent, r, bCtx.EventBus)
		}
	}
}

func raiseProcessedRequestEvent(incomingTime time.Time,
	elapsedTime int64, serverId, connectionName, internalId, countryCode, continentCode, ua string,
	statusCode int, bytesSent int64, r *http.Request, evt bus.Bus) {
	if evt == nil {
		return
	}
	ctxStruct := GetContextStruct(r)
	clientIP, err := netutil.GetClientIP(r)
	if err != nil {
		clientIP = r.RemoteAddr
	}

	var blocked *bus.RequestBlocked
	if ctxStruct.BlockInfo != nil {
		blocked = &bus.RequestBlocked{
			Middleware: ctxStruct.BlockInfo.Middleware,
			Reason:     ctxStruct.BlockInfo.Reason,
			StatusCode: statusCode,
			Action:     bus.BlockActionDeny,
		}
	}

	rp := bus.RequestProcessed{
		ArrivedAt:        incomingTime.UnixMilli(),
		ProcessingTimeMs: elapsedTime,
		StatusCode:       statusCode,
		InternalId:       internalId,
		ConnectionName:   connectionName,
		CountryCode:      countryCode,
		ContinentCode:    continentCode,
		UserAgent:        ua,
		BytesSent:        bytesSent,
		BytesReceived:    r.ContentLength,
		Blocked:          blocked,
	}

	evt.Publish(bus.NewEventFromRequest(
		[]bus.Topic{bus.TopicRequestProcessed},
		statusCode,
		serverId,
		ctxStruct.InternalRequestId,
		clientIP,
		r,
		rp,
	))
}

func generateInternalRequestId() string {
	nextId := atomic.AddUint64(&requestIDCounter, 1)
	return fmt.Sprintf("internal-id-%d", nextId)
}

func GetContextStruct(r *http.Request) *RequestContextStruct {
	if r == nil {
		zap.S().Error("GetContextStruct: request is nil")
		return &RequestContextStruct{}
	}
	v := r.Context().Value(ctxkeys.ContextStruct)
	if v == nil {
		// This should never happen, but if it does, we return an empty contextStruct to avoid panics in the middleware.
		return &RequestContextStruct{}
	}
	c, ok := v.(*RequestContextStruct)
	if !ok {
		// This should also never happen, but we log an error and return an empty contextStruct to avoid panics.
		zap.S().Error("GetContextStruct: context value is not of type *contextStruct")
		return &RequestContextStruct{}
	}
	return c
}
