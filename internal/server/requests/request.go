package requests

import (
	"context"
	"net/http"
	"strconv"

	"github.com/goodieshq/sweettooth/internal/server/roles"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_ATTEMPTS_MAX = 5   //
	LIMIT_ATTEMPTS_MAX   = 100 // set a maximum limit of 100 attempts
)

type RequestState struct {
	// All requests can have an error associated with them
	Err error

	// Node Requests have an ID associated with them
	NodeID *uuid.UUID
}

func (rs *RequestState) IsNodeRequest() bool {
	return rs.NodeID != nil
}

type ContextKey string

// get the query parameter of attempts max
func RequestQueryAttemptsMax(r *http.Request) int {
	var attemptsMax int = 0

	// get and parse the attempts_max parameter
	if attemptsMaxStr := r.URL.Query().Get("attempts_max"); attemptsMaxStr != "" {
		// attempts_max should be an integer
		i, err := strconv.Atoi(attemptsMaxStr)
		if err != nil {
			log.Warn().Str("attempts_max", attemptsMaxStr).Msg("invalid attempts_max value")
		} else {
			if i > LIMIT_ATTEMPTS_MAX {
				i = LIMIT_ATTEMPTS_MAX
			}
			attemptsMax = i
		}
	}

	if attemptsMax == 0 {
		// default or invalid attepmts_max value set to a sane default
		attemptsMax = DEFAULT_ATTEMPTS_MAX
	} else if attemptsMax < 0 {
		// if any negative value was provided, set it to -1
		attemptsMax = -1
	}

	return attemptsMax
}

func NodeNID(r *http.Request) *uuid.UUID {
	if state := State(r); state != nil {
		return state.NodeID
	}
	return nil
}

func Nid(r *http.Request) *uuid.UUID {
	return r.Context().Value(ContextKey("nodeid")).(*uuid.UUID)
}

func Uid(r *http.Request) *uuid.UUID {
	return r.Context().Value(ContextKey("userid")).(*uuid.UUID)
}

func Oid(r *http.Request) *uuid.UUID {
	return r.Context().Value(ContextKey("orgid")).(*uuid.UUID)
}

func State(r *http.Request) *RequestState {
	if state, ok := r.Context().Value(ContextKey("state")).(*RequestState); ok {
		return state
	}
	return nil
}

func WithRequestContextValue(r *http.Request, key ContextKey, value interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, value))
}

func WithRequestState(r *http.Request, state *RequestState) *http.Request {
	// return r.WithContext(context.WithValue(r.Context(), ContextKey("state"), state))
	return WithRequestContextValue(r, "state", state)
}

// request coming from a super admin
func IsSuperAdmin(r *http.Request) bool {
	if superadmin, ok := r.Context().Value("superadmin").(bool); ok {
		return superadmin
	}
	return false
}

// request organization roles (web auth)
func ORoles(r *http.Request) *roles.OrgRoles {
	return r.Context().Value("orgroles").(*roles.OrgRoles)
}

// request pagination value
func Paging(r *http.Request) *api.Pagination {
	return r.Context().Value("pagination").(*api.Pagination)
}

func Err(r *http.Request) error {
	state := State(r)
	if state != nil {
		return state.Err
	}
	return nil
}

func SetRequestError(r *http.Request, err error) {
	state := State(r)
	if state == nil {
		log.Warn().Msg("request state is nil!")
		return
	}
	state.Err = err
}

func WithRequestPagination(r *http.Request, pagination *api.Pagination) *http.Request {
	return WithRequestContextValue(r, "pagination", pagination)
}

func SetNodeID(r *http.Request, nodeid uuid.UUID) {
	if state := State(r); state != nil {
		state.NodeID = &nodeid
	}
}

func WithRequestNodeID(r *http.Request, nodeid uuid.UUID) *http.Request {
	return WithRequestContextValue(r, "nodeid", &nodeid)
}

func WithRequestUserID(r *http.Request, userid uuid.UUID) *http.Request {
	return WithRequestContextValue(r, "userid", &userid)
}

func WithRequestOrgRoles(r *http.Request, orgRoles roles.OrgRoles) *http.Request {
	return WithRequestContextValue(r, "orgroles", &orgRoles)
}

func WithRequestOrgID(r *http.Request, orgid uuid.UUID) *http.Request {
	return WithRequestContextValue(r, "orgid", &orgid)
}

func WithRequestSuperAdmin(r *http.Request, superadmin bool) *http.Request {
	return WithRequestContextValue(r, "superadmin", superadmin)
}
