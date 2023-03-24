package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
)

type BitriseClient struct {
	logger     log.Logger
	httpClient *http.Client
	url        string
	authToken  string
}

func NewBitriseClient(appURL, buildSLUG, authToken string, logger log.Logger) (*BitriseClient, error) {
	httpClient := retryhttp.NewClient(logger)
	url := fmt.Sprintf("%s/pipeline/workflow_builds/%s/env_vars", appURL, buildSLUG)

	return &BitriseClient{
		logger:     logger,
		httpClient: httpClient.StandardClient(),
		url:        url,
		authToken:  authToken,
	}, nil
}

type EnvVar struct {
	Key   string
	Value string
}

type SharedEnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ShareEnvVarsRequest struct {
	SharedEnvs []SharedEnvVar `json:"shared_envs"`
}

/*
curl --show-error --fail -X POST \
    "https://app.bitrise.io/app/$BITRISE_APP_SLUG/pipeline/workflow_builds/$BITRISE_BUILD_SLUG/env_vars" \
    -H 'content-type: application/json; charset=UTF-8' \
    -H "X-HTTP_BUILD_API_TOKEN: $BITRISE_BUILD_API_TOKEN" \
    -d '{"shared_envs":[{"key":"WAY_OF_KINGS","value":"Life before Death...","is_expand":false}]}'
*/

func (c BitriseClient) ShareEnvVars(envVars []EnvVar) error {
	shareEnvVarsReq := ShareEnvVarsRequest{}
	for _, envVar := range envVars {
		shareEnvVarsReq.SharedEnvs = append(shareEnvVarsReq.SharedEnvs, SharedEnvVar{
			Key:   envVar.Key,
			Value: envVar.Value,
		})
	}

	body, err := json.Marshal(shareEnvVarsReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json; charset=UTF-8")
	req.Header.Set("X-HTTP_BUILD_API_TOKEN", c.authToken)

	reqDump, err := httputil.DumpRequest(req, true)
	if err == nil {
		c.logger.Debugf("Request: %s", string(reqDump))
		c.logger.Debugf("")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err == nil {
		c.logger.Debugf("Response: %s", string(respDump))
		c.logger.Debugf("")
	}

	if err := checkEnvVarShareResponse(resp); err != nil {
		return err
	}

	return nil
}

func checkEnvVarShareResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("unsuccessful status: %d", resp.StatusCode)
}
