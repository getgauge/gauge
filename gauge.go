// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/version"
	flag "github.com/getgauge/mflag"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	specsDirName      = "specs"
	skelFileName      = "hello_world.spec"
	envDefaultDirName = "default"
)

var defaultPlugins = []string{"html-report"}

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "Run as a daemon")
var gaugeVersion = flag.Bool([]string{"v", "-version", "version"}, false, "Print the current version and exit. Eg: gauge --version")
var verbosity = flag.Bool([]string{"-verbose"}, false, "Enable verbose logging for debugging")
var logLevel = flag.String([]string{"-log-level"}, "", "Set level of logging to debug, info, warning, error or critical")
var simpleConsoleOutput = flag.Bool([]string{"-simple-console"}, false, "Removes colouring and simplifies from the console output")
var initialize = flag.String([]string{"-init"}, "", "Initializes project structure in the current directory. Eg: gauge --init java")
var install = flag.String([]string{"-install"}, "", "Downloads and installs a plugin. Eg: gauge --install java")
var installAll = flag.Bool([]string{"-install-all"}, false, "Installs all the plugins specified in project manifest, if not installed. Eg: gauge --install-all")
var update = flag.String([]string{"-update"}, "", "Updates a plugin. Eg: gauge --update java")
var installVersion = flag.String([]string{"-plugin-version"}, "", "Version of plugin to be installed. This is used with --install")
var installZip = flag.String([]string{"-file", "f"}, "", "Installs the plugin from zip file. This is used with --install. Eg: gauge --install java -f ZIP_FILE")
var currentEnv = flag.String([]string{"-env"}, "default", "Specifies the environment. If not specified, default will be used")
var addPlugin = flag.String([]string{"-add-plugin"}, "", "Adds the specified non-language plugin to the current project")
var pluginArgs = flag.String([]string{"-plugin-args"}, "", "Specified additional arguments to the plugin. This is used together with --add-plugin")
var specFilesToFormat = flag.String([]string{"-format"}, "", "Formats the specified spec files")
var executeTags = flag.String([]string{"-tags"}, "", "Executes the specs and scenarios tagged with given tags. Eg: gauge --tags tag1,tag2 specs")
var tableRows = flag.String([]string{"-table-rows"}, "", "Executes the specs and scenarios only for the selected rows. Eg: gauge --table-rows \"1-3\" specs/hello.spec")
var apiPort = flag.String([]string{"-api-port"}, "", "Specifies the api port to be used. Eg: gauge --daemonize --api-port 7777")
var refactor = flag.String([]string{"-refactor"}, "", "Refactor steps")
var parallel = flag.Bool([]string{"-parallel", "p"}, false, "Execute specs in parallel")
var numberOfExecutionStreams = flag.Int([]string{"n"}, numberOfCores(), "Specify number of parallel execution streams")
var distribute = flag.Int([]string{"g", "-group"}, -1, "Specify which group of specification to execute based on -n flag")
var workingDir = flag.String([]string{"-dir"}, ".", "Set the working directory for the current command, accepts a path relative to current directory.")
var doNotRandomize = flag.Bool([]string{"-sort", "s"}, false, "run specs in Alphabetical Order. Eg: gauge -s specs")

func main() {
	flag.Parse()
	setWorkingDir(*workingDir)
	validGaugeProject := true
	err := config.SetProjectRoot(flag.Args())
	if err != nil {
		validGaugeProject = false
	}
	env.LoadEnv(*currentEnv, true)
	logger.Initialize(*verbosity, *logLevel)
	if *gaugeVersion {
		printVersion()
	} else if *daemonize && validGaugeProject {
		runInBackground()
	} else if *specFilesToFormat != "" && validGaugeProject {
		formatSpecFilesIn(*specFilesToFormat)
	} else if *initialize != "" {
		initializeProject(*initialize)
	} else if *installZip != "" && *install != "" {
		installPluginZip(*installZip, *install)
	} else if *install != "" {
		downloadAndInstallPlugin(*install, *installVersion)
	} else if *installAll {
		installAllPlugins()
	} else if *update != "" {
		updatePlugin(*update)
	} else if *addPlugin != "" {
		addPluginToProject(*addPlugin)
	} else if *refactor != "" && validGaugeProject {
		refactorSteps(*refactor, newStepName())
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		} else if validGaugeProject {
			if *distribute != -1 {
				*doNotRandomize = true
			}
			executeSpecs(*parallel)
		} else {
			logger.Log.Error("Could not set project root: %s", err.Error())
			os.Exit(1)
		}
	}
}

func newStepName() string {
	if len(flag.Args()) != 1 {
		printUsage()
	}
	return flag.Args()[0]
}

func refactorSteps(oldStep, newStep string) {
	refactoringResult := performRephraseRefactoring(oldStep, newStep)
	printRefactoringSummary(refactoringResult)
}

func printRefactoringSummary(refactoringResult *refactoringResult) {
	exitCode := 0
	if !refactoringResult.success {
		exitCode = 1
		for _, err := range refactoringResult.errors {
			logger.Log.Error("%s \n", err)
		}
	}
	for _, warning := range refactoringResult.warnings {
		logger.Log.Warning("%s \n", warning)
	}
	logger.Log.Info("%d specifications changed.\n", len(refactoringResult.specsChanged))
	logger.Log.Info("%d concepts changed.\n", len(refactoringResult.conceptsChanged))
	logger.Log.Info("%d files in code changed.\n", len(refactoringResult.runnerFilesChanged))
	os.Exit(exitCode)
}

func saveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		logger.Log.Error("Failed to refactor '%s': %s\n", fileName, err)
	}
}

func printUsage() {
	fmt.Printf("gauge - version %s\n", version.CurrentGaugeVersion.String())
	fmt.Printf("Copyright %d Thoughtworks\n\n", time.Now().Year())
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	os.Exit(2)
}

func runInBackground() {
	var port int
	var err error
	if *apiPort != "" {
		port, err = strconv.Atoi(*apiPort)
		os.Setenv(common.ApiPortEnvVariableName, *apiPort)
		if err != nil {
			handleCriticalError(errors.New(fmt.Sprintf("Failed to parse the port number :", *apiPort, "\n", err.Error())))
		}
	} else {
		env.LoadEnv(*currentEnv, false)
		port, err = conn.GetPortFromEnvironmentVariable(common.ApiPortEnvVariableName)
		if err != nil {
			handleCriticalError(errors.New(fmt.Sprintf("Failed to start API Service. %s \n", err.Error())))
		}
	}
	var wg sync.WaitGroup
	runAPIServiceIndefinitely(port, &wg)
	wg.Wait()
}

func formatSpecFilesIn(filesLocation string) {
	specFiles := getSpecFiles(filesLocation)
	parseResults := formatSpecFiles(specFiles...)
	handleParseResult(parseResults...)
}

func initializeProject(language string) {
	wd, err := os.Getwd()
	if err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("Failed to find working directory. %s\n", err.Error())))
	}
	config.ProjectRoot = wd
	err = createProjectTemplate(language)
	if err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("Failed to initialize. %s\n", err.Error())))
	}
	logger.Log.Info("\nSuccessfully initialized the project. Run specifications with \"gauge specs/\"")
}

func addPluginToProject(pluginName string) {
	additionalArgs := make(map[string]string)
	if *pluginArgs != "" {
		// plugin args will be comma separated values
		// eg: version=1.0, foo_version = 2.41
		args := strings.Split(*pluginArgs, ",")
		for _, arg := range args {
			keyValuePair := strings.Split(arg, "=")
			if len(keyValuePair) == 2 {
				additionalArgs[strings.TrimSpace(keyValuePair[0])] = strings.TrimSpace(keyValuePair[1])
			}
		}
	}
	manifest, err := getProjectManifest()
	if err != nil {
		handleCriticalError(err)
	}
	if err := addPluginToTheProject(pluginName, additionalArgs, manifest); err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("Failed to add plugin %s to project : %s\n", pluginName, err.Error())))
	} else {
		logger.Log.Info("Plugin %s was successfully added to the project\n", pluginName)
	}
}

func executeSpecs(inParallel bool) {
	env.LoadEnv(*currentEnv, false)
	conceptsDictionary, conceptParseResult := createConceptsDictionary(false)
	handleParseResult(conceptParseResult)
	specsToExecute, specsSkipped := getSpecsToExecute(conceptsDictionary)
	if len(specsToExecute) == 0 {
		printExecutionStatus(nil, 0)
	}
	parallelInfo := &parallelInfo{inParallel: inParallel, numberOfStreams: *numberOfExecutionStreams}
	if !parallelInfo.isValid() {
		os.Exit(1)
	}
	manifest, err := getProjectManifest()
	if err != nil {
		handleCriticalError(err)
	}
	err, apiHandler := startAPIService(0)
	if err != nil {
		apiHandler.runner.kill(getCurrentLogger())
		handleCriticalError(errors.New(fmt.Sprintf("Failed to start gauge API. %s\n", err.Error())))
	}
	if apiHandler.runner == nil {
		handleCriticalError(errors.New("Failed to start a runner\n"))
	}

	validateSpecs(manifest, specsToExecute, apiHandler.runner, conceptsDictionary)
	pluginHandler := startPlugins(manifest)
	execution := newExecution(manifest, specsToExecute, apiHandler.runner, pluginHandler, parallelInfo, getCurrentLogger())
	result := execution.start()
	execution.finish()
	exitCode := printExecutionStatus(result, specsSkipped)
	os.Exit(exitCode)
}

func handleCriticalError(err error) {
	getCurrentLogger().Critical(err.Error())
	os.Exit(1)
}

func getDataTableRows(rowCount int) indexRange {
	if *tableRows == "" {
		return indexRange{start: 0, end: rowCount - 1}
	}
	indexes, err := getDataTableRowsRange(*tableRows, rowCount)
	if err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("Table rows validation failed. %s\n", err.Error())))
	}
	return indexes
}

func shuffleSpecs(allSpecs []*specification) []*specification {
	dest := make([]*specification, len(allSpecs))
	rand.Seed(int64(time.Now().Nanosecond()))
	perm := rand.Perm(len(allSpecs))
	for i, v := range perm {
		dest[v] = allSpecs[i]
	}
	return dest
}

func validateSpecs(manifest *manifest, specsToExecute []*specification, runner *testRunner, conceptDictionary *conceptDictionary) {
	validator := newValidator(manifest, specsToExecute, runner, conceptDictionary)
	validationErrors := validator.validate()
	if len(validationErrors) > 0 {
		printValidationFailures(validationErrors)
		runner.kill(getCurrentLogger())
		os.Exit(1)
	}
}

func getSpecsToExecute(conceptsDictionary *conceptDictionary) ([]*specification, int) {
	specsToExecute := specsFromArgs(conceptsDictionary)
	totalSpecs := specsToExecute
	specsToExecute = applyFilters(specsToExecute, specsFilters())
	return sortSpecsList(specsToExecute), len(totalSpecs) - len(specsToExecute)
}

func specsFilters() []specsFilter {
	return []specsFilter{&tagsFilter{*executeTags}, &specsGroupFilter{*distribute, *numberOfExecutionStreams}, &specRandomizer{*doNotRandomize}}
}

func applyFilters(specsToExecute []*specification, filters []specsFilter) []*specification {
	for _, specsFilter := range filters {
		specsToExecute = specsFilter.filter(specsToExecute)
	}
	return specsToExecute
}

func printValidationFailures(validationErrors executionValidationErrors) {
	logger.Log.Warning("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			getCurrentLogger().PrintError(fmt.Sprintf("%s:%d: %s. %s\n", stepValidationError.fileName, s.lineNo, stepValidationError.message, s.getLineText()))
		}
	}
}

func getSpecName(specSource string) string {
	if isIndexedSpec(specSource) {
		specSource, _ = GetIndexedSpecName(specSource)
	}
	return specSource
}

// All the environment variables loaded from the
// current environments JSON files will live here
type environmentVariables struct {
	Variables map[string]string
}

func showMessage(action, filename string) {
	logger.Log.Info(" %s  %s\n", action, filename)
}

func createProjectTemplate(language string) error {
	if !isCompatibleLanguagePluginInstalled(language) {
		logger.Log.Info("Compatible %s plugin is not installed \n", language)
		logger.Log.Info("Installing plugin => %s ... \n\n", language)

		if result := installPlugin(language, ""); !result.success {
			return errors.New(fmt.Sprintf("Failed to install plugin %s . %s \n", language, result.getMessage()))
		}
	}
	// Create the project manifest
	showMessage("create", common.ManifestFile)
	if common.FileExists(common.ManifestFile) {
		showMessage("skip", common.ManifestFile)
	}
	manifest := &manifest{Language: language, Plugins: defaultPlugins}
	if err := manifest.save(); err != nil {
		return err
	}

	// creating the spec directory
	showMessage("create", specsDirName)
	if !common.DirExists(specsDirName) {
		err := os.Mkdir(specsDirName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specsDirName, err.Error()))
		}
	} else {
		showMessage("skip", specsDirName)
	}

	// Copying the skeleton file
	skelFile, err := common.GetSkeletonFilePath(skelFileName)
	if err != nil {
		return err
	}
	specFile := path.Join(specsDirName, skelFileName)
	showMessage("create", specFile)
	if common.FileExists(specFile) {
		showMessage("skip", specFile)
	} else {
		err = common.CopyFile(skelFile, specFile)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specFile, err.Error()))
		}
	}

	// Creating the env directory
	showMessage("create", common.EnvDirectoryName)
	if !common.DirExists(common.EnvDirectoryName) {
		err = os.Mkdir(common.EnvDirectoryName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", common.EnvDirectoryName, err.Error()))
		}
	}
	defaultEnv := path.Join(common.EnvDirectoryName, envDefaultDirName)
	showMessage("create", defaultEnv)
	if !common.DirExists(defaultEnv) {
		err = os.Mkdir(defaultEnv, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultEnv, err.Error()))
		}
	}
	defaultJson, err := common.GetSkeletonFilePath(path.Join(common.EnvDirectoryName, common.DefaultEnvFileName))
	if err != nil {
		return err
	}
	defaultJsonDest := path.Join(defaultEnv, common.DefaultEnvFileName)
	showMessage("create", defaultJsonDest)
	err = common.CopyFile(defaultJson, defaultJsonDest)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultJsonDest, err.Error()))
	}

	return executeInitHookForRunner(language)
}

func handleParseResult(results ...*parseResult) {
	for _, result := range results {
		if !result.ok {
			logger.Log.Critical(result.Error())
			os.Exit(1)
		}
		if result.warnings != nil {
			for _, warning := range result.warnings {
				logger.Log.Warning("%s : %v", result.fileName, warning)
			}
		}
	}
}

func startRunnerAndMakeConnection(manifest *manifest, writer executionLogger) (*testRunner, error) {
	port, err := conn.GetPortFromEnvironmentVariable(common.GaugePortEnvName)
	if err != nil {
		port = 0
	}
	gaugeConnectionHandler, connHandlerErr := conn.NewGaugeConnectionHandler(port, nil)
	if connHandlerErr != nil {
		return nil, connHandlerErr
	}
	testRunner, err := startRunner(manifest, strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber()), writer)
	if err != nil {
		return nil, err
	}

	runnerConnection, connectionError := gaugeConnectionHandler.AcceptConnection(config.RunnerConnectionTimeout(), testRunner.errorChannel)
	testRunner.connection = runnerConnection
	if connectionError != nil {
		writer.Debug("Runner connection error: %s", connectionError)
		err := testRunner.killRunner()
		if err != nil {
			writer.Debug("Error while killing runner: %s", err)
		}
		return nil, connectionError
	}
	return testRunner, nil
}

func printExecutionStatus(suiteResult *suiteResult, specsSkipped int) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view
	if suiteResult == nil {
		logger.Log.Info("No specifications found.")
		os.Exit(0)
	}
	noOfSpecificationsExecuted := len(suiteResult.specResults)
	noOfScenariosExecuted := 0
	noOfSpecificationsFailed := suiteResult.specsFailedCount
	noOfScenariosFailed := 0
	exitCode := 0
	if suiteResult.isFailed {
		logger.Log.Info("\nThe following failures occured:\n")
		exitCode = 1
	}

	printHookError(suiteResult.preSuite)

	for _, specResult := range suiteResult.specResults {
		noOfScenariosExecuted += specResult.scenarioCount
		noOfScenariosFailed += specResult.scenarioFailedCount
		printSpecFailure(specResult)
	}

	printHookError(suiteResult.postSuite)

	for _, unhandledErr := range suiteResult.unhandledErrors {
		specsSkipped += (unhandledErr).(streamExecError).numberOfSpecsSkipped()
	}
	logger.Log.Info("%d scenarios executed, %d failed\n", noOfScenariosExecuted, noOfScenariosFailed)
	logger.Log.Info("%d specifications executed, %d failed\n", noOfSpecificationsExecuted, noOfSpecificationsFailed)
	logger.Log.Info("%d specifications skipped\n", specsSkipped)
	logger.Log.Info("%s\n", time.Millisecond*time.Duration(suiteResult.executionTime))
	for _, unhandledErr := range suiteResult.unhandledErrors {
		logger.Log.Error(unhandledErr.Error())
	}
	return exitCode
}

func printHookError(hook *(gauge_messages.ProtoHookFailure)) {
	if hook != nil {
		console := getCurrentLogger()
		console.PrintError(hook.GetErrorMessage())
		console.PrintError(hook.GetStackTrace())
	}
}

func printError(execResult *gauge_messages.ProtoExecutionResult) {
	if execResult.GetFailed() {
		console := getCurrentLogger()
		console.PrintError(execResult.GetErrorMessage() + "\n")
		console.PrintError(execResult.GetStackTrace() + "\n")
	}
}

func printSpecFailure(specResult *specResult) {
	if specResult.isFailed {
		getCurrentLogger().PrintError(fmt.Sprintf("%s : %s \n", specResult.protoSpec.GetFileName(), specResult.protoSpec.GetSpecHeading()))
		printHookError(specResult.protoSpec.GetPreHookFailure())

		for _, specItem := range specResult.protoSpec.Items {
			if specItem.GetItemType() == gauge_messages.ProtoItem_Scenario {
				printScenarioFailure(specItem.GetScenario())
			} else if specItem.GetItemType() == gauge_messages.ProtoItem_TableDrivenScenario {
				printTableDrivenScenarioFailure(specItem.GetTableDrivenScenario())
			}
		}

		printHookError(specResult.protoSpec.GetPostHookFailure())
	}
}

func printTableDrivenScenarioFailure(tableDrivenScenario *gauge_messages.ProtoTableDrivenScenario) {
	for _, scenario := range tableDrivenScenario.GetScenarios() {
		printScenarioFailure(scenario)
	}
}

func printScenarioFailure(scenario *gauge_messages.ProtoScenario) {
	if scenario.GetFailed() {
		getCurrentLogger().PrintError(fmt.Sprintf(" %s: \n", scenario.GetScenarioHeading()))
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
		getCurrentLogger().PrintError(fmt.Sprintf("\t %s\n", step.GetActualText()))
		printHookError(stepExecResult.GetPreHookFailure())
		printError(stepExecResult.ExecutionResult)
		printHookError(stepExecResult.GetPostHookFailure())
	}
}

func printConceptFailure(concept *gauge_messages.ProtoConcept) {
	conceptExecResult := concept.ConceptExecutionResult
	if conceptExecResult != nil && conceptExecResult.GetExecutionResult().GetFailed() {
		getCurrentLogger().PrintError(fmt.Sprintf("\t %s\n", concept.ConceptStep.GetActualText()))
		printError(conceptExecResult.ExecutionResult)
	}
}

func getSpecFiles(specSource string) []string {
	specFiles := make([]string, 0)
	if common.DirExists(specSource) {
		specFiles = append(specFiles, util.FindSpecFilesIn(specSource)...)
	} else if common.FileExists(specSource) && util.IsValidSpecExtension(specSource) {
		specFile, _ := filepath.Abs(specSource)
		specFiles = append(specFiles, specFile)
	}
	return specFiles
}

func specsFromArgs(conceptDictionary *conceptDictionary) []*specification {
	allSpecs := make([]*specification, 0)
	specs := make([]*specification, 0)
	var specParseResults []*parseResult
	for _, arg := range flag.Args() {
		specSource := arg
		if isIndexedSpec(specSource) {
			specs, specParseResults = getSpecWithScenarioIndex(specSource, conceptDictionary)
		} else {
			specs, specParseResults = findSpecs(specSource, conceptDictionary)
		}
		handleParseResult(specParseResults...)
		allSpecs = append(allSpecs, specs...)
	}
	return allSpecs
}

func getSpecWithScenarioIndex(specSource string, conceptDictionary *conceptDictionary) ([]*specification, []*parseResult) {
	specName, indexToFilter := GetIndexedSpecName(specSource)
	parsedSpecs, parseResult := findSpecs(specName, conceptDictionary)
	return filterSpecsItems(parsedSpecs, newScenarioIndexFilterToRetain(indexToFilter)), parseResult
}

func findSpecs(specSource string, conceptDictionary *conceptDictionary) ([]*specification, []*parseResult) {
	specFiles := getSpecFiles(specSource)

	return parseSpecFiles(specFiles, conceptDictionary)

}

func handleWarningMessages(warnings []string) {
	for _, warning := range warnings {
		logger.Log.Warning(warning)
	}
}

func printVersion() {
	fmt.Printf("Gauge version: %s\n\n", version.CurrentGaugeVersion.String())
	fmt.Println("Plugins\n-------")
	allPluginsWithVersion, err := common.GetAllInstalledPluginsWithVersion()
	if err != nil {
		fmt.Println("No plugins found")
		fmt.Println("Plugins can be installed with `gauge --install {plugin-name}`")
		os.Exit(0)
	}
	for _, pluginInfo := range allPluginsWithVersion {
		fmt.Printf("%s (%s)\n", pluginInfo.Name, pluginInfo.Version.String())
	}
}

type ByFileName []*specification

func (s ByFileName) Len() int {
	return len(s)
}

func (s ByFileName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByFileName) Less(i, j int) bool {
	return s[i].fileName < s[j].fileName
}

func sortSpecsList(allSpecs []*specification) []*specification {
	sort.Sort(ByFileName(allSpecs))
	return allSpecs
}

func setWorkingDir(workingDir string) {
	targetDir, err := filepath.Abs(workingDir)
	if err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("Unable to set working directory : %s\n", err)))
	}
	if !common.DirExists(targetDir) {
		err = os.Mkdir(targetDir, 0777)
		if err != nil {
			handleCriticalError(errors.New(fmt.Sprintf("Unable to set working directory : %s\n", err)))
		}
	}
	err = os.Chdir(targetDir)
	_, err = os.Getwd()
	if err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("Unable to set working directory : %s\n", err)))
	}
}
