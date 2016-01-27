package execution

import (
	"os"
	"strings"
	"time"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
)

var NumberOfExecutionStreams int
var checkUpdatesDuringExecution = false

type execution interface {
	start() *result.SuiteResult
	finish()
}

type executionInfo struct {
	manifest        *manifest.Manifest
	specStore       *specStore
	runner          *runner.TestRunner
	pluginHandler   *plugin.Handler
	parallelRunInfo *parallelInfo
	consoleReporter reporter.Reporter
	errMaps         *validationErrMaps
}

func ExecuteSpecs(inParallel bool, args []string) int {
	if checkUpdatesDuringExecution && config.CheckUpdates() {
		i := &install.UpdateFacade{}
		i.BufferUpdateDetails()
		defer i.PrintUpdateBuffer()
	}

	specsToExecute, conceptsDictionary := parseSpecs(args)
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	runner := startApi()
	errMap := validateSpecs(manifest, specsToExecute, runner, conceptsDictionary)
	pluginHandler := plugin.StartPlugins(manifest)
	parallelInfo := &parallelInfo{inParallel: inParallel, numberOfStreams: NumberOfExecutionStreams}
	if !parallelInfo.isValid() {
		os.Exit(1)
	}
	execution := newExecution(&executionInfo{manifest, &specStore{specs: specsToExecute}, runner, pluginHandler, parallelInfo, reporter.Current(), errMap})
	result := execution.start()
	execution.finish()
	exitCode := printExecutionStatus(result, errMap)
	return exitCode
}

func Validate(args []string) {
	specsToExecute, conceptsDictionary := parseSpecs(args)
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	runner := startApi()
	errMap := validateSpecs(manifest, specsToExecute, runner, conceptsDictionary)
	runner.Kill()
	if len(errMap.stepErrs) > 0 {
		os.Exit(1)
	}
	logger.Info("No error found.")
	os.Exit(0)
}

func parseSpecs(args []string) ([]*gauge.Specification, *gauge.ConceptDictionary) {
	conceptsDictionary, conceptParseResult := parser.CreateConceptsDictionary(false)
	parser.HandleParseResult(conceptParseResult)
	specsToExecute, _ := filter.GetSpecsToExecute(conceptsDictionary, args)
	if len(specsToExecute) == 0 {
		logger.Info("No specifications found in %s.", strings.Join(args, ", "))
		os.Exit(0)
	}
	return specsToExecute, conceptsDictionary
}

func newExecution(executionInfo *executionInfo) execution {
	if executionInfo.parallelRunInfo.inParallel {
		return newParallelExecution(executionInfo)
	}
	return newSimpleExecution(executionInfo)
}

func startApi() *runner.TestRunner {
	startChan := &runner.StartChannels{RunnerChan: make(chan *runner.TestRunner), ErrorChan: make(chan error), KillChan: make(chan bool)}
	go api.StartAPIService(0, startChan)
	select {
	case runner := <-startChan.RunnerChan:
		return runner
	case err := <-startChan.ErrorChan:
		logger.Fatalf("Failed to start gauge API: %s", err.Error())
	}
	return nil
}

type validationErrMaps struct {
	specErrs     map[*gauge.Specification][]*stepValidationError
	scenarioErrs map[*gauge.Scenario][]*stepValidationError
	stepErrs     map[*gauge.Step]*stepValidationError
}

func validateSpecs(manifest *manifest.Manifest, specsToExecute []*gauge.Specification, runner *runner.TestRunner, conceptDictionary *gauge.ConceptDictionary) *validationErrMaps {
	validator := newValidator(manifest, specsToExecute, runner, conceptDictionary)
	//TODO: validator.validate() should return validationErrMaps so that it has scenario/spec info with error(Which is currently done by fillErrors())
	validationErrors := validator.validate()
	errMap := &validationErrMaps{make(map[*gauge.Specification][]*stepValidationError), make(map[*gauge.Scenario][]*stepValidationError), make(map[*gauge.Step]*stepValidationError)}
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
		fillSpecErrors(spec, errMap, spec.TearDownSteps)

	}
}

func fillScenarioErrors(scenario *gauge.Scenario, errMap *validationErrMaps, steps []*gauge.Step) {
	for _, step := range steps {
		if step.IsConcept {
			fillScenarioErrors(scenario, errMap, step.ConceptSteps)
		}
		if err, ok := errMap.stepErrs[step]; ok {
			errMap.scenarioErrs[scenario] = append(errMap.scenarioErrs[scenario], err)
		}
	}
}

func fillSpecErrors(spec *gauge.Specification, errMap *validationErrMaps, steps []*gauge.Step) {
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
	nSkippedScenarios := len(errMap.scenarioErrs)
	nSkippedSpecs := len(errMap.specErrs)
	nExecutedSpecs := len(suiteResult.SpecResults) - nSkippedSpecs
	nFailedSpecs := suiteResult.SpecsFailedCount
	nPassedSpecs := nExecutedSpecs - nFailedSpecs

	nExecutedScenarios := 0
	nFailedScenarios := 0
	nPassedScenarios := 0
	for _, specResult := range suiteResult.SpecResults {
		nExecutedScenarios += specResult.ScenarioCount
		nFailedScenarios += specResult.ScenarioFailedCount
	}
	nExecutedScenarios -= nSkippedScenarios
	nPassedScenarios = nExecutedScenarios - nFailedScenarios

	logger.Info("Specifications:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedSpecs, nPassedSpecs, nFailedSpecs, nSkippedSpecs)
	logger.Info("Scenarios:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedScenarios, nPassedScenarios, nFailedScenarios, nSkippedScenarios)
	logger.Info("\nTotal time taken: %s", time.Millisecond*time.Duration(suiteResult.ExecutionTime))

	for _, unhandledErr := range suiteResult.UnhandledErrors {
		logger.Error(unhandledErr.Error())
	}
	exitCode := 0
	if suiteResult.IsFailed || (nSkippedSpecs+nSkippedScenarios) > 0 {
		exitCode = 1
	}
	return exitCode
}

func printValidationFailures(validationErrors validationErrors) {
	logger.Error("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			logger.Error("%s:%d: %s. %s", stepValidationError.fileName, s.LineNo, stepValidationError.message, s.LineText)
		}
	}
}
