package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

const DEFAULT_CACHE_TIME = 10 * time.Minute
const DEFAULT_PORT = uint16(7777)

type SweetToothServerConfig struct {
	CacheTimeID time.Duration // duration the cache should last (generally 2 or 3 check-in frequencies is good)
	CheckInFreq time.Duration // the time in between check-ins from the client
	Host        string        // local address to listen on (default :: or 0.0.0.0)
	Port        uint16        // local port to listen on (default 7777)
	Secret      string        // used for JWT HMAC creation/validation for web interactions
}

type SweetToothServer struct {
	config            SweetToothServerConfig
	cacheValidNodeIDs cache.Cache // recently-seen node IDs which are known to be valid/invalid
	core              Core
}

func NewSweetToothServer(config SweetToothServerConfig, core Core) (*SweetToothServer, error) {
	// set the cache time, use the default if not provided
	if config.CacheTimeID <= 0 {
		config.CacheTimeID = DEFAULT_CACHE_TIME
	}

	// use the default port if not set
	if config.Port == 0 {
		config.Port = DEFAULT_PORT
	}

	return &SweetToothServer{
		config:            config,
		cacheValidNodeIDs: *cache.New(config.CacheTimeID, time.Minute), // TODO: determine if magic number 1 minute cleanup interval is ok
		core:              core,
	}, nil
}

func (srv *SweetToothServer) Run() error {
	// create the base router
	router := http.NewServeMux()

	// base middleware for all endpoints
	mwgBase := MiddlewareGroup{}
	mwgBase.Add(MiddlewareLogger) // simple request logging
	mwgBase.Add(MiddlewarePanic)  // recover from and log panics
	mwgBase.Add(MiddlewareJSON)   // all responses should be JSON

	// middleware for all node interaction endpoints
	mwgNode := MiddlewareGroup{}
	mwgNode.Add(mwgBase.Apply)
	mwgNode.Add(srv.MiddlewareNodeAuth)

	// middleware for all web administration endpoints
	mwgWeb := MiddlewareGroup{}
	mwgWeb.Add(mwgBase.Apply)

	// basic index page, TODO: host static application
	router.Handle("GET /", mwgBase.Apply(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JsonResponse(w, http.StatusOK, map[string]interface{}{"status": "success", "a": 123})
	})))

	// implement endpoints
	router.Handle("GET /api/v1/node/check", mwgNode.ApplyFunc(srv.handleGetNodeCheck))
	router.Handle("POST /api/v1/node/register", mwgBase.ApplyFunc(srv.handlePostNodeRegister))

	// listen on the configured host/port
	listenStr := fmt.Sprintf("%s:%d", srv.config.Host, srv.config.Port)
	log.Info().Str("listen", listenStr).Msgf("Starting %s Server", config.APP_NAME)
	return http.ListenAndServe(listenStr, router)
}
