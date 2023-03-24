package step

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-steplib/bitrise-step-share-env-vars-between-stages/api"
)

type Input struct {
	EnvVars       string `env:"env_vars,required"`
	AppURL        string `env:"app_url,required"`
	BuildSlug     string `env:"build_slug,required"`
	BuildAPIToken string `env:"build_api_token,required"`
}

type EnvVar struct {
	Key   string
	Value string
}

type Config struct {
	EnvVars       []EnvVar
	AppURL        string
	BuildSlug     string
	BuildAPIToken string
}

func (c Config) APIEnvVars() []api.EnvVar {
	var apiEnvVars []api.EnvVar
	for _, envVar := range c.EnvVars {
		apiEnvVars = append(apiEnvVars, api.EnvVar{
			Key:   envVar.Key,
			Value: envVar.Value,
		})
	}
	return apiEnvVars
}

type EnvVarSharer struct {
	logger      log.Logger
	inputParser stepconf.InputParser
}

func NewEnvVarSharer(logger log.Logger, inputParser stepconf.InputParser) EnvVarSharer {
	return EnvVarSharer{
		logger:      logger,
		inputParser: inputParser,
	}
}

func (s EnvVarSharer) ProcessConfig() (*Config, error) {
	var input Input
	if err := s.inputParser.Parse(&input); err != nil {
		return nil, err
	}
	stepconf.Print(input)
	s.logger.Println()

	s.logger.EnableDebugLog(true)

	envVars, err := parseEnvVars(input.EnvVars)
	if err != nil {
		return nil, err
	}

	return &Config{
		EnvVars:       envVars,
		AppURL:        input.AppURL,
		BuildSlug:     input.BuildSlug,
		BuildAPIToken: input.BuildAPIToken,
	}, nil
}

func (s EnvVarSharer) Run(config Config) error {
	s.logger.Infof("Sharing %d env vars", len(config.EnvVars))

	client, err := api.NewBitriseClient(config.AppURL, config.BuildSlug, config.BuildAPIToken, s.logger)
	if err != nil {
		return err
	}
	return client.ShareEnvVars(config.APIEnvVars())
}

func parseEnvVars(s string) ([]EnvVar, error) {
	var envVars []EnvVar

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		split := strings.Split(line, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("env var should be in a format (KEY=value): %s", line)
		}

		envVars = append(envVars, EnvVar{
			Key:   split[0],
			Value: split[1],
		})
	}

	return envVars, nil
}
