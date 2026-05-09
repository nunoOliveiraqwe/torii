package requestctx

import (
	"net/http"

	"github.com/nunoOliveiraqwe/torii/internal/bus"
)

type BlockInfo struct {
	Middleware string
	Reason     string
	Topic      bus.Topic
}

func CreateAndAddBlockInfoToRequestContext(r *http.Request, middleware, reason string, topic bus.Topic) {
	b := &BlockInfo{
		Middleware: middleware,
		Reason:     reason,
		Topic:      topic,
	}
	SetBlockInfo(r, b)
}

func SetBlockInfo(r *http.Request, info *BlockInfo) {
	ctxStruct := GetContextStruct(r)
	ctxStruct.BlockInfo = info
}

func GetBlockInfo(r *http.Request) *BlockInfo {
	ctxStruct := GetContextStruct(r)
	return ctxStruct.BlockInfo
}
