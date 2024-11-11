package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goodieshq/sweettooth/pkg/api/server/middlewares"
	"github.com/goodieshq/sweettooth/pkg/api/server/responses"
	"github.com/goodieshq/sweettooth/pkg/cache"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/rs/zerolog/log"
)

const DEFAULT_CACHE_TIME = 10 * time.Minute
const DEFAULT_PORT = uint16(7777)
const DEFAULT_REDIS_PORT = uint16(6789)

type SweetToothServerConfig struct {
	CacheTime time.Duration // duration the cache should last (generally 2 or 3 check-in frequencies is good)
	RedisHost string        // redis cache host (if redis is desired)
	RedisPass string        // redis cache password (if redis is desired)
	RedisPort uint16        // redis cache port (if redis is desired, default redis port used)
	Host      string        // local address to listen on (default :: or 0.0.0.0)
	Port      uint16        // local port to listen on (default 7777)
	Secret    string        // used for JWT HMAC creation/validation for web interactions
}

type SweetToothServer struct {
	config SweetToothServerConfig
	cache  cache.Cache
	core   Core
}

func NewSweetToothServer(config SweetToothServerConfig, core Core) (*SweetToothServer, error) {
	// set the cache time, use the default if not provided
	if config.CacheTime <= 0 {
		config.CacheTime = DEFAULT_CACHE_TIME
	}

	// use the default port if not set
	if config.Port == 0 {
		config.Port = DEFAULT_PORT
	}

	var c cache.Cache

	cacheTime := config.CacheTime
	if cacheTime == 0 {
		cacheTime = DEFAULT_CACHE_TIME
	}

	if config.RedisHost != "" {
		port := config.RedisPort
		if port == 0 {
			port = DEFAULT_REDIS_PORT
		}

		c = cache.NewCacheRedis(config.RedisHost, port, config.RedisPass, config.CacheTime)
	} else {
		c = cache.NewCacheGo(cacheTime, 1*time.Minute)
	}

	return &SweetToothServer{
		config: config,
		cache:  c,
		core:   core,
	}, nil
}

func (srv *SweetToothServer) Run() error {
	// create the base router
	router := http.NewServeMux()

	// base middleware for all endpoints
	mwgBase := middlewares.MiddlewareGroup{}
	mwgBase.Add(middlewares.MiddlewareLogger()) // simple request and error logging
	mwgBase.Add(middlewares.MiddlewarePanic())  // recover from and log panics (technically if a panic occurs in the logger it could still crash)

	// middleware for all node interaction endpoints
	mwgNode := middlewares.MiddlewareGroup{}
	mwgNode.Add(mwgBase.Apply)                // include all middleware from the base group
	mwgNode.Add(middlewares.MiddlewareJSON()) // all responses should be JSON
	mwgNode.Add(srv.MiddlewareNodeAuth)       // add the Node authentication middleware

	/* Node API endpoints */

	// temporary, just here to let me clear the cache during development
	router.Handle("DELETE /api/v1/cache", mwgBase.ApplyFunc(srv.handleDeleteCache)) // temporary function to easily clear the cache

	// the simplest API endpoint: returns 204 on success. Use it to verify JWT auth.
	router.Handle("GET /api/v1/node/check", mwgNode.ApplyFunc(srv.handleGetNodeCheck)) // simple status check to determine registration/approval status

	// unauthenticated endpoint (though does require a valid signature).
	router.Handle("POST /api/v1/node/register", mwgBase.ApplyFunc(srv.handlePostNodeRegister)) // register and onboard a new node, unauthenticated but valid signature required

	// node should acquire a single array of all schedule entries that apply to it (assigned to node, group, etc...)
	router.Handle("GET /api/v1/node/schedule", mwgNode.ApplyFunc(srv.handleGetNodeSchedule)) // acquire the combined schedule entries for this node

	// node can query or update the server's inventory of its packages
	router.Handle("GET /api/v1/node/packages", mwgNode.ApplyFunc(srv.handleGetNodePackages))
	router.Handle("PUT /api/v1/node/packages", mwgNode.ApplyFunc(srv.handlePutNodePackages))

	// list package jobs
	router.Handle("GET /api/v1/node/packages/jobs", mwgNode.ApplyFunc(srv.handleGetNodePackagesJobs))
	router.Handle("GET /api/v1/node/packages/jobs/{$}", mwgNode.ApplyFunc(srv.handleGetNodePackagesJobs))

	// get details of a specific job (counts as an attempt as the details should only be acquired when ready to )
	router.Handle("GET /api/v1/node/packages/jobs/{id}", mwgNode.ApplyFunc(srv.handleGetNodePackagesJob))
	router.Handle("POST /api/v1/node/packages/jobs/{id}", mwgNode.ApplyFunc(srv.handlePostNodePackagesJob))

	// default API handler
	router.Handle("GET /api/", mwgBase.ApplyFunc(func(w http.ResponseWriter, r *http.Request) {
		responses.JsonErr(w, r, http.StatusNotImplemented, errors.New("this endpoint is not implemented"))
	}))

	/* Web endpoints */

	// middleware for all web administration endpoints
	mwgWeb := middlewares.MiddlewareGroup{}
	mwgWeb.Add(mwgBase.Apply) // include all middleware from the base group

	// Web interface
	router.Handle("GET /", mwgBase.ApplyFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// router should listen on the configured host/port
	listenStr := fmt.Sprintf("%s:%d", srv.config.Host, srv.config.Port)
	log.Info().Str("listen", listenStr).Msgf("Starting %s Server", info.APP_NAME)
	return http.ListenAndServe(listenStr, router)
}
