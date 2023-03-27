package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
)

type BitriseClient struct {
	logger     log.Logger
	httpClient *http.Client
	url        string
	authToken  string
}

func NewBitriseClient(appURL, buildSLUG, authToken string, logger log.Logger) BitriseClient {
	httpClient := retryhttp.NewClient(logger)
	url := fmt.Sprintf("%s/pipeline/workflow_builds/%s/env_vars", appURL, buildSLUG)

	return BitriseClient{
		logger:     logger,
		httpClient: httpClient.StandardClient(),
		url:        url,
		authToken:  authToken,
	}
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
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
	return fmt.Errorf("request failed with status: %d", resp.StatusCode)
}
