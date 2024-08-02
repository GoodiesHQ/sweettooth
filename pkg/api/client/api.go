package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/goodieshq/sweettooth/crypto"
	"github.com/goodieshq/sweettooth/system"
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

func (client *SweetToothClient) doRequest(method, path string, body []byte) (*http.Response, error) {
	// build the request with the base URL and the provided path
	url := client.ServerURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Create the signature for the body and the current timestamp in RFC3339 format
	now := time.Now()
	timestamp := now.Format(time.RFC3339)
	sigReq := crypto.SignBase64(append(body, []byte(";"+timestamp)...))

	// Set the custom headers for the signature and the public key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", sigReq)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-PublicKey", crypto.GetPublicKeyBase64())

	// craft the web client and perform the request
	web := &http.Client{}
	res, err := web.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// read body into a buffer
	bodyRes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// retreive the server's response signature
	sigRes := res.Header.Get("X-Signature")
	if sigRes == "" {
		return nil, errors.New("missing signature in response")
	}

	// verify that the response is signed
	b, err := crypto.VerifyServerBase64(bodyRes, sigRes)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New("server key failed to verify")
	}

	return res, nil
}

func (api *SweetToothClient) CheckIn(withInfo bool) (*http.Response, error) {
	var body []byte = nil
	if withInfo {
		info, err := system.GetSystemInfo()
		if err != nil {
			return nil, err
		}
		body, err = json.Marshal(info)
		if err != nil {
			return nil, err
		}
	}
	return api.doRequest(http.MethodPost, "/checkin", body)
}
