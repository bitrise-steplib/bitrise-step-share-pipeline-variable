package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/errorutil"
	. "github.com/bitrise-io/go-utils/v2/exitcode"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-steplib/bitrise-step-share-pipeline-variable/step"
)

func main() {
	exitCode := run()
	os.Exit(int(exitCode))
}

func run() ExitCode {
	logger := log.NewLogger()
	envVarSharer := createEnvVarSharer(logger)

	config, err := envVarSharer.ProcessConfig()
	if err != nil {
		logger.Println()
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to process Step inputs: %w", err)))
		return Failure
	}

	if err := envVarSharer.Run(*config); err != nil {
		logger.Println()
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to execute Step: %w", err)))
		return Failure
	}

	return Success
}

func createEnvVarSharer(logger log.Logger) step.EnvVarSharer {
	osEnvs := env.NewRepository()
	inputParser := stepconf.NewInputParser(osEnvs)
	envRepository := env.NewRepository()

	return step.NewEnvVarSharer(logger, inputParser, envRepository)
}
