package apinode

import (
	"github.com/goodieshq/sweettooth/internal/server/cache"
	"github.com/goodieshq/sweettooth/internal/server/core"
)

type ApiNodeHandler struct {
	cache cache.Cache
	core  core.Core
}

func NewApiNodeHandler(cache cache.Cache, core core.Core) *ApiNodeHandler {
	return &ApiNodeHandler{
		cache: cache,
		core:  core,
	}
}
