package step

import (
	"fmt"
	"strings"

	"github.com/bitrise-steplib/bitrise-step-share-pipeline-variable/api"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

type Input struct {
	EnvVars       string `env:"variables,required"`
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

func (c Config) APIEnvVars() []api.SharedEnvVar {
	var apiEnvVars []api.SharedEnvVar
	for _, envVar := range c.EnvVars {
		apiEnvVars = append(apiEnvVars, api.SharedEnvVar{
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

	client := api.NewBitriseClient(config.AppURL, config.BuildSlug, config.BuildAPIToken, e.logger)
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
		line = strings.TrimSpace(line)
		if line == "" { 
			// empty line is ignored 
			continue 
		}
		
		key, value, _ := strings.Cut(line, "=")
		if key == "" {
			// line starting with = is invalid
			return nil, fmt.Errorf("env var should be in a format: KEY=value or KEY: %s", line)
		}
		if value == "" {
			value = e.envRepository.Get(key)
		}

		envVars = append(envVars, EnvVar{
			Key:   key,
			Value: value,
		})
	}

	return envVars, nil
}
