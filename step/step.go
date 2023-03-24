package step

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
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
	logger        log.Logger
	inputParser   stepconf.InputParser
	envRepository env.Repository
}

func NewEnvVarSharer(logger log.Logger, inputParser stepconf.InputParser, envRepository env.Repository) EnvVarSharer {
	return EnvVarSharer{
		logger:        logger,
		inputParser:   inputParser,
		envRepository: envRepository,
	}
}

func (e EnvVarSharer) ProcessConfig() (*Config, error) {
	var input Input
	if err := e.inputParser.Parse(&input); err != nil {
		return nil, err
	}

	stepconf.Print(input)
	e.logger.Println()

	envVars, err := e.parseEnvVars(input.EnvVars)
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

func (e EnvVarSharer) Run(config Config) error {
	e.logger.Infof("Sharing %d env vars", len(config.EnvVars))

	client, err := api.NewBitriseClient(config.AppURL, config.BuildSlug, config.BuildAPIToken, e.logger)
	if err != nil {
		return err
	}

	if err := client.ShareEnvVars(config.APIEnvVars()); err != nil {
		return err
	}

	e.logger.Donef("Finished")

	return nil
}

func (e EnvVarSharer) parseEnvVars(s string) ([]EnvVar, error) {
	var envVars []EnvVar

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		split := strings.Split(line, "=")
		if len(split) > 2 || len(split) == 0 {
			return nil, fmt.Errorf("env var should be in a format: KEY=value or KEY: %s", line)
		}

		key := split[0]

		var value string
		if len(split) == 1 {
			value = e.envRepository.Get(key)
		} else {
			value = split[1]
		}

		envVars = append(envVars, EnvVar{
			Key:   key,
			Value: value,
		})
	}

	return envVars, nil
}
