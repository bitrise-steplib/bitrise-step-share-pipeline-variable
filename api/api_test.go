package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

func TestBitriseClient_ShareEnvVars_SuccessfulRequest(t *testing.T) {
	buildSlug := "slug"
	apiToken := "token"
	envVar1 := SharedEnvVar{
		Key:       "KEY",
		Value:     "value",
		Sensitive: false,
	}
	envVar2 := SharedEnvVar{
		Key:       "SECRET_KEY",
		Value:     "secret value",
		Sensitive: true,
	}
	envVars := []SharedEnvVar{envVar1, envVar2}

	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true

		if r.URL.Path != "/pipeline/workflow_builds/slug/env_vars" {
			t.Errorf("Expected to request '/pipeline/workflow_builds/slug/env_vars', got: %s", r.URL.Path)
		}

		if r.Header.Get("content-type") != "application/json; charset=UTF-8" {
			t.Errorf("Expected content-type: application/json; charset=UTF-8 header, got: %s", r.Header.Get("content-type"))
		}

		if r.Header.Get("X-HTTP_BUILD_API_TOKEN") != apiToken {
			t.Errorf("Expected X-HTTP_BUILD_API_TOKEN: %s header, got: %s", apiToken, r.Header.Get("X-HTTP_BUILD_API_TOKEN"))
		}

		var body struct {
			Envs []struct {
				Key       string `json:"key"`
				Value     string `json:"value"`
				Sensitive bool   `json:"is_sensitive"`
			} `json:"shared_envs"`
		}

		requestBody, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		err = json.Unmarshal(requestBody, &body)
		require.NoError(t, err)

		require.Equal(t, 2, len(body.Envs))
		require.Equal(t, envVar1.Key, body.Envs[0].Key)
		require.Equal(t, envVar1.Value, body.Envs[0].Value)
		require.Equal(t, envVar1.Sensitive, body.Envs[0].Sensitive)
		require.Equal(t, envVar2.Key, body.Envs[1].Key)
		require.Equal(t, envVar2.Value, body.Envs[1].Value)
		require.Equal(t, envVar2.Sensitive, body.Envs[1].Sensitive)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewBitriseClient(server.URL, buildSlug, apiToken, log.NewLogger())
	err := c.ShareEnvVars(envVars)
	require.NoError(t, err)
	require.Equal(t, true, serverCalled)
}

func TestBitriseClient_ShareEnvVars_FailingRequest(t *testing.T) {
	buildSlug := "slug"
	apiToken := "token"
	envVarKey := "KEY"
	envVarValue := "value"
	envVars := []SharedEnvVar{{Key: envVarKey, Value: envVarValue}}

	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error_msg":"some error"}`))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	c := NewBitriseClient(server.URL, buildSlug, apiToken, log.NewLogger())
	err := c.ShareEnvVars(envVars)
	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("request to %s/pipeline/workflow_builds/slug/env_vars failed: status code should be 2xx (400), message: some error", server.URL), err.Error())
	require.Equal(t, true, serverCalled)
}
