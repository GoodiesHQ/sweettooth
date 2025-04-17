package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/goodieshq/sweettooth/internal/server/api/apinode"
	"github.com/goodieshq/sweettooth/internal/server/api/apiweb"
	"github.com/goodieshq/sweettooth/internal/server/cache"
	"github.com/goodieshq/sweettooth/internal/server/core"
	"github.com/goodieshq/sweettooth/internal/server/middlewares"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/rs/zerolog/log"
)

const DEFAULT_CACHE_TIME = 10 * time.Minute
const DEFAULT_PORT = uint16(7373)
const DEFAULT_REDIS_PORT = uint16(6789)

type SweetToothServerConfig struct {
	CacheTime time.Duration // duration the cache should last (generally 2 or 3 check-in frequencies is good)
	RedisHost string        // redis cache host (if redis is desired)
	RedisPass string        // redis cache password (if redis is desired)
	RedisPort uint16        // redis cache port (if redis is desired, default redis port used)
	DBConnStr string        // DB connection string
	Host      string        // local address to listen on (default :: or 0.0.0.0)
	Port      uint16        // local port to listen on (default 7777)
	Secret    string        // used for JWT HMAC creation/validation for web interactions
}

type SweetToothServer struct {
	config *SweetToothServerConfig
	cache  cache.Cache
	core   core.Core
}

func NewSweetToothServer(config *SweetToothServerConfig, core core.Core) (*SweetToothServer, error) {
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

func (srv *SweetToothServer) ApiNodeHandlers(routerNode chi.Router) {
	// create the handler for all node-related API endpoints
	handlerNode := apinode.NewApiNodeHandler(srv.cache, srv.core)

	// Unauthorized endpoints do not require JWT tokens to be passed, e.g. for registering a new node
	routerNode.Group(func(routerNodeUnauthorized chi.Router) {
		// register and onboard a new node, unauthenticated but valid signature required
		routerNodeUnauthorized.Post("/api/v1/node/register", handlerNode.HandlePostNodeRegister)
	})

	// Authorized endpoints require a valid JWT token from a registered node
	routerNode.Group(func(routerNodeAuthorized chi.Router) {
		// all endpoints in this group require a valid JWT token signed by the node's private key
		routerNodeAuthorized.Use(
			middlewares.MiddlewareAuthNode(srv.core, srv.cache),
		)
		// the simplest API endpoint: returns 204 on success. Use it to verify JWT auth.
		routerNodeAuthorized.Get("/api/v1/node/check", handlerNode.HandleGetNodeCheck)
		// node should acquire a single array of all schedule entries that apply to it (assigned to node, group, etc...)
		routerNodeAuthorized.Get("/api/v1/node/schedule", handlerNode.HandleGetNodeSchedule) // acquire the combined schedule entries for this node
		// node can query or update the server's inventory of its packages
		routerNodeAuthorized.Get("/api/v1/node/packages", handlerNode.HandleGetNodePackages)
		routerNodeAuthorized.Put("/api/v1/node/packages", handlerNode.HandlePutNodePackages)
		// list package jobs
		routerNodeAuthorized.Get("/api/v1/node/packages/jobs", handlerNode.HandleGetNodePackagesJobs)
		routerNodeAuthorized.Get("/api/v1/node/packages/jobs/", handlerNode.HandleGetNodePackagesJobs)
		// get details of a specific job (counts as an attempt as the details should only be acquired when ready to )
		routerNodeAuthorized.Get("/api/v1/node/packages/jobs/{id}", handlerNode.HandleGetNodePackagesJob)
		routerNodeAuthorized.Post("/api/v1/node/packages/jobs/{id}", handlerNode.HandlePostNodePackagesJob)
	})
}

func (srv *SweetToothServer) ApiWebHandlers(routerWeb chi.Router) {
	handlerWeb := apiweb.NewApiNodeHandler(srv.cache, srv.core)
	routerWeb.Group(func(routerWebUnauthorized chi.Router) {
		routerWebUnauthorized.Post("/api/v1/web/login", handlerWeb.HandlePostWebLogin)
	})

	routerWeb.Group(func(routerWebAuthorized chi.Router) {
		routerWebAuthorized.Use(
			// TODO: add authentication middleware
		)

		// GET /api/v1/web/organizations
		routerWebAuthorized.Get("/api/v1/web/organizations", handlerWeb.HandleGetWebOrganizations)
		// GET /api/v1/web/organizations/summaries
		routerWebAuthorized.Get("/api/v1/web/organizations/summaries", handlerWeb.HandleGetWebOrganizationSummaries)
		// GET /api/v1/web/organizations/{orgid}
		routerWebAuthorized.Get("/api/v1/web/organizations/{orgid}", handlerWeb.HandleGetWebOrganization)
	})
}

func (srv *SweetToothServer) Run() error {
	// create the base router
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		// All API endpoints will use the logger and panic middleware
		r.Use(
			middlewares.MiddlewareLogger, // chi middleware to log requests
			middlewares.MiddlewarePanic,  // chi middleware to recover from panics
			middlewares.MiddlewareJSON,   // chi middleware to handle JSON requests
		)

		r.Group(srv.ApiNodeHandlers)
		r.Group(srv.ApiWebHandlers)
	})

	// serve static files
	staticFileServer := http.FileServer(http.Dir("./internal/server/web/static"))
	router.Handle("/static/*", http.StripPrefix("/static/", staticFileServer))

	// router should listen on the configured host/port
	listenStr := fmt.Sprintf("%s:%d", srv.config.Host, srv.config.Port)
	log.Info().Str("listen", listenStr).Msgf("Starting %s Server", info.APP_NAME)
	err := http.ListenAndServe(listenStr, router)

	if err != nil {
		log.Error().Err(err).Msg("failed to listen")
		return err
	}

	return nil
}
