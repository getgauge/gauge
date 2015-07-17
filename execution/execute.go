package execution

import (
	"errors"
	"fmt"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/logger/execLogger"
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
		execLogger.CriticalError(err)
	}
	runner := startApi()
	validateSpecs(manifest, specsToExecute, runner, conceptsDictionary)
	pluginHandler := plugin.StartPlugins(manifest)
	execution := newExecution(manifest, specsToExecute, runner, pluginHandler, parallelInfo, execLogger.Current())
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
		execLogger.CriticalError(errors.New(fmt.Sprintf("Failed to start gauge API. %s\n", err.Error())))
	}
	return nil
}

func validateSpecs(manifest *manifest.Manifest, specsToExecute []*parser.Specification, runner *runner.TestRunner, conceptDictionary *parser.ConceptDictionary) {
	validator := newValidator(manifest, specsToExecute, runner, conceptDictionary)
	validationErrors := validator.validate()
	if len(validationErrors) > 0 {
		printValidationFailures(validationErrors)
		runner.Kill(execLogger.Current())
		os.Exit(1)
	}
}

func printExecutionStatus(suiteResult *result.SuiteResult, specsSkipped int) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view
	if suiteResult == nil {
		logger.Log.Info("No specifications found.")
		os.Exit(0)
	}
	noOfSpecificationsExecuted := len(suiteResult.SpecResults)
	noOfScenariosExecuted := 0
	noOfSpecificationsFailed := suiteResult.SpecsFailedCount
	noOfScenariosFailed := 0
	exitCode := 0
	if suiteResult.IsFailed {
		logger.Log.Info("\nThe following failures occured:\n")
		exitCode = 1
	}

	printHookError(suiteResult.PreSuite)

	for _, specResult := range suiteResult.SpecResults {
		noOfScenariosExecuted += specResult.ScenarioCount
		noOfScenariosFailed += specResult.ScenarioFailedCount
		printSpecFailure(specResult)
	}

	printHookError(suiteResult.PostSuite)

	for _, unhandledErr := range suiteResult.UnhandledErrors {
		specsSkipped += (unhandledErr).(streamExecError).numberOfSpecsSkipped()
	}
	logger.Log.Info("%d scenarios executed, %d failed\n", noOfScenariosExecuted, noOfScenariosFailed)
	logger.Log.Info("%d specifications executed, %d failed\n", noOfSpecificationsExecuted, noOfSpecificationsFailed)
	logger.Log.Info("%d specifications skipped\n", specsSkipped)
	logger.Log.Info("%s\n", time.Millisecond*time.Duration(suiteResult.ExecutionTime))
	for _, unhandledErr := range suiteResult.UnhandledErrors {
		logger.Log.Error(unhandledErr.Error())
	}
	return exitCode
}

func printSpecFailure(specResult *result.SpecResult) {
	if specResult.IsFailed {
		execLogger.Current().PrintError(fmt.Sprintf("%s : %s \n", specResult.ProtoSpec.GetFileName(), specResult.ProtoSpec.GetSpecHeading()))
		printHookError(specResult.ProtoSpec.GetPreHookFailure())

		for _, specItem := range specResult.ProtoSpec.Items {
			if specItem.GetItemType() == gauge_messages.ProtoItem_Scenario {
				printScenarioFailure(specItem.GetScenario())
			} else if specItem.GetItemType() == gauge_messages.ProtoItem_TableDrivenScenario {
				printTableDrivenScenarioFailure(specItem.GetTableDrivenScenario())
			}
		}

		printHookError(specResult.ProtoSpec.GetPostHookFailure())
	}
}

func printTableDrivenScenarioFailure(tableDrivenScenario *gauge_messages.ProtoTableDrivenScenario) {
	for _, scenario := range tableDrivenScenario.GetScenarios() {
		printScenarioFailure(scenario)
	}
}

func printScenarioFailure(scenario *gauge_messages.ProtoScenario) {
	if scenario.GetFailed() {
		execLogger.Current().PrintError(fmt.Sprintf(" %s: \n", scenario.GetScenarioHeading()))
		printHookError(scenario.GetPreHookFailure())

		for _, scenarioItem := range scenario.GetScenarioItems() {
			if scenarioItem.GetItemType() == gauge_messages.ProtoItem_Step {
				printStepFailure(scenarioItem.GetStep())
			} else if scenarioItem.GetItemType() == gauge_messages.ProtoItem_Concept {
				printConceptFailure(scenarioItem.GetConcept())
			}
		}
		printHookError(scenario.GetPostHookFailure())
	}

}

func printStepFailure(step *gauge_messages.ProtoStep) {
	stepExecResult := step.StepExecutionResult
	if stepExecResult != nil && stepExecResult.ExecutionResult.GetFailed() {
		execLogger.Current().PrintError(fmt.Sprintf("\t %s\n", step.GetActualText()))
		printHookError(stepExecResult.GetPreHookFailure())
		printError(stepExecResult.ExecutionResult)
		printHookError(stepExecResult.GetPostHookFailure())
	}
}

func printConceptFailure(concept *gauge_messages.ProtoConcept) {
	conceptExecResult := concept.ConceptExecutionResult
	if conceptExecResult != nil && conceptExecResult.GetExecutionResult().GetFailed() {
		execLogger.Current().PrintError(fmt.Sprintf("\t %s\n", concept.ConceptStep.GetActualText()))
		printError(conceptExecResult.ExecutionResult)
	}
}

func printError(execResult *gauge_messages.ProtoExecutionResult) {
	if execResult.GetFailed() {
		console := execLogger.Current()
		console.PrintError(execResult.GetErrorMessage() + "\n")
		console.PrintError(execResult.GetStackTrace() + "\n")
	}
}

func printHookError(hook *gauge_messages.ProtoHookFailure) {
	if hook != nil {
		console := execLogger.Current()
		console.PrintError(hook.GetErrorMessage())
		console.PrintError(hook.GetStackTrace())
	}
}

func printValidationFailures(validationErrors executionValidationErrors) {
	logger.Log.Warning("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			execLogger.Current().PrintError(fmt.Sprintf("%s:%d: %s. %s\n", stepValidationError.fileName, s.LineNo, stepValidationError.message, s.LineText))
		}
	}
}
