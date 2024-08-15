package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/util"
)

// DELETE /api/v1/cache
func (srv *SweetToothServer) handleDeleteCache(w http.ResponseWriter, r *http.Request) {
	srv.cacheValidNodeIDs.Flush()
	JsonResponse(w, r, http.StatusNoContent, nil)
}

// GET /api/v1/node/check
func (srv *SweetToothServer) handleGetNodeCheck(w http.ResponseWriter, r *http.Request) {
	// a successful check should only yield a 204
	// this indicates a properly formed and valid JWT, node ID exists in the database and is enabled
	JsonResponse(w, r, http.StatusNoContent, nil)
}

// GET /api/v1/node/schedules
func (srv *SweetToothServer) handleGetNodeSchedule(w http.ResponseWriter, r *http.Request) {
	sched, err := srv.core.GetNodeSchedule(r.Context(), *util.Rid(r))
	if err != nil {
		ErrServerError(w, r, err)
		return
	}

	JsonResponse(w, r, http.StatusOK, sched)
}

// POST /api/v1/node/register
func (srv *SweetToothServer) handlePostNodeRegister(w http.ResponseWriter, r *http.Request) {
	var req api.RegistrationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	// verify the signature of the request so we aren't wasting our time

	// extract the public key
	pubkeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil {
		ErrInvalidRequestBody(w, r, err)
		return
	}

	if !crypto.VerifyBase64(ed25519.PublicKey(pubkeyBytes), pubkeyBytes, req.PublicKeySig) {
		ErrInvalidRequestBody(w, r, errors.New("request signature is invalid"))
		return
	}

	fprint := crypto.Fingerprint(pubkeyBytes)
	util.SetRequestNodeID(r, fprint)

	// node registration
	node, err := srv.core.GetNode(r.Context(), fprint)
	if err != nil {
		ErrServerError(w, r, err)
		return
	}

	if node != nil {
		if node.Approved {
			JsonResponse(w, r, http.StatusNoContent, nil) // node is already registered
			return
		}
		ErrNodeNotApproved(w, r, nil)
		return
	}

	if req.OrganizationID != nil {
		org, err := srv.core.GetOrganization(r.Context(), *req.OrganizationID)
		if err != nil {
			ErrServerError(w, r, err)
			return
		}
		if org == nil {
			ErrOrgNotFound(w, r, err)
			return
		}
		// org is valid
	}

	// node does not exist
	err = srv.core.CreateNode(r.Context(), req)
	if err != nil {
		ErrServerError(w, r, err)
		return
	}

	JsonResponse(w, r, http.StatusCreated, nil)
}

// PUT /api/v1/node/packages
func (srv *SweetToothServer) handlePutNodePackages(w http.ResponseWriter, r *http.Request) {
	var packages api.Packages
	err := json.NewDecoder(r.Body).Decode(&packages)
	if err != nil {
		ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	err = srv.core.UpdateNodePackages(r.Context(), *util.Rid(r), &packages)
	if err != nil {
		ErrServiceUnavailable(w, r, err)
		return
	}
	JsonResponse(w, r, http.StatusNoContent, nil)
}
