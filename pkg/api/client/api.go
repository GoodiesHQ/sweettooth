package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/util"
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
	if !client.Registered {
		return nil
	}
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

func (client *SweetToothClient) Schedule() (schedule.Schedule, error) {
	var sched any

	_, err := client.doRequest(&requestParams{
		method:     http.MethodGet,
		path:       "/api/v1/node/schedule",
		authorized: true,
		target:     &sched,
	})
	if err != nil {
		return nil, err
	}
	fmt.Println(util.Dumps(sched))

	return nil, nil
}

func (client *SweetToothClient) UpdatePackages(packages *api.Packages) error {
	var res http.Response

	_, err := client.doRequest(&requestParams{
		method: http.MethodPost,
		path:   "/api/v1/node/register",
		body:   packages,
		optionalMap: StatusMap{ // all of these responses for this endpoint mean the device is successfully registered
			http.StatusNoContent: nil,
		},
		response: &res,
	})

	if err != nil {
		log.Debug().Msg("Successfully updated package inventory")
	}

	return err
}

func (client *SweetToothClient) Register(registration *api.RegistrationRequest) (int, error) {
	var res http.Response

	_, err := client.doRequest(&requestParams{
		method: http.MethodPost,
		path:   "/api/v1/node/register",
		body:   registration,
		optionalMap: StatusMap{ // all of these responses for this endpoint mean the device is successfully registered
			http.StatusOK:        nil,
			http.StatusCreated:   nil,
			http.StatusConflict:  nil,
			http.StatusForbidden: nil,
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
