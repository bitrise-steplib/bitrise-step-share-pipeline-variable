package step

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-steputils/v2/secretkeys"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-steplib/bitrise-step-share-pipeline-variable/mocks"
	"github.com/stretchr/testify/require"
)

func TestEnvVarSharer_ProcessConfig(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "Simple inputs",
			envs: map[string]string{
				"variables":       "MY_ENV_KEY=my value",
				"app_url":         "https://app.bitrise.io/app/abcd",
				"build_slug":      "asdf",
				"build_api_token": "1234",
			},
			want: &Config{
				EnvVars:       []EnvVar{{Key: "MY_ENV_KEY", Value: "my value"}},
				AppURL:        "https://app.bitrise.io/app/abcd",
				BuildSlug:     "asdf",
				BuildAPIToken: "1234",
			},
			wantErr: false,
		},
		{
			name: "Multiple = inputs",
			envs: map[string]string{
				"variables":       "MY_ENV_KEY=my value that contains = so that we can have (=^･ｪ･^=))ﾉ彡☆",
				"app_url":         "https://app.bitrise.io/app/abcd",
				"build_slug":      "asdf",
				"build_api_token": "1234",
			},
			want: &Config{
				EnvVars: []EnvVar{
					{
						Key:   "MY_ENV_KEY",
						Value: "my value that contains = so that we can have (=^･ｪ･^=))ﾉ彡☆",
					},
				},
				AppURL:        "https://app.bitrise.io/app/abcd",
				BuildSlug:     "asdf",
				BuildAPIToken: "1234",
			},
			wantErr: false,
		},
		{
			name: "Existing env sharing",
			envs: map[string]string{
				"EXISTING_ENV_KEY": "existing env",
				"variables":        "EXISTING_ENV_KEY",
				"app_url":          "https://app.bitrise.io/app/abcd",
				"build_slug":       "asdf",
				"build_api_token":  "1234",
			},
			want: &Config{
				EnvVars:       []EnvVar{{Key: "EXISTING_ENV_KEY", Value: "existing env"}},
				AppURL:        "https://app.bitrise.io/app/abcd",
				BuildSlug:     "asdf",
				BuildAPIToken: "1234",
			},
			wantErr: false,
		},
		{
			name: "variables can't start with =",
			envs: map[string]string{
				"variables":       "=some_key=some_value",
				"app_url":         "https://app.bitrise.io/app/abcd",
				"build_slug":      "asdf",
				"build_api_token": "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "variables is required",
			envs: map[string]string{
				"variables":       "",
				"app_url":         "https://app.bitrise.io/app/abcd",
				"build_slug":      "asdf",
				"build_api_token": "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "app_url is required",
			envs: map[string]string{
				"variables":       "MY_ENV_KEY=my value",
				"app_url":         "",
				"build_slug":      "asdf",
				"build_api_token": "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "build_slug is required",
			envs: map[string]string{
				"variables":       "MY_ENV_KEY=my value",
				"app_url":         "https://app.bitrise.io/app/abcd",
				"build_slug":      "",
				"build_api_token": "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "build_api_token is required",
			envs: map[string]string{
				"variables":       "MY_ENV_KEY=my value",
				"app_url":         "https://app.bitrise.io/app/abcd",
				"build_slug":      "asdf",
				"build_api_token": "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envRepository := new(mocks.Repository)
			envRepository.On("Get", "BITRISE_SECRET_ENV_KEY_LIST").Return("")
			for key, value := range tt.envs {
				envRepository.On("Get", key).Return(value)
			}

			inputParser := stepconf.NewInputParser(envRepository)
			secretKeysProvider := secretkeys.NewManager()

			e := EnvVarSharer{
				logger:             log.NewLogger(),
				inputParser:        inputParser,
				envRepository:      envRepository,
				secretKeysProvider: secretKeysProvider,
			}
			got, err := e.ProcessConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvVarSharer_Run(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "API gets called",
			config: Config{
				EnvVars:       []EnvVar{{Key: "ENV_KEY", Value: "env_value", Sensitive: false}},
				BuildSlug:     "slug",
				BuildAPIToken: "token",
			},
			wantErr: false,
		},
		{
			name: "Failing response",
			config: Config{
				EnvVars:   []EnvVar{{Key: "ENV_KEY", Value: "env_value", Sensitive: false}},
				BuildSlug: "slug",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverCalled := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.config.BuildAPIToken == "" {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
				serverCalled = true
			}))
			defer server.Close()

			e := EnvVarSharer{
				logger: log.NewLogger(),
			}
			tt.config.AppURL = server.URL
			if err := e.Run(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, true, serverCalled)
		})
	}
}
