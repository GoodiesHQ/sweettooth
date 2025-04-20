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
	"github.com/goodieshq/sweettooth/internal/server/roles"
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
	Secret    []byte        // used for JWT HMAC creation/validation for web interactions
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

		c = cache.NewCacheRedis(config.RedisHost, port, config.RedisPass, cacheTime)
	} else {
		c = cache.NewCacheGo(cacheTime, 1*time.Minute)
	}

	return &SweetToothServer{
		config: config,
		cache:  c,
		core:   core,
	}, nil
}

// Apply all Node related API endpoints to the router
func (srv *SweetToothServer) ApiNodeHandlers(routerNode chi.Router) {
	handlerNode := apinode.NewApiNodeHandler(srv.cache, srv.core)

	// Unauthorized endpoints do not require JWT tokens to be passed, e.g. for registering a new node
	routerNode.Group(func(routerNodeUnauthorized chi.Router) {
		// register a new node
		routerNodeUnauthorized.Post("/register", handlerNode.HandlePostNodeRegister)
	})

	// Authorized endpoints require a valid JWT token from a registered node
	routerNode.Group(func(routerNodeAuthorized chi.Router) {
		// all endpoints in this group require a valid JWT token signed by the node's private key
		routerNodeAuthorized.Use(
			middlewares.MiddlewareAuthNode(srv.core, srv.cache),
		)
		// the simplest API endpoint: returns 204 on success. Use it to verify JWT auth.
		routerNodeAuthorized.Get("/check", handlerNode.HandleGetNodeCheck)
		// node should acquire a single array of all schedule entries that apply to it (assigned to node, group, etc...)
		routerNodeAuthorized.Get("/schedule", handlerNode.HandleGetNodeSchedule) // acquire the combined schedule entries for this node
		// node can query or update the server's inventory of its packages
		routerNodeAuthorized.Get("/packages", handlerNode.HandleGetNodePackages)
		routerNodeAuthorized.Put("/packages", handlerNode.HandlePutNodePackages)
		// list package jobs
		routerNodeAuthorized.Get("/packages/jobs", handlerNode.HandleGetNodePackagesJobs)
		routerNodeAuthorized.Get("/packages/jobs/", handlerNode.HandleGetNodePackagesJobs)
		// get details of a specific job (counts as an attempt as the details should only be acquired when ready to )
		routerNodeAuthorized.Get("/packages/jobs/{id}", handlerNode.HandleGetNodePackagesJob)
		routerNodeAuthorized.Post("/packages/jobs/{id}", handlerNode.HandlePostNodePackagesJob)
	})
}

func (srv *SweetToothServer) ApiWebHandlers(routerWeb chi.Router) {
	handlerWeb := apiweb.NewApiNodeHandler(srv.cache, srv.core)
	routerWeb.Group(func(routerWebUnauthorized chi.Router) {
		routerWebUnauthorized.Post("/login", handlerWeb.HandlePostWebLogin(srv.config.Secret))
	})

	routerWeb.Group(func(routerWebAuthorized chi.Router) {
		routerWebAuthorized.Use(
			middlewares.MiddlewareAuthWeb(srv.core, srv.cache, srv.config.Secret),
			middlewares.MiddlewarePaginate,
		)

		// GET /api/v1/web/organizations
		routerWebAuthorized.Get("/organizations", handlerWeb.HandleGetWebOrganizations)
		routerWebAuthorized.Get("/organizations_summary", handlerWeb.HandleGetWebOrganizationSummaries)

		routerWebAuthorized.Route("/organizations/{orgid}", func(routerOrg chi.Router) {
			routerOrg.Use(
				// Every request context will have an org ID extracted from the URL
				middlewares.MiddlewareOrganization,
			)

			// GET /api/v1/web/organizations/{orgid}
			routerOrg.Get(
				"/",
				middlewares.OrgRoleMinimum(handlerWeb.HandleGetWebOrganization, roles.READER),
			)

			// GET /api/v1/web/organizations/{orgid}/nodes
			routerOrg.Get(
				"/nodes",
				middlewares.OrgRoleMinimum(handlerWeb.HandleGetWebOrganizationNodes, roles.READER),
			)
		})
	})
}

func (srv *SweetToothServer) Run() error {
	// create the base router
	router := chi.NewRouter()
	router.Use(
		middlewares.MiddlewareState,  // chi middleware to set the request state
		middlewares.MiddlewareLogger, // chi middleware to log requests
		middlewares.MiddlewarePanic,  // chi middleware to recover from panics
		middlewares.MiddlewareJSON,   // chi middleware to handle JSON requests
	)

	router.Group(func(r chi.Router) {
		// All API endpoints will use the logger and panic middleware
		r.Route("/api/v1/node", srv.ApiNodeHandlers)
		r.Route("/api/v1/web", srv.ApiWebHandlers)
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
