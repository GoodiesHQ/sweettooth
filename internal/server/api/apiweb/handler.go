package apiweb

import (
	"github.com/goodieshq/sweettooth/internal/server/cache"
	"github.com/goodieshq/sweettooth/internal/server/core"
)

type ApiWebHandler struct {
	cache cache.Cache
	core  core.Core
}

func NewApiNodeHandler(cache cache.Cache, core core.Core) *ApiWebHandler {
	return &ApiWebHandler{
		cache: cache,
		core:  core,
	}
}
