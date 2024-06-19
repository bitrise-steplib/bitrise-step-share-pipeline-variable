package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type SharedEnvVar struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Sensitive bool   `json:"is_sensitive"`
}

type ShareEnvVarsRequest struct {
	SharedEnvs []SharedEnvVar `json:"shared_envs"`
}

func (c BitriseClient) ShareEnvVars(envVars []SharedEnvVar) error {
	shareEnvVarsReq := ShareEnvVarsRequest{SharedEnvs: envVars}

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
	respErr := fmt.Errorf("request to %s failed: status code should be 2xx (%d)", resp.Request.URL, resp.StatusCode)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return respErr
	}
	var respBodyJSON map[string]string
	err = json.Unmarshal(respBody, &respBodyJSON)
	if err != nil {
		return respErr
	}
	clientRespErrMsg, isSet := respBodyJSON["error_msg"]
	if !isSet || respBodyJSON["error_msg"] == "" {
		return respErr
	}
	respErr = errors.New(respErr.Error() + fmt.Sprintf(", message: %s", clientRespErrMsg))

	return respErr
}
