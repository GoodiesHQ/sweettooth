package middlewares

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/server/cache"
	"github.com/goodieshq/sweettooth/internal/server/core"
	"github.com/goodieshq/sweettooth/internal/server/requests"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/rs/zerolog/log"
)

func MiddlewareAuthNode(core core.Core, cache cache.Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Trace().Msg("starting middleware node auth")
			// extract the bearer token from the Authorization header
			tokenString := ExtractBearerToken(r)
			if tokenString == "" {
				responses.ErrNodeTokenInvalid(w, r, nil)
				return
			}
			log.Trace().Msg("extracted bearer token")

			// verify that the JWT was signed with the key that it says it was signed with
			nodeid, _, err := crypto.VerifyNodeJWT(tokenString)
			if err != nil {
				log.Debug().Err(err).Msg("jwt was unverified")
				responses.ErrNodeTokenInvalid(w, r, err)
				return
			}

			// set the node ID for the request via the state
			nodeidString := nodeid.String()
			requests.SetNodeID(r, nodeid)
			log.Trace().Str("nodeid", nodeidString).Msg("node ID added to the request")

			// At this point, all we know is that the signature is valid and well-formed. Check cache/db for node validity
			found, authorized := cache.GetNodeAuth(nodeidString)
			if !found {
				// node ID was not found in the cache, check the database
				log.Debug().Str("nodeid", nodeidString).Msg("node ID auth cache miss, checking database")
				node, err := core.GetNode(r.Context(), nodeid)

				if err != nil {
					log.Trace().Any("nodeid", nodeidString).Bool("approved", node.Approved).Msg("node found in database")
				}

				// authorized is true if there is no auth error
				authorized = !nodeAuthErr(w, r, err, node)
				if !authorized {
					responses.ErrNodeNotApproved(w, r, nil)
					return
				}

				// at this point, we know fprint validity. Put it in the cache.
				cache.SetNodeAuth(nodeidString, authorized)
			}

			if !authorized {
				return
			}

			if err := core.Seen(r.Context(), nodeid); err != nil {
				responses.ErrServiceUnavailable(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// sends an API error if there is an error to send, returns true if there was an error
func nodeAuthErr(w http.ResponseWriter, r *http.Request, err error, node *api.Node) bool {
	if err == nil {
		if node == nil {
			responses.ErrNodeNotFound(w, r, nil)
			return true
		}
		if !node.Approved {
			log.Debug().Msg("node '" + node.ID.String() + "' is registered but not approved")
			responses.ErrNodeNotApproved(w, r, nil)
			return true
		}
		return false
	} else {
		responses.ErrServerError(w, r, err)
		return true
	}
}
