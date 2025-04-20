package apinode

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/server/requests"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
)

const (
	DEFAULT_ATTEMPTS_MAX int = 5
)

// GET /api/v1/node/check
func (h *ApiNodeHandler) HandleGetNodeCheck(w http.ResponseWriter, r *http.Request) {
	// a successful check should only yield a 204
	// this should indicate a properly formed and valid JWT, node ID exists in the database and is enabled
	// must be protected by middleware that does the above
	responses.JsonResponse(w, r, http.StatusNoContent, nil)
}

// GET /api/v1/node/schedules
func (h *ApiNodeHandler) HandleGetNodeSchedule(w http.ResponseWriter, r *http.Request) {
	sched, err := h.core.GetNodeSchedule(r.Context(), *requests.NodeNID(r))
	if err != nil {
		responses.ErrServerError(w, r, err)
		return
	}

	responses.JsonResponse(w, r, http.StatusOK, sched)
}

// POST /api/v1/node/register
func (h *ApiNodeHandler) HandlePostNodeRegister(w http.ResponseWriter, r *http.Request) {
	var req api.RegistrationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		responses.ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	// verify the signature of the request so we aren't wasting our time

	// extract the public key
	pubkeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil {
		responses.ErrInvalidRequestBody(w, r, err)
		return
	}

	if !crypto.VerifyBase64(ed25519.PublicKey(pubkeyBytes), pubkeyBytes, req.PublicKeySig) {
		responses.ErrInvalidRequestBody(w, r, errors.New("request signature is invalid"))
		return
	}

	fprint := crypto.Fingerprint(pubkeyBytes)
	r = requests.WithRequestNodeID(r, fprint)

	// check node existince
	node, err := h.core.GetNode(r.Context(), fprint)
	if err != nil {
		responses.ErrServerError(w, r, err)
		return
	}

	if node != nil {
		// node is already registered, do not continue
		if node.Approved {
			responses.JsonResponse(w, r, http.StatusNoContent, nil)
			return
		}
		// node is not approved, should 403
		responses.ErrNodeNotApproved(w, r, nil)
		return
	}

	// get the registration token and verify it is valid
	orgid, err := h.core.ProcessRegistrationToken(r.Context(), req.Token)
	if err != nil {
		if h.core.ErrNotFound(err) {
			responses.ErrRegistrationTokenInvalid(w, r, err)
			return
		}
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	if orgid == nil {
		responses.ErrRegistrationTokenInvalid(w, r, nil)
		return
	}

	// node does not exist and registration token does point to a valid organization
	_, err = h.core.CreateNode(r.Context(), req)
	if err != nil {
		responses.ErrServerError(w, r, err)
		return
	}

	responses.JsonResponse(w, r, http.StatusCreated, nil)
}

// GET /api/v1/node/packages
func (h *ApiNodeHandler) HandleGetNodePackages(w http.ResponseWriter, r *http.Request) {
	// get a list of currently installed packages (choco, system, outdated)
	pkg, err := h.core.GetNodePackages(r.Context(), *requests.NodeNID(r))
	if err != nil {
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	if pkg == nil {
		responses.ErrNodeNotFound(w, r, errors.New("couldn't acquire node package"))
		return
	}

	// return the existing packages
	responses.JsonResponse(w, r, http.StatusOK, pkg)
}

// PUT /api/v1/node/packages
func (h *ApiNodeHandler) HandlePutNodePackages(w http.ResponseWriter, r *http.Request) {
	var packages api.Packages

	// decode the submitted packages
	err := json.NewDecoder(r.Body).Decode(&packages)
	if err != nil {
		responses.ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	// update the node's packages in the database if they are different
	err = h.core.UpdateNodePackages(r.Context(), *requests.NodeNID(r), &packages)
	if err != nil {
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	// always return with a 204 if everything was successful, regardless of if packages were updated or not
	responses.JsonResponse(w, r, http.StatusNoContent, nil)
}

// GET /api/v1/node/packages/jobs
func (h *ApiNodeHandler) HandleGetNodePackagesJobs(w http.ResponseWriter, r *http.Request) {
	// get the node ID from the request's JWT
	nodeid := *requests.NodeNID(r)
	attemptsMax := requests.RequestQueryAttemptsMax(r)

	// get job list from the database
	jobs, err := h.core.GetPackageJobList(r.Context(), nodeid, attemptsMax)
	if err != nil {
		responses.ErrServiceUnavailable(w, r, errors.New("failed to get the job list"))
		return
	}

	// return the jobs to the client
	responses.JsonResponse(w, r, http.StatusOK, jobs)
}

// GET /api/v1/node/packages/jobs/{id}
func (h *ApiNodeHandler) HandleGetNodePackagesJob(w http.ResponseWriter, r *http.Request) {
	// get the node ID
	jobid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		responses.ErrInvalidJobID(w, r, err)
		return
	}

	nodeid := *requests.NodeNID(r)
	attemptsMax := requests.RequestQueryAttemptsMax(r)

	// get the job from the database if it exists
	job, err := h.core.GetPackageJob(r.Context(), jobid)
	if err != nil {
		if h.core.ErrNotFound(err) {
			responses.ErrJobMissingOrExpired(w, r, errors.New("unable to find the job ID"))
		} else {
			responses.ErrServerError(w, r, err)
		}
		return
	}

	// check that the job is assigned to this node
	if job.NodeID != nodeid {
		responses.ErrJobMissingOrExpired(w, r, errors.New("the job ID is not assigned to this node"))
		return
	}

	job, err = h.core.AttemptPackageJob(r.Context(), jobid, nodeid, attemptsMax)
	if err != nil {
		responses.ErrServerError(w, r, err)
		return
	}

	if job == nil {
		responses.ErrJobMissingOrExpired(w, r, errors.New("no job ID found for this node"))
		return
	}

	responses.JsonResponse(w, r, http.StatusOK, job)
}

// POST /api/v1/node/packages/jobs/{id}
func (h *ApiNodeHandler) HandlePostNodePackagesJob(w http.ResponseWriter, r *http.Request) {
	var result api.PackageJobResult

	// get the node ID
	jobid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		responses.ErrInvalidJobID(w, r, err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		responses.ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	nodeid := *requests.NodeNID(r)

	// get the job from the database if it exists
	job, err := h.core.GetPackageJob(r.Context(), jobid)
	if err != nil {
		if h.core.ErrNotFound(err) {
			responses.ErrJobMissingOrExpired(w, r, errors.New("unable to find the job ID"))
		} else {
			responses.ErrServerError(w, r, err)
		}
		return
	}

	// check that the job is assigned to this node
	if job.NodeID != nodeid {
		responses.ErrJobMissingOrExpired(w, r, errors.New("the job ID is not assigned to this node"))
		return
	}

	if job.Result.Status != 0 {
		responses.ErrJobAlreadyCompleted(w, r, errors.New("job has already been completed"))
		return
	}

	err = h.core.CompletePackageJob(r.Context(), jobid, nodeid, &result)
	if err != nil {
		if h.core.ErrNotFound(err) {
			responses.ErrJobMissingOrExpired(w, r, errors.New("the jobs ID faild to be completed"))
			return
		}
		responses.ErrServiceUnavailable(w, r, err)
		return
	}

	responses.JsonResponse(w, r, http.StatusOK, nil)
}
