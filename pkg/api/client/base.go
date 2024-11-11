package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/client/keys"
	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/rs/zerolog/log"
)

type requestParams struct {
	method      string         // method string for http.NewRequest
	path        string         // URL path
	body        interface{}    // source body to marshal and submit as the body
	target      interface{}    // the target to store the unmarshaled response
	response    *http.Response // store the actual response if
	authorized  bool           // if true, add a JWT token to the request
	optionalMap StatusMap      // map of errors to return based on the status code
}

/*
 * - method = http method
 */
func (cli *SweetToothClient) doRequest(params *requestParams) (*http.Response, error) {
	// convert the request data to JSON
	var body []byte
	if params.body != nil {
		var err error
		body, err = json.Marshal(params.body)
		if err != nil {
			return nil, err
		}
	}

	// build the request with the base URL and the provided path
	url := cli.ServerURL + params.path

	// add URL to logs
	log := log.With().Str("url", url).Logger()

	req, err := http.NewRequest(params.method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// Set the JSON header for all requests
	req.Header.Set("Content-Type", "application/json")

	// set the authorization header if the request is authorized
	if params.authorized {
		// create a new JWT signed by the node's private key
		sig, err := keys.CreateNodeJWT()
		if err != nil {
			log.Panic().Err(err).Send() // should never happen
		}

		req.Header.Set("Authorization", "Bearer "+sig)
	}

	// craft the web client and perform the request
	web := &http.Client{}
	res, err := web.Do(req)
	if err != nil {
		log.Panic().Err(err).Send() // failure to perform the request, not an HTTP error
	}
	defer res.Body.Close()
	if params.response != nil {
		*params.response = *res
	}

	if err := cli.handleResponse(res, params.optionalMap, params.target); err != nil {
		return nil, err
	}

	return res, nil
}

func (cli *SweetToothClient) handleResponse(res *http.Response, optional StatusMap, target interface{}) error {
	// this function should only be called on successfully received responses, whatever status they may be
	var found bool
	var err error

	if res.StatusCode == http.StatusNotFound {
		cli.Registered = false
	}

	if optional != nil {
		err, found = optional[res.StatusCode]
		if found && err != nil {
			return err
		}
	}

	err, found = GeneralStatusMap[res.StatusCode]
	if found && err != nil {
		return err
	}

	if err != nil {
		var apierr api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&apierr); err != nil {
			apierr.StatusCode = res.StatusCode
			return apierr
		}

		return err
	}

	if target != nil {
		if err := json.NewDecoder(res.Body).Decode(target); err != nil {
			return err
		}
	}

	return nil
}

func (cli *SweetToothClient) NodeID() string {
	return crypto.Fingerprint(keys.GetPublicKey()).String()
}
