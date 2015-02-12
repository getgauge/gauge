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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge_messages"
	flag "github.com/getgauge/mflag"
	"io"
	"io/ioutil"
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

func init() {
	acceptedExtensions[".spec"] = true
	acceptedExtensions[".md"] = true
}

var acceptedExtensions = make(map[string]bool)
var defaultPlugins = []string{"html-report"}

type manifest struct {
	Language string
	Plugins  []string
}

func main() {
	flag.Parse()
	setWorkingDir(*workingDir)
	validGaugeProject := true
	err := config.SetProjectRoot(flag.Args())
	if err != nil {
		validGaugeProject = false
	}
	initLoggers()
	if *daemonize && validGaugeProject {
		runInBackground()
	} else if *gaugeVersion {
		printVersion()
	} else if *specFilesToFormat != "" && validGaugeProject {
		formatSpecFiles(*specFilesToFormat)
	} else if *initialize != "" {
		initializeProject(*initialize)
	} else if *install != "" {
		downloadAndInstallPlugin(*install, *installVersion)
	} else if *addPlugin != "" {
		addPluginToProject(*addPlugin)
	} else if *refactor != "" && validGaugeProject {
		refactorSteps(*refactor, newStepName())
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		} else if validGaugeProject {
			executeSpecs(*parallel)
		} else {
			log.Error("Could not set project root: %s", err.Error())

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
			log.Error("%s \n", err)
		}
	}
	log.Info("%d specifications changed.\n", len(refactoringResult.specsChanged))
	log.Info("%d concepts changed.\n", len(refactoringResult.conceptsChanged))
	log.Info("%d files in code changed.\n", len(refactoringResult.runnerFilesChanged))
	os.Exit(exitCode)
}

func saveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		log.Error("Failed to refactor '%s': %s\n", fileName, err)
	}
}

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "Run as a daemon")
var gaugeVersion = flag.Bool([]string{"v", "-version"}, false, "Print the current version and exit. Eg: gauge -version")
var verbosity = flag.Bool([]string{"-verbose"}, false, "Enable verbose logging for debugging")
var logLevel = flag.String([]string{"-log-level"}, "", "Set level of logging to debug, info, warning, error or critical")
var simpleConsoleOutput = flag.Bool([]string{"-simple-console"}, false, "Removes colouring and simplifies from the console output")
var initialize = flag.String([]string{"-init"}, "", "Initializes project structure in the current directory. Eg: gauge --init java")
var install = flag.String([]string{"-install"}, "", "Downloads and installs a plugin. Eg: gauge --install java")
var installVersion = flag.String([]string{"-plugin-version"}, "", "Version of plugin to be installed. This is used with --install")
var currentEnv = flag.String([]string{"-env"}, "default", "Specifies the environment. If not specified, default will be used")
var addPlugin = flag.String([]string{"-add-plugin"}, "", "Adds the specified plugin to the current project")
var pluginArgs = flag.String([]string{"-plugin-args"}, "", "Specified additional arguments to the plugin. This is used together with --add-plugin")
var specFilesToFormat = flag.String([]string{"-format"}, "", "Formats the specified spec files")
var executeTags = flag.String([]string{"-tags"}, "", "Executes the specs and scenarios tagged with given tags. Eg: gauge --tags tag1,tag2 specs")
var apiPort = flag.String([]string{"-api-port"}, "", "Specifies the api port to be used. Eg: gauge --daemonize --api-port 7777")
var refactor = flag.String([]string{"-refactor"}, "", "Refactor steps")
var parallel = flag.Bool([]string{"-parallel"}, false, "Execute specs in parallel")
var workingDir = flag.String([]string{"-dir"}, ".", "Set the working directory for the current command, accepts a path relative to current directory.")
var doNotRandomize = flag.Bool([]string{"-sort", "-s"}, false, "run specs in Alphabetical Order. Eg: gauge --s specs")

func printUsage() {
	fmt.Printf("gauge - version %s\n", currentGaugeVersion.String())
	fmt.Printf("Copyright %d Thoughtworks\n\n", time.Now().Year())
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	os.Exit(2)
}

func downloadAndInstallPlugin(plugin, version string) {
	if err := installPlugin(plugin, version); err != nil {
		log.Warning("Failed to install plugin %s : %s\n", plugin, err)
	} else {
		log.Info("Successfully installed plugin => %s %s", plugin, version)
	}
}

func runInBackground() {
	var port int
	var err error
	if *apiPort != "" {
		port, err = strconv.Atoi(*apiPort)
		os.Setenv(common.ApiPortEnvVariableName, *apiPort)
		if err != nil {
			log.Critical("Failed to parse the port number :", *apiPort, "\n", err.Error())
			os.Exit(1)
		}
	} else {
		loadGaugeEnvironment()
		port, err = getPortFromEnvironmentVariable(common.ApiPortEnvVariableName)
		if err != nil {
			log.Critical("Failed to start API Service. %s \n", err.Error())
			os.Exit(1)
		}
	}
	var wg sync.WaitGroup
	runAPIServiceIndefinitely(port, &wg)
	wg.Wait()
}

func formatSpecFiles(filesToFormat string) {
	specs, specParseResults := findSpecs(filesToFormat, &conceptDictionary{})
	handleParseResult(specParseResults...)
	failed := false
	for _, spec := range specs {
		formatted := formatSpecification(spec)
		err := common.SaveFile(spec.fileName, formatted, true)
		if err != nil {
			failed = true
			log.Error("Failed to format '%s': %s\n", spec.fileName, err)
		}
	}
	if failed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func initializeProject(language string) {
	err := createProjectTemplate(language)
	if err != nil {
		log.Critical("Failed to initialize. %s\n", err.Error())
		os.Exit(1)
	}
	log.Info("\nSuccessfully initialized the project. Run specifications with \"gauge specs/\"")
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
	if err := addPluginToTheProject(pluginName, additionalArgs, getProjectManifest()); err != nil {
		log.Critical("Failed to add plugin %s to project : %s\n", pluginName, err.Error())
		os.Exit(1)
	} else {
		log.Info("Plugin %s was successfully added to the project\n", pluginName)
	}
}

func executeSpecs(inParallel bool) {
	loadGaugeEnvironment()
	conceptsDictionary, conceptParseResult := createConceptsDictionary(false)
	handleParseResult(conceptParseResult)
	specsToExecute, specsSkipped := getSpecsToExecute(conceptsDictionary)
	manifest := getProjectManifest()
	err := startAPIService(0)
	if err != nil {
		log.Critical("Failed to start gauge API. %s\n", err.Error())
		os.Exit(1)
	}

	runner, runnerError := startRunnerAndMakeConnection(manifest)
	if runnerError != nil {
		log.Critical("Failed to start a runner. %s\n", runnerError.Error())
		os.Exit(1)
	}
	validateSpecs(manifest, specsToExecute, runner, conceptsDictionary)
	if !*doNotRandomize {
		specsToExecute = shuffleSpecs(specsToExecute)
	}

	pluginHandler := startPlugins(manifest)
	execution := newExecution(manifest, specsToExecute, runner, pluginHandler, *parallel)

	result := execution.start()
	execution.finish()
	exitCode := printExecutionStatus(result, specsSkipped)
	os.Exit(exitCode)
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
func startPlugins(manifest *manifest) *pluginHandler {
	pluginHandler, warnings := startPluginsForExecution(manifest)
	handleWarningMessages(warnings)
	return pluginHandler
}

func validateSpecs(manifest *manifest, specsToExecute []*specification, runner *testRunner, conceptDictionary *conceptDictionary) {
	validator := newValidator(manifest, specsToExecute, runner, conceptDictionary)
	validationErrors := validator.validate()
	if len(validationErrors) > 0 {
		printValidationFailures(validationErrors)
		os.Exit(1)
	}
}

func getSpecsToExecute(conceptsDictionary *conceptDictionary) ([]*specification, int) {
	specsToExecute := specsFromArgs(conceptsDictionary)

	totalSpecs := specsToExecute
	if *executeTags != "" {
		validateTagExpression(*executeTags)
		specsToExecute = filterSpecsByTags(specsToExecute, *executeTags)
	}
	return sortSpecsList(specsToExecute), len(totalSpecs) - len(specsToExecute)
}

func printValidationFailures(validationErrors executionValidationErrors) {
	log.Warning("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			getCurrentConsole().writeError(fmt.Sprintf("%s:%d: %s. %s\n", stepValidationError.fileName, s.lineNo, stepValidationError.message, s.getLineText()))
		}
	}
}

func getSpecName(specSource string) string {
	if isIndexedSpec(specSource) {
		specSource, _ = GetIndexedSpecName(specSource)
	}
	return specSource
}

func (m *manifest) save() error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(common.ManifestFile, b, common.NewFilePermissions)
}

// All the environment variables loaded from the
// current environments JSON files will live here
type environmentVariables struct {
	Variables map[string]string
}

func getProjectManifest() *manifest {
	value := ""
	if len(flag.Args()) != 0 {
		value = flag.Args()[0]
	}
	projectRoot, err := common.GetProjectRootFromSpecPath(value)
	if err != nil {
		log.Critical("Failed to read manifest: %s \n", err.Error())
		os.Exit(1)
	}
	contents, err := common.ReadFileContents(path.Join(projectRoot, common.ManifestFile))
	if err != nil {
		log.Critical(err.Error())
		os.Exit(1)
	}
	dec := json.NewDecoder(strings.NewReader(contents))

	var m manifest
	for {
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Critical("Failed to read manifest. %s\n", err.Error())
			os.Exit(1)
		}
	}

	return &m
}

func showMessage(action, filename string) {
	log.Info(" %s  %s\n", action, filename)
}

func createProjectTemplate(language string) error {
	if !common.IsASupportedLanguage(language) {
		log.Info("%s plugin is not installed \n", language)
		log.Info("Installing plugin => %s ... \n\n", language)

		if err := installPlugin(language, ""); err != nil {
			return errors.New(fmt.Sprintf("Failed to install plugin %s . %s \n", language, err))
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

// Loads all the properties files available in the specified env directory
func loadEnvironment(env string) error {
	var err error
	var envDir string
	if len(flag.Args()) == 0 {
		envDir, err = common.GetDirInProject(common.EnvDirectoryName, "")
	} else {
		envDir, err = common.GetDirInProject(common.EnvDirectoryName, getSpecName(flag.Args()[0]))
	}
	if err != nil {
		log.Critical("Failed to Load environment: %s\n", err.Error())
		os.Exit(1)
	}

	dirToRead := path.Join(envDir, env)
	if !common.DirExists(dirToRead) {
		return errors.New(fmt.Sprintf("%s is an invalid environment", env))
	}

	isProperties := func(fileName string) bool {
		return filepath.Ext(fileName) == ".properties"
	}

	err = filepath.Walk(dirToRead, func(path string, info os.FileInfo, err error) error {
		if isProperties(path) {
			p, e := properties.Load(path)
			if e != nil {
				return errors.New(fmt.Sprintf("Failed to parse: %s. %s", path, e.Error()))
			}

			for k, v := range p {
				err := common.SetEnvVariable(k, v)
				if err != nil {
					return errors.New(fmt.Sprintf("%s: %s", path, err.Error()))
				}
			}
		}
		return nil
	})

	return err
}

func handleParseResult(results ...*parseResult) {
	for _, result := range results {
		if !result.ok {
			log.Critical(fmt.Sprintf("[ParseError] %s : %s", result.fileName, result.error.Error()))
			os.Exit(1)
		}
		if result.warnings != nil {
			for _, warning := range result.warnings {
				log.Warning("%s : %v", result.fileName, warning)
			}
		}
	}
}

func loadGaugeEnvironment() {
	// Loading default environment and loading user specified env
	// this way user specified env variable can override default if required
	err := loadEnvironment(envDefaultDirName)
	if err != nil {
		log.Critical("Failed to load the default environment. %s\n", err.Error())
		os.Exit(1)
	}

	if *currentEnv != envDefaultDirName {
		err := loadEnvironment(*currentEnv)
		if err != nil {
			log.Critical("Failed to load the environment: %s. %s\n", *currentEnv, err.Error())
			os.Exit(1)
		}
	}

}

func startRunnerAndMakeConnection(manifest *manifest) (*testRunner, error) {
	port, err := getPortFromEnvironmentVariable(common.GaugePortEnvName)
	if err != nil {
		port = 0
	}
	gaugeConnectionHandler, connHandlerErr := newGaugeConnectionHandler(port, nil)
	if connHandlerErr != nil {
		return nil, connHandlerErr
	}
	if err := common.SetEnvVariable(common.GaugeInternalPortEnvName, strconv.Itoa(gaugeConnectionHandler.connectionPortNumber())); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to set %s. %s", common.GaugePortEnvName, err.Error()))
	}

	testRunner, err := startRunner(manifest)
	if err != nil {
		return nil, err
	}

	runnerConnection, connectionError := gaugeConnectionHandler.acceptConnection(config.RunnerConnectionTimeout(), testRunner.errorChannel)
	testRunner.connection = runnerConnection
	if connectionError != nil {
		log.Debug("Runner connection error: %s", connectionError)
		testRunner.kill()
		return nil, connectionError
	}
	return testRunner, nil
}

func printExecutionStatus(suiteResult *suiteResult, specsSkipped int) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view

	noOfSpecificationsExecuted := len(suiteResult.specResults)
	noOfScenariosExecuted := 0
	noOfSpecificationsFailed := suiteResult.specsFailedCount
	noOfScenariosFailed := 0
	exitCode := 0
	if suiteResult.isFailed {
		log.Info("\nThe following failures occured:\n")
		exitCode = 1
	}

	printHookError(suiteResult.preSuite)

	for _, specResult := range suiteResult.specResults {
		noOfScenariosExecuted += specResult.scenarioCount
		noOfScenariosFailed += specResult.scenarioFailedCount
		printSpecFailure(specResult)
	}

	printHookError(suiteResult.postSuite)
	log.Info("%d scenarios executed, %d failed\n", noOfScenariosExecuted, noOfScenariosFailed)
	log.Info("%d specifications executed, %d failed\n", noOfSpecificationsExecuted, noOfSpecificationsFailed)
	log.Info("%d specifications skipped\n", specsSkipped)
	log.Info("%s\n", time.Millisecond*time.Duration(suiteResult.executionTime))
	return exitCode
}

func printHookError(hook *(gauge_messages.ProtoHookFailure)) {
	if hook != nil {
		console := getCurrentConsole()
		console.writeError(hook.GetErrorMessage())
		console.writeError(hook.GetStackTrace())
	}
}

func printError(execResult *gauge_messages.ProtoExecutionResult) {
	if execResult.GetFailed() {
		console := getCurrentConsole()
		console.writeError(execResult.GetErrorMessage() + "\n")
		console.writeError(execResult.GetStackTrace() + "\n")
	}
}

func printSpecFailure(specResult *specResult) {
	if specResult.isFailed {
		getCurrentConsole().writeError(fmt.Sprintf("%s : %s \n", specResult.protoSpec.GetFileName(), specResult.protoSpec.GetSpecHeading()))
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
		getCurrentConsole().writeError(fmt.Sprintf(" %s: \n", scenario.GetScenarioHeading()))
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
		getCurrentConsole().writeError(fmt.Sprintf("\t %s\n", step.GetActualText()))
		printHookError(stepExecResult.GetPreHookFailure())
		printError(stepExecResult.ExecutionResult)
		printHookError(stepExecResult.GetPostHookFailure())
	}
}

func printConceptFailure(concept *gauge_messages.ProtoConcept) {
	conceptExecResult := concept.ConceptExecutionResult
	if conceptExecResult != nil && conceptExecResult.GetExecutionResult().GetFailed() {
		getCurrentConsole().writeError(fmt.Sprintf("\t %s\n", concept.ConceptStep.GetActualText()))
		printError(conceptExecResult.ExecutionResult)
	}
}

func findConceptFiles() []string {
	value := ""
	if len(flag.Args()) != 0 {
		value = flag.Args()[0]
	}
	conceptsDir, err := common.GetDirInProject(common.SpecsDirectoryName, value)
	if err != nil {
		return []string{}
	}

	return common.FindFilesInDir(conceptsDir, func(path string) bool {
		return filepath.Ext(path) == common.ConceptFileExtension
	})

}

func createConceptsDictionary(shouldIgnoreErrors bool) (*conceptDictionary, *parseResult) {
	conceptFiles := findConceptFiles()
	conceptsDictionary := new(conceptDictionary)
	for _, conceptFile := range conceptFiles {
		if err := addConcepts(conceptFile, conceptsDictionary); err != nil {
			if shouldIgnoreErrors {
				apiLog.Error("Concept parse failure: %s %s", conceptFile, err)
				continue
			}
			log.Error(err.Error())
			return nil, &parseResult{error: err, fileName: conceptFile}
		}
	}
	return conceptsDictionary, &parseResult{ok: true}
}

func addConcepts(conceptFile string, conceptDictionary *conceptDictionary) *parseError {
	fileText, fileReadErr := common.ReadFileContents(conceptFile)
	if fileReadErr != nil {
		return &parseError{message: fmt.Sprintf("failed to read concept file %s", conceptFile)}
	}
	concepts, err := new(conceptParser).parse(fileText)
	if err != nil {
		return err
	}
	err = conceptDictionary.add(concepts, conceptFile)
	return err
}

func getSpecFiles(specSource string) []string {
	specFiles := make([]string, 0)
	if common.DirExists(specSource) {
		specFiles = append(specFiles, findSpecsFilesIn(specSource)...)
	} else if common.FileExists(specSource) && isValidSpecExtension(specSource) {
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
	parseResults := make([]*parseResult, 0)
	specs := make([]*specification, 0)
	for _, specFile := range specFiles {
		specFileContent, err := common.ReadFileContents(specFile)
		if err != nil {
			parseResults = append(parseResults, &parseResult{error: &parseError{message: err.Error()}, ok: false, fileName: specFile})
		}
		spec, parseResult := new(specParser).parse(specFileContent, conceptDictionary)
		parseResult.fileName = specFile
		if !parseResult.ok {
			return nil, append(parseResults, parseResult)
		} else {
			parseResults = append(parseResults, parseResult)
		}
		spec.fileName = specFile
		specs = append(specs, spec)
	}
	return specs, parseResults
}

func findSpecsFilesIn(dirRoot string) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	specFiles := common.FindFilesInDir(absRoot, isValidSpecExtension)
	return specFiles
}

func isValidSpecExtension(path string) bool {
	return acceptedExtensions[filepath.Ext(path)]
}

func handleWarningMessages(warnings []string) {
	for _, warning := range warnings {
		log.Warning(warning)
	}
}

func printVersion() {
	fmt.Printf("%s\n", currentGaugeVersion.String())
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
	pwd, err := os.Getwd()
	if err != nil {
		log.Critical("Unable to read current directory : %s\n", err)
		os.Exit(1)
	}
	targetDir := path.Join(pwd, workingDir)
	if !common.DirExists(targetDir) {
		err = os.Mkdir(targetDir, 0777)
		if err != nil {
			log.Critical("Unable to set working directory : %s\n", err)
			os.Exit(1)
		}
	}
	err = os.Chdir(targetDir)
	pwd, err = os.Getwd()
	if err != nil {
		log.Critical("Unable to set working directory : %s\n", err)
		os.Exit(1)
	}
}
