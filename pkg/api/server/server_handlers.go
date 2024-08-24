package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
)

const (
	DEFAULT_ATTEMPTS_MAX int = 5
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

	// check node existince
	node, err := srv.core.GetNode(r.Context(), fprint)
	if err != nil {
		ErrServerError(w, r, err)
		return
	}

	if node != nil {
		// node is already registered, do not continue
		if node.Approved {
			JsonResponse(w, r, http.StatusNoContent, nil)
			return
		}
		// node is not approved, should 403
		ErrNodeNotApproved(w, r, nil)
		return
	}

	// get the registration token and verify it is valid
	orgid, err := srv.core.ProcessRegistrationToken(r.Context(), req.Token)
	if err != nil {
		if srv.core.ErrNotFound(err) {
			ErrRegistrationTokenInvalid(w, r, err)
			return
		}
		ErrServiceUnavailable(w, r, err)
		return
	}

	if orgid == nil {
		ErrRegistrationTokenInvalid(w, r, nil)
		return
	}

	// node does not exist and registration token does point to a valid organization
	_, err = srv.core.CreateNode(r.Context(), req)
	if err != nil {
		ErrServerError(w, r, err)
		return
	}

	JsonResponse(w, r, http.StatusCreated, nil)
}

// GET /api/v1/node/packages
func (srv *SweetToothServer) handleGetNodePackages(w http.ResponseWriter, r *http.Request) {
	// get a list of currently installed packages (choco, system, outdated)
	pkg, err := srv.core.GetNodePackages(r.Context(), *util.Rid(r))
	if err != nil {
		ErrServiceUnavailable(w, r, err)
		return
	}

	if pkg == nil {
		ErrNodeNotFound(w, r, errors.New("couldnt acquire node package"))
	}

	// return the existing packages
	JsonResponse(w, r, http.StatusOK, pkg)
}

// PUT /api/v1/node/packages
func (srv *SweetToothServer) handlePutNodePackages(w http.ResponseWriter, r *http.Request) {
	var packages api.Packages

	// decode the submitted packages
	err := json.NewDecoder(r.Body).Decode(&packages)
	if err != nil {
		ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	// update the node's packages in the database if they are different
	err = srv.core.UpdateNodePackages(r.Context(), *util.Rid(r), &packages)
	if err != nil {
		ErrServiceUnavailable(w, r, err)
		return
	}

	// always return with a 204 if everything was successful, regardless of if packages were updated or not
	JsonResponse(w, r, http.StatusNoContent, nil)
}

// GET /api/v1/node/packages/jobs
func (srv *SweetToothServer) handleGetNodePackagesJobs(w http.ResponseWriter, r *http.Request) {
	// get the node ID from the request's JWT
	nodeid := *util.Rid(r)
	attemptsMax := util.RequestQueryAttemptsMax(r)

	// get job list from the database
	jobs, err := srv.core.GetPackageJobList(r.Context(), nodeid, attemptsMax)
	if err != nil {
		ErrServiceUnavailable(w, r, errors.New("failed to get the job list"))
		return
	}

	// return the jobs to the client
	JsonResponse(w, r, http.StatusOK, jobs)
}

// GET /api/v1/node/packages/jobs/{id}
func (srv *SweetToothServer) handleGetNodePackagesJob(w http.ResponseWriter, r *http.Request) {
	// get the node ID
	jobid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		ErrInvalidJobID(w, r, err)
		return
	}

	nodeid := *util.Rid(r)
	attemptsMax := util.RequestQueryAttemptsMax(r)

	// get the job from the database if it exists
	job, err := srv.core.GetPackageJob(r.Context(), jobid)
	if err != nil {
		if srv.core.ErrNotFound(err) {
			ErrJobMissingOrExpired(w, r, errors.New("unable to find the job ID"))
		} else {
			ErrServerError(w, r, err)
		}
		return
	}

	// check that the job is assigned to this node
	if job.NodeID != nodeid {
		ErrJobMissingOrExpired(w, r, errors.New("the job ID is not assigned to this node"))
		return
	}

	job, err = srv.core.AttemptPackageJob(r.Context(), jobid, nodeid, attemptsMax)
	if err != nil {
		ErrServerError(w, r, err)
		return
	}

	if job == nil {
		ErrJobMissingOrExpired(w, r, errors.New("no job ID found for this node"))
		return
	}

	JsonResponse(w, r, http.StatusOK, job)
}

// POST /api/v1/node/packages/jobs/{id}
func (srv *SweetToothServer) handlePostNodePackagesJob(w http.ResponseWriter, r *http.Request) {
	var result api.PackageJobResult

	// get the node ID
	jobid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		ErrInvalidJobID(w, r, err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		ErrInvalidRequestBody(w, r, err) // improper form submission
		return
	}

	nodeid := *util.Rid(r)

	// get the job from the database if it exists
	job, err := srv.core.GetPackageJob(r.Context(), jobid)
	if err != nil {
		if srv.core.ErrNotFound(err) {
			ErrJobMissingOrExpired(w, r, errors.New("unable to find the job ID"))
		} else {
			ErrServerError(w, r, err)
		}
		return
	}

	// check that the job is assigned to this node
	if job.NodeID != nodeid {
		ErrJobMissingOrExpired(w, r, errors.New("the job ID is not assigned to this node"))
		return
	}

	if job.Result.Status != 0 {
		ErrJobAlreadyCompleted(w, r, errors.New("job has already been completed"))
		return
	}

	err = srv.core.CompletePackageJob(r.Context(), jobid, nodeid, &result)
	if err != nil {
		if srv.core.ErrNotFound(err) {
			ErrJobMissingOrExpired(w, r, errors.New("the jobs ID faild to be completed"))
			return
		}
		ErrServiceUnavailable(w, r, err)
		return
	}

	JsonResponse(w, r, http.StatusOK, nil)
}
