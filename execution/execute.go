package execution

import (
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"os"
	"time"
)

var NumberOfExecutionStreams int

func ExecuteSpecs(inParallel bool, args []string) {
	env.LoadEnv(false)
	conceptsDictionary, conceptParseResult := parser.CreateConceptsDictionary(false)
	parser.HandleParseResult(conceptParseResult)
	specsToExecute, specsSkipped := filter.GetSpecsToExecute(conceptsDictionary, args)
	if len(specsToExecute) == 0 {
		printExecutionStatus(nil, 0)
	}
	parallelInfo := &parallelInfo{inParallel: inParallel, numberOfStreams: NumberOfExecutionStreams}
	if !parallelInfo.isValid() {
		os.Exit(1)
	}
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Log.Critical(err.Error())
	}
	runner := startApi()
	validateSpecs(manifest, specsToExecute, runner, conceptsDictionary)
	pluginHandler := plugin.StartPlugins(manifest)
	execution := newExecution(manifest, specsToExecute, runner, pluginHandler, parallelInfo, &logger.Log)
	result := execution.start()
	execution.finish()
	exitCode := printExecutionStatus(result, specsSkipped)
	os.Exit(exitCode)
}

func startApi() *runner.TestRunner {
	startChan := &runner.StartChannels{RunnerChan: make(chan *runner.TestRunner), ErrorChan: make(chan error), KillChan: make(chan bool)}
	go api.StartAPIService(0, startChan)
	select {
	case runner := <-startChan.RunnerChan:
		return runner
	case err := <-startChan.ErrorChan:
		logger.Log.Critical("Failed to start gauge API: %s", err.Error())
		os.Exit(1)
	}
	return nil
}

func validateSpecs(manifest *manifest.Manifest, specsToExecute []*parser.Specification, runner *runner.TestRunner, conceptDictionary *parser.ConceptDictionary) {
	validator := newValidator(manifest, specsToExecute, runner, conceptDictionary)
	validationErrors := validator.validate()
	if len(validationErrors) > 0 {
		printValidationFailures(validationErrors)
		runner.Kill()
		os.Exit(1)
	}
}

func printExecutionStatus(suiteResult *result.SuiteResult, specsSkippedCount int) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view
	if suiteResult == nil {
		logger.Log.Info("No specifications found.")
		os.Exit(0)
	}

	specsExecCount := len(suiteResult.SpecResults)
	specsFailedCount := suiteResult.SpecsFailedCount
	specsPassedCount := specsExecCount - specsFailedCount

	scenarioExecCount := 0
	scenarioFailedCount := 0
	scenarioPassedCount := 0

	exitCode := 0
	if suiteResult.IsFailed {
		exitCode = 1
	}

	for _, specResult := range suiteResult.SpecResults {
		scenarioExecCount += specResult.ScenarioCount
		scenarioFailedCount += specResult.ScenarioFailedCount
	}

	scenarioPassedCount = scenarioExecCount - scenarioFailedCount

	for _, unhandledErr := range suiteResult.UnhandledErrors {
		specsSkippedCount += (unhandledErr).(streamExecError).numberOfSpecsSkipped()
	}

	logger.Log.Info("Specifications: \t%d executed, %d passed, %d failed, %d skipped", specsExecCount, specsPassedCount, specsFailedCount, specsSkippedCount)
	logger.Log.Info("Scenarios: \t%d executed, %d passed, %d failed", scenarioExecCount, scenarioPassedCount, scenarioFailedCount)
	logger.Log.Info("Total time taken: %s", time.Millisecond*time.Duration(suiteResult.ExecutionTime))

	for _, unhandledErr := range suiteResult.UnhandledErrors {
		logger.Log.Error(unhandledErr.Error())
	}
	return exitCode
}

func printValidationFailures(validationErrors executionValidationErrors) {
	logger.Log.Warning("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			logger.Log.Error("%s:%d: %s. %s", stepValidationError.fileName, s.LineNo, stepValidationError.message, s.LineText)
		}
	}
}
