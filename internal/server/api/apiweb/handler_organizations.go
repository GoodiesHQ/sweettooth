package apiweb

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/requests"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/rs/zerolog/log"
)

// GET /api/v1/web/organizations
func (h *ApiWebHandler) HandleGetWebOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.core.GetOrganizations(r.Context())
	if err != nil {
		responses.ErrServiceUnavailable(w, r, err)
		return
	}
	responses.JsonResponse(w, r, http.StatusOK, orgs)
}

// GET /api/v1/web/organizations/summaries
func (h *ApiWebHandler) HandleGetWebOrganizationSummaries(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.core.GetOrganizationSummaries(r.Context())
	if err != nil {
		responses.ErrServiceUnavailable(w, r, err)
		return
	}
	responses.JsonResponse(w, r, http.StatusOK, orgs)
}

// GET /api/v1/web/organizations/{orgid}
func (h *ApiWebHandler) HandleGetWebOrganization(w http.ResponseWriter, r *http.Request) {
	orgid := requests.Oid(r)
	if orgid == nil {
		responses.ErrInvalidOrgID(w, r, nil)
		return
	}

	org, err := h.core.GetOrganization(r.Context(), *orgid)
	if err != nil {
		log.Error().Err(err).Msg("failed to get organizations")
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	responses.JsonResponse(w, r, http.StatusOK, org)
}

// GET /api/v1/web/organizations/{orgid}/nodes
func (h *ApiWebHandler) HandleGetWebOrganizationNodes(w http.ResponseWriter, r *http.Request) {
	orgid := requests.Oid(r)
	if orgid == nil {
		responses.ErrInvalidOrgID(w, r, nil)
		return
	}

	nodes, err := h.core.GetNodes(r.Context(), *orgid)
	if err != nil {
		log.Error().Err(err).Msg("failed to get organizations")
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	responses.JsonResponse(w, r, http.StatusOK, nodes)
}
