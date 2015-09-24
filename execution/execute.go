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
	specsToExecute, _ := filter.GetSpecsToExecute(conceptsDictionary, args)
	if len(specsToExecute) == 0 {
		printExecutionStatus(nil, &validationErrMaps{})
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
	errMap := validateSpecs(manifest, specsToExecute, runner, conceptsDictionary)
	pluginHandler := plugin.StartPlugins(manifest)
	execution := newExecution(&executionInfo{manifest, specsToExecute, runner, pluginHandler, parallelInfo, &logger.Log, errMap})
	result := execution.start()
	execution.finish()
	exitCode := printExecutionStatus(result, errMap)
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

type validationErrMaps struct {
	specErrs     map[*parser.Specification][]*stepValidationError
	scenarioErrs map[*parser.Scenario][]*stepValidationError
	stepErrs     map[*parser.Step]*stepValidationError
}

func validateSpecs(manifest *manifest.Manifest, specsToExecute []*parser.Specification, runner *runner.TestRunner, conceptDictionary *parser.ConceptDictionary) *validationErrMaps {
	validator := newValidator(manifest, specsToExecute, runner, conceptDictionary)
	//TODO: validator.validate() should return validationErrMaps so that it has scenario/spec info with error(Which is currently done by fillErrors())
	validationErrors := validator.validate()
	errMap := &validationErrMaps{make(map[*parser.Specification][]*stepValidationError), make(map[*parser.Scenario][]*stepValidationError), make(map[*parser.Step]*stepValidationError)}
	if len(validationErrors) > 0 {
		printValidationFailures(validationErrors)
		fillErrors(errMap, validationErrors)
	}
	return errMap
}

func fillErrors(errMap *validationErrMaps, validationErrors validationErrors) {
	for spec, errors := range validationErrors {
		for _, err := range errors {
			errMap.stepErrs[err.step] = err
		}
		for _, scenario := range spec.Scenarios {
			fillScenarioErrors(scenario, errMap, scenario.Steps)
		}
		fillSpecErrors(spec, errMap, spec.Contexts)

	}
}

func fillScenarioErrors(scenario *parser.Scenario, errMap *validationErrMaps, steps []*parser.Step) {
	for _, step := range steps {
		if step.IsConcept {
			fillScenarioErrors(scenario, errMap, step.ConceptSteps)
		}
		if err, ok := errMap.stepErrs[step]; ok {
			errMap.scenarioErrs[scenario] = append(errMap.scenarioErrs[scenario], err)
		}
	}
}

func fillSpecErrors(spec *parser.Specification, errMap *validationErrMaps, steps []*parser.Step) {
	for _, context := range steps {
		if context.IsConcept {
			fillSpecErrors(spec, errMap, context.ConceptSteps)
		}
		if err, ok := errMap.stepErrs[context]; ok {
			errMap.specErrs[spec] = append(errMap.specErrs[spec], err)
			for _, scenario := range spec.Scenarios {
				if _, ok := errMap.scenarioErrs[scenario]; !ok {
					errMap.scenarioErrs[scenario] = append(errMap.scenarioErrs[scenario], err)
				}
			}
		}
	}
}

func printExecutionStatus(suiteResult *result.SuiteResult, errMap *validationErrMaps) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view
	if suiteResult == nil {
		logger.Log.Info("No specifications found.")
		os.Exit(0)
	}
	pendingScenarios := len(errMap.scenarioErrs)
	pendingSpecs := len(errMap.specErrs)
	specsExecCount := len(suiteResult.SpecResults) - pendingSpecs
	specsFailedCount := suiteResult.SpecsFailedCount
	specsPassedCount := specsExecCount - specsFailedCount

	scenarioExecCount := 0
	scenarioFailedCount := 0
	scenarioPassedCount := 0

	exitCode := 0
	if suiteResult.IsFailed || suiteResult.SpecsSkippedCount > 0 {
		exitCode = 1
	}
	for _, specResult := range suiteResult.SpecResults {
		scenarioExecCount += specResult.ScenarioCount
		scenarioFailedCount += specResult.ScenarioFailedCount
	}
	scenarioExecCount -= pendingScenarios
	scenarioPassedCount = scenarioExecCount - scenarioFailedCount

	logger.Log.Info("Specifications: \t%d executed, %d passed, %d failed, %d skipped", specsExecCount, specsPassedCount, specsFailedCount, pendingSpecs)
	logger.Log.Info("Scenarios: \t%d executed, %d passed, %d failed, %d skipped", scenarioExecCount, scenarioPassedCount, scenarioFailedCount, pendingScenarios)
	logger.Log.Info("Total time taken: %s", time.Millisecond*time.Duration(suiteResult.ExecutionTime))

	for _, unhandledErr := range suiteResult.UnhandledErrors {
		logger.Log.Error(unhandledErr.Error())
	}
	return exitCode
}

func printValidationFailures(validationErrors validationErrors) {
	logger.Log.Error("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			logger.Log.Error("%s:%d: %s. %s", stepValidationError.fileName, s.LineNo, stepValidationError.message, s.LineText)
		}
	}
}
