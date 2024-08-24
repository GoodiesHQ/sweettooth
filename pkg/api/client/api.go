package client

import (
	"net/http"
	"strings"

	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SweetToothClient struct {
	ServerURL  string
	Registered bool
}

func NewSweetToothClient(serverUrl string) *SweetToothClient {
	return &SweetToothClient{
		ServerURL:  strings.Trim(serverUrl, "/"),
		Registered: false,
	}
}

func (client *SweetToothClient) Check() error {
	/*if !client.Registered {
		return ErrNodeNotRegistered
	}*/
	_, err := client.doRequest(&requestParams{
		method:     http.MethodGet,
		path:       "/api/v1/node/check",
		authorized: true,
	})
	if err == nil {
		client.Registered = true
	}
	return err
}

func (client *SweetToothClient) GetSchedule() (schedule.Schedule, error) {
	log.Trace().Msg("client.Schedule called")

	var sched api.Schedule

	_, err := client.doRequest(&requestParams{
		method:     http.MethodGet,
		path:       "/api/v1/node/schedule",
		authorized: true,
		target:     &sched,
	})
	if err != nil {
		return nil, err
	}

	return schedule.Schedule(sched), nil
}

func (client *SweetToothClient) GetPackages() (*api.Packages, error) {
	log.Trace().Msg("client.GetPackages called")

	var pkg api.Packages

	_, err := client.doRequest(&requestParams{
		method:     http.MethodGet,
		path:       "/api/v1/node/packages",
		authorized: true,
		target:     &pkg,
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to acquire package inventory")
		return nil, err
	}

	return &pkg, nil
}

func (client *SweetToothClient) UpdatePackages(packages *api.Packages) error {
	log.Trace().Msg("client.UpdatePackages called")

	_, err := client.doRequest(&requestParams{
		method:     http.MethodPut,
		path:       "/api/v1/node/packages",
		authorized: true,
		body:       packages,
		optionalMap: StatusMap{ // all of these responses for this endpoint mean the device is successfully registered
			http.StatusNoContent: nil,
		},
	})

	if err != nil {
		log.Debug().Msg("failed to update server's package inventory")
		return err
	}

	return nil
}

func (client *SweetToothClient) Register(registration *api.RegistrationRequest) (int, error) {
	log.Trace().Msg("client.Register called")

	// we need to track the response to get the status code
	var res http.Response

	_, err := client.doRequest(&requestParams{
		method: http.MethodPost,
		path:   "/api/v1/node/register",
		body:   registration,
		optionalMap: StatusMap{ // all of these responses for this endpoint mean the device is successfully registered
			http.StatusOK:           nil,
			http.StatusCreated:      nil,
			http.StatusConflict:     nil,
			http.StatusForbidden:    nil,
			http.StatusUnauthorized: nil,
		},
		response: &res,
	})

	if err == nil {
		client.Registered = true
		log.Debug().Msg("Successfully registered")
		return res.StatusCode, nil
	}

	return 0, err
}

func (client *SweetToothClient) GetPackageJob(jobid uuid.UUID) (*api.PackageJob, error) {
	var job api.PackageJob

	_, err := client.doRequest(&requestParams{
		method:     http.MethodGet,
		path:       "/api/v1/node/packages/jobs/" + jobid.String(),
		authorized: true,
		target:     &job,
	})

	if err != nil {
		log.Error().Msg("failed to get package jobs list")
		return nil, err
	}

	return &job, nil
}

func (client *SweetToothClient) CompletePackageJob(jobid uuid.UUID, result *api.PackageJobResult) error {
	_, err := client.doRequest(&requestParams{
		method:     http.MethodPost,
		path:       "/api/v1/node/packages/jobs/" + jobid.String(),
		authorized: true,
		body:       result,
		optionalMap: StatusMap{
			http.StatusOK:       nil, // successful
			http.StatusConflict: nil, // ignore conflicts for now
		},
	})

	return err
}

func (client *SweetToothClient) GetPackageJobs() (api.PackageJobList, error) {
	var jobs api.PackageJobList

	_, err := client.doRequest(&requestParams{
		method:     http.MethodGet,
		path:       "/api/v1/node/packages/jobs",
		authorized: true,
		target:     &jobs,
	})

	if err != nil {
		log.Panic().Msg("failed to get package jobs list")
		return nil, err
	}

	return jobs, nil
}
