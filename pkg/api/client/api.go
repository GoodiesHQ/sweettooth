package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/system"
)

type SweetToothClient struct {
	ServerURL string
}

func NewSweetToothClient(serverUrl string) (*SweetToothClient, error) {
	cli := SweetToothClient{
		ServerURL: strings.Trim(serverUrl, "/"),
	}
	return &cli, nil
}

func (client *SweetToothClient) doRequest(method, path string, data interface{}) (*http.Response, error) {
	// convert the data to JSON
	var body []byte = nil
	if data != nil {
		var err error
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	// build the request with the base URL and the provided path
	url := client.ServerURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// create a JWT signed by the node's private key
	sig, err := crypto.CreateNodeJWT()
	if err != nil {
		return nil, err
	}

	// Set the custom headers for the signature and the public key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sig)

	// craft the web client and perform the request
	web := &http.Client{}
	return web.Do(req)
}

func (client *SweetToothClient) CheckIn(withInfo bool) (*http.Response, error) {
	var info any = nil
	if withInfo {
		var err error
		info, err = system.GetSystemInfo()
		if err != nil {
			return nil, err
		}
	}
	return client.doRequest(http.MethodPost, "/api/v1/node/checkin", info)
}

func (client *SweetToothClient) Register(organization_id *string) (*http.Response, error) {
	var registration api.RegistrationRequest

	info, err := system.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	registration.Hostname = info.Hostname
	registration.ID = crypto.Fingerprint(crypto.GetPublicKey())
	registration.PublicKey = crypto.GetPublicKeyBase64()
	registration.Label = nil
	registration.OSKernel = info.OSInfo.Kernel
	registration.OSName = info.OSInfo.Name
	registration.OSMajor = info.OSInfo.Major
	registration.OSMinor = info.OSInfo.Minor
	registration.OSBuild = info.OSInfo.Build

	return client.doRequest(http.MethodPost, "/api/v1/node/register", registration)
}
