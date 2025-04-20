package middlewares

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/requests"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/goodieshq/sweettooth/internal/server/roles"
	"github.com/google/uuid"
)

// middleware to extract the org ID from the url path /organizations/{orgid}...
func MiddlewareOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract the org ID from the URL path
		orgid, err := uuid.Parse(r.PathValue("orgid"))
		if err != nil {
			responses.ErrInvalidOrgID(w, r, err)
			return
		}

		// set the org ID for the request within the context
		r = requests.WithRequestOrgID(r, orgid)
		next.ServeHTTP(w, r)
	})
}

// middleware to extract the node ID from the url path /organizations/{orgid}/nodes/{nodeid}...
func MiddlewareOrganizationNode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract the org ID from the URL path
		nodeid, err := uuid.Parse(r.PathValue("nodeid"))
		if err != nil {
			responses.ErrInvalidNodeID(w, r, err)
			return
		}

		// set the org ID for the request within the context
		r = requests.WithRequestNodeID(r, nodeid)
		next.ServeHTTP(w, r)
	})
}

// handler middleware to set a minimum role for interaction
func OrgRoleMinimum(handler http.HandlerFunc, roleMinimum roles.OrgRole) http.HandlerFunc {
	if DEV_BYPASS_WEBAUTH {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if !requests.IsSuperAdmin(r) {
			orgid := requests.Oid(r)
			if orgid == nil {
				responses.ErrInvalidOrgID(w, r, nil)
				return
			}

			orgRoles := requests.ORoles(r)
			if orgRoles == nil {
				responses.ErrInvalidOrgRoles(w, r, nil)
				return
			}

			if requests.ORoles(r).GetRole(*orgid) <= roleMinimum {
				responses.ErrForbidden(w, r, nil)
				return
			}
		}

		// user has the required role
		handler.ServeHTTP(w, r)
	}
}
