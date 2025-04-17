package apiweb

import (
	"net/http"

	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/google/uuid"
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
	orgid, err := uuid.Parse(r.PathValue("orgid"))
	if err != nil {
		responses.ErrInvalidOrgID(w, r, err)
		return
	}

	org, err := h.core.GetOrganization(r.Context(), orgid)
	if err != nil {
		log.Error().Err(err).Msg("failed to get organizations")
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	responses.JsonResponse(w, r, http.StatusOK, org)
}
