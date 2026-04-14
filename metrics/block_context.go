package metrics

import (
	"context"
	"net/http"

	"github.com/nunoOliveiraqwe/torii/internal/ctxkeys"
)

var blockInfoContextKey = ctxkeys.BlockInfo

type BlockInfo struct {
	Middleware string
	Reason     string
}

func CreateAndAddBlockInfo(r *http.Request, middleware, reason string) {
	b := &BlockInfo{
		Middleware: middleware,
		Reason:     reason,
	}
	SetBlockInfo(r, b)
}

func SetBlockInfo(r *http.Request, info *BlockInfo) {
	ctx := context.WithValue(r.Context(), blockInfoContextKey, info)
	*r = *r.WithContext(ctx)
}

func GetBlockInfo(r *http.Request) *BlockInfo {
	v := r.Context().Value(blockInfoContextKey)
	if v == nil {
		return nil
	}
	info, ok := v.(*BlockInfo)
	if !ok {
		return nil
	}
	return info
}
