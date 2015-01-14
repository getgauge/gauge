package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	flag "github.com/getgauge/mflag"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	specsDirName      = "specs"
	skelFileName      = "hello_world.spec"
	envDefaultDirName = "default"
)

var acceptedExtensions = make(map[string]bool)
var defaultPlugins = []string{"html-report"}

func init() {
	acceptedExtensions[".spec"] = true
	acceptedExtensions[".md"] = true
}

type manifest struct {
	Language string
	Plugins  []string
}

func main() {
	flag.Parse()
	if *daemonize {
		runInBackground()
	} else if *gaugeVersion {
		printVersion()
	} else if *specFilesToFormat != "" {
		formatSpecFiles(*specFilesToFormat)
	} else if *initialize != "" {
		initializeProject(*initialize)
	} else if *install != "" {
		downloadAndInstallPlugin(*install, *installVersion)
	} else if *addPlugin != "" {
		addPluginToProject(*addPlugin)
	} else if *refactor != "" {
		refactorSteps(*refactor)
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		}
		executeSpecs()
	}
}

func refactorSteps(oldStep string) {
	if len(flag.Args()) != 1 {
		printUsage()
	}
	if oldStep == flag.Args()[0] {
		return
	}
	projectRoot, err := common.GetProjectRoot()
	if err != nil {
		fmt.Printf("Failed to find project root: %s", err)
		os.Exit(1)
	}
	specs, specParseResult := findSpecs(projectRoot, &conceptDictionary{})
	handleParseResult(specParseResult...)
	agent, err := getRefactorAgent(oldStep, flag.Args()[0])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	runner := agent.startRunner()
	err, stepName, isStepPresent := agent.getStepNameFromRunner(runner)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	conceptDictionary, parseResult := createConceptsDictionary(false)
	handleParseResult(parseResult)
	specsRefactored, conceptFilesRefactored := agent.refactor(&specs, conceptDictionary)

	specCount, conceptCount := writeToConceptAndSpecFiles(specs, conceptDictionary, specsRefactored, conceptFilesRefactored)
	printSummary(specCount, conceptCount)
	if isStepPresent {
		agent.requestRunnerForRefactoring(runner, stepName)
	}
}

func printSummary(specFilesCount int, conceptFilesCount int) {
	fmt.Println(strconv.Itoa(specFilesCount) + " specifications changed.")
	fmt.Println(strconv.Itoa(conceptFilesCount) + " concept files changed.")
}

func writeToConceptAndSpecFiles(specs []*specification, conceptDictionary *conceptDictionary, specsRefactored map[*specification]bool, conceptFilesRefactored map[string]bool) (int, int) {
	specFilesCount := 0
	conceptFilesCount := 0
	for _, spec := range specs {
		if specsRefactored[spec] {
			specFilesCount++
			formatted := formatSpecification(spec)
			saveFile(spec.fileName, formatted, true)
		}
	}
	conceptMap := formatConcepts(conceptDictionary)
	for fileName, concept := range conceptMap {
		if conceptFilesRefactored[fileName] {
			conceptFilesCount++
			saveFile(fileName, concept, true)
		}
	}
	return specFilesCount, conceptFilesCount
}

func saveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		fmt.Printf("Failed to refactor '%s': %s\n", fileName, err)
	}
}

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "Run as a daemon")
var gaugeVersion = flag.Bool([]string{"v", "-version"}, false, "Print the current version and exit. Eg: gauge -version")
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

func printUsage() {
	fmt.Printf("gauge - version %s\n", currentGaugeVersion.String())
	fmt.Println("Copyright 2014 Thoughtworks\n")
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	os.Exit(2)
}

func downloadAndInstallPlugin(plugin, version string) {
	if err := installPlugin(plugin, version); err != nil {
		fmt.Printf("[Error] Failed to install plugin %s : %s\n", plugin, err)
	} else {
		fmt.Printf("Successfully installed plugin => %s %s", plugin, version)
	}
}

func setCurrentProjectEnvVariable() error {
	value := ""
	if len(flag.Args()) != 0 {
		value = flag.Args()[0]
	}
	projectRoot, err := common.GetProjectRootFromSpecPath(value)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to find gauge project root: %s ", err.Error()))
	}
	err = common.SetEnvVariable(common.GaugeProjectRootEnv, projectRoot)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting environment variable project_root : %s ", err.Error()))
	}
	return nil
}

func runInBackground() {
	var port int
	var err error
	if *apiPort != "" {
		port, err = strconv.Atoi(*apiPort)
		os.Setenv(common.ApiPortEnvVariableName, *apiPort)
		if err != nil {
			fmt.Println("Failed to parse the port number :", *apiPort, "\n", err.Error())
			os.Exit(1)
		}
	} else {
		loadGaugeEnvironment()
		port, err = getPortFromEnvironmentVariable(common.ApiPortEnvVariableName)
		if err != nil {
			fmt.Printf("Failed to start API Service. %s \n", err.Error())
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
			fmt.Printf("Failed to format '%s': %s\n", spec.fileName, err)
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
		fmt.Printf("[Error] Failed to initialize. %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("\nSuccessfully initialized the project. Run specifications with \"gauge specs/\"")
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
		fmt.Printf("Failed to add plugin %s to project : %s\n", pluginName, err.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Plugin %s was successfully added to the project\n", pluginName)
	}
}

func executeSpecs() {
	loadGaugeEnvironment()
	conceptsDictionary, conceptParseResult := createConceptsDictionary(false)
	handleParseResult(conceptParseResult)
	specsToExecute := getSpecsToExecute(conceptsDictionary)
	manifest := getProjectManifest()
	err := startAPIService(0)
	if err != nil {
		fmt.Printf("Failed to start gauge API. %s\n", err.Error())
		os.Exit(1)
	}

	runner, runnerError := startRunnerAndMakeConnection(manifest)
	if runnerError != nil {
		fmt.Printf("Failed to start a runner. %s\n", runnerError.Error())
		os.Exit(1)
	}
	pluginHandler := startPlugins(manifest)

	execution := newExecution(manifest, specsToExecute, runner, pluginHandler)
	validationErrors := execution.validate(conceptsDictionary)
	if len(validationErrors) > 0 {
		printValidationFailures(validationErrors)
		execution.stopAllPlugins()
		os.Exit(1)
	} else {
		status := execution.start()
		exitCode := printExecutionStatus(status)
		os.Exit(exitCode)
	}
}

func startPlugins(manifest *manifest) *pluginHandler {
	pluginHandler, warnings := startPluginsForExecution(manifest)
	handleWarningMessages(warnings)
	return pluginHandler
}

func getSpecsToExecute(conceptsDictionary *conceptDictionary) []*specification {
	specsToExecute := specsFromArgs(conceptsDictionary)

	if *executeTags != "" {
		validateTagExpression(*executeTags)
		specsToExecute = filterSpecsByTags(specsToExecute, *executeTags)
	}
	return sortSpecsList(specsToExecute)
}

func printValidationFailures(validationErrors executionValidationErrors) {
	fmt.Println("Validation failed. The following steps have errors")
	for _, stepValidationErrors := range validationErrors {
		for _, stepValidationError := range stepValidationErrors {
			s := stepValidationError.step
			getCurrentConsole().writeError(fmt.Sprintf("%s:%d: %s. %s\n", stepValidationError.fileName, s.lineNo, stepValidationError.message, s.lineText))
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
		fmt.Printf("Failed to read manifest: %s \n", err.Error())
		os.Exit(1)
	}
	contents, err := common.ReadFileContents(path.Join(projectRoot, common.ManifestFile))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dec := json.NewDecoder(strings.NewReader(contents))

	var m manifest
	for {
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Failed to read manifest. %s\n", err.Error())
			// common.PrintError(fmt.Sprintf("Failed to read: %s. %s\n", manifestFile, err.Error()))
			os.Exit(1)
		}
	}

	return &m
}

func showMessage(action, filename string) {
	fmt.Printf(" %s  %s\n", action, filename)
}

func createProjectTemplate(language string) error {
	if !common.IsASupportedLanguage(language) {
		fmt.Printf("%s plugin is not installed \n", language)
		fmt.Printf("Installing plugin => %s ... \n\n", language)

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
		fmt.Printf("Failed to Load environment: %s\n", err.Error())
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
			fmt.Println(fmt.Sprintf("[ParseError] %s : %s", result.fileName, result.error.Error()))
			os.Exit(1)
		}
		if result.warnings != nil {
			for _, warning := range result.warnings {
				fmt.Println(fmt.Sprintf("[Warning] %s : %v", result.fileName, warning))
			}
		}
	}
}

func loadGaugeEnvironment() {
	// Loading default environment and loading user specified env
	// this way user specified env variable can override default if required
	err := loadEnvironment(envDefaultDirName)
	if err != nil {
		fmt.Printf("Failed to load the default environment. %s\n", err.Error())
		os.Exit(1)
	}

	if *currentEnv != envDefaultDirName {
		err := loadEnvironment(*currentEnv)
		if err != nil {
			fmt.Printf("Failed to load the environment: %s. %s\n", *currentEnv, err.Error())
			os.Exit(1)
		}
	}

}

func startRunnerAndMakeConnection(manifest *manifest) (*testRunner, error) {
	if err := setCurrentProjectEnvVariable(); err != nil {
		return nil, err
	}

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
		testRunner.kill()
		return nil, connectionError
	}
	return testRunner, nil
}

func printExecutionStatus(suiteResult *suiteResult) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view

	noOfSpecificationsExecuted := len(suiteResult.specResults)
	noOfScenariosExecuted := 0
	noOfSpecificationsFailed := suiteResult.specsFailedCount
	noOfScenariosFailed := 0
	exitCode := 0
	if suiteResult.isFailed {
		fmt.Println("\nThe following failures occured:\n")
		exitCode = 1
	}

	printHookError(suiteResult.preSuite)

	for _, specResult := range suiteResult.specResults {
		noOfScenariosExecuted += specResult.scenarioCount
		noOfScenariosFailed += specResult.scenarioFailedCount
		printSpecFailure(specResult)
	}

	printHookError(suiteResult.postSuite)
	fmt.Printf("\n\n%d scenarios executed, %d failed\n", noOfScenariosExecuted, noOfScenariosFailed)
	fmt.Printf("%d specifications executed, %d failed\n", noOfSpecificationsExecuted, noOfSpecificationsFailed)
	return exitCode
}

func printHookError(hook *ProtoHookFailure) {
	if hook != nil {
		console := getCurrentConsole()
		console.writeError(hook.GetErrorMessage())
		console.writeError(hook.GetStackTrace())
	}
}

func printError(execResult *ProtoExecutionResult) {
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
			if specItem.GetItemType() == ProtoItem_Scenario {
				printScenarioFailure(specItem.GetScenario())
			} else if specItem.GetItemType() == ProtoItem_TableDrivenScenario {
				printTableDrivenScenarioFailure(specItem.GetTableDrivenScenario())
			}
		}

		printHookError(specResult.protoSpec.GetPostHookFailure())
	}
}

func printTableDrivenScenarioFailure(tableDrivenScenario *ProtoTableDrivenScenario) {
	for _, scenario := range tableDrivenScenario.GetScenarios() {
		printScenarioFailure(scenario)
	}
}

func printScenarioFailure(scenario *ProtoScenario) {
	if scenario.GetFailed() {
		getCurrentConsole().writeError(fmt.Sprintf(" %s: \n", scenario.GetScenarioHeading()))
		printHookError(scenario.GetPreHookFailure())

		for _, scenarioItem := range scenario.GetScenarioItems() {
			if scenarioItem.GetItemType() == ProtoItem_Step {
				printStepFailure(scenarioItem.GetStep())
			} else if scenarioItem.GetItemType() == ProtoItem_Concept {
				printConceptFailure(scenarioItem.GetConcept())
			}
		}
		printHookError(scenario.GetPostHookFailure())
	}

}

func printStepFailure(step *ProtoStep) {
	stepExecResult := step.StepExecutionResult
	if stepExecResult != nil && stepExecResult.ExecutionResult.GetFailed() {
		getCurrentConsole().writeError(fmt.Sprintf("\t %s\n", step.GetActualText()))
		printHookError(stepExecResult.GetPreHookFailure())
		printError(stepExecResult.ExecutionResult)
		printHookError(stepExecResult.GetPostHookFailure())
	}
}

func printConceptFailure(concept *ProtoConcept) {
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
				continue
			}
			fmt.Println(err)
			return nil, &parseResult{error: err, fileName: conceptFile}
		}
	}
	//	result := conceptsDictionary.resolveNestedConcepts()
	//	return conceptsDictionary, result
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
		return specFiles
	} else if common.FileExists(specSource) && isValidSpecExtension(specSource) {
		specFile, _ := filepath.Abs(specSource)
		specFiles = append(specFiles, specFile)
		return specFiles
	}
	return nil
}

func filterSpecsByScenarioIndex(specs []*specification, specSource string) []*specification {
	_, indexToFilter := GetIndexedSpecName(specSource)
	return filterSpecsItems(specs, newScenarioIndexFilterToRetain(indexToFilter))
}

func specsFromArgs(conceptDictionary *conceptDictionary) []*specification {
	allSpecs := make([]*specification, 0)
	for _, arg := range flag.Args() {
		specSource := arg
		specs, specParseResults := findSpecs(specSource, conceptDictionary)
		if isIndexedSpec(specSource) {
			specs = filterSpecsByScenarioIndex(specs, specSource)
		}
		handleParseResult(specParseResults...)
		allSpecs = append(allSpecs, specs...)
	}
	return allSpecs
}

func findSpecs(specSource string, conceptDictionary *conceptDictionary) ([]*specification, []*parseResult) {
	specFiles := getSpecFiles(specSource)
	if specFiles == nil {
		fmt.Printf("Spec file or directory does not exist: %s\n", specSource)
		os.Exit(1)
	} else if len(specFiles) == 0 {
		fmt.Printf("No spec files were found in %s\n", specSource)
		os.Exit(1)
	}
	parseResults := make([]*parseResult, 0)
	specs := make([]*specification, 0)
	for _, specFile := range specFiles {
		specFileContent, err := common.ReadFileContents(specFile)
		if err != nil {
			fmt.Println(err)
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
		fmt.Println(fmt.Sprintf("[Warning] %s", warning))
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
