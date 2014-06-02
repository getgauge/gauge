// This file is part of twist
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"io"
	"io/ioutil"
	flag "mflag"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	specsDirName      = "specs"
	skelFileName      = "hello_world.spec"
	envDefaultDirName = "default"
)

var availableSpecs = make([]*specification, 0)
var availableStepsMap = make(map[string]bool)
var acceptedExtensions = make(map[string]bool)

func init() {
	acceptedExtensions[".spec"] = true
	acceptedExtensions[".md"] = true
}

type pluginDetails struct {
	Id      string
	Version string
}

type manifest struct {
	Language string
	Plugins  []pluginDetails
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
	projectRoot, err := common.GetProjectRoot()
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
		return errors.New(fmt.Sprintf("%s is not a supported language", language))
	}

	// Create the project manifest
	showMessage("create", common.ManifestFile)
	if common.FileExists(common.ManifestFile) {
		showMessage("skip", common.ManifestFile)
	}
	manifest := &manifest{Language: language}
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
	envDir, err := common.GetDirInProject(common.EnvDirectoryName)
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

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "Run as a daemon")
var version = flag.Bool([]string{"v", "-version"}, false, "Print the current version and exit")
var initialize = flag.String([]string{"-init"}, "", "Initializes project structure in the current directory. Eg: gauge --init java")
var currentEnv = flag.String([]string{"-env"}, "default", "Specifies the environment. If not specified, default will be used")
var addPlugin = flag.String([]string{"-add-plugin"}, "", "Adds the specified plugin to the current project")
var pluginArgs = flag.String([]string{"-plugin-args"}, "", "Specified additional arguments to the plugin. This is used together with --add-plugin")
var specFilesToFormat = flag.String([]string{"-format"}, "", "Formats the specified spec files")

func printUsage() {
	fmt.Printf("gauge - version %d.%d.%d\n", MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)
	fmt.Println("Copyright 2014 Thoughtworks\n")
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	os.Exit(2)
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

func main() {
	flag.Parse()

	if *daemonize {
		loadGaugeEnvironment()
		makeListOfAvailableSteps()
		startAPIService()
	} else if *version {
		printVersion()
	} else if *specFilesToFormat != "" {
		specs, specParseResults := findSpecs(*specFilesToFormat, &conceptDictionary{})
		handleParseResult(specParseResults...)
		failed := false
		for _, spec := range specs {
			formatted := formatSpecification(spec)
			err := common.SaveFile(spec.fileName, formatted, true)
			if err != nil {
				failed = true
				fmt.Printf("Failed to format '%s': %s\n", spec.fileName)
			}
		}
		if failed {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else if *initialize != "" {
		err := createProjectTemplate(*initialize)
		if err != nil {
			fmt.Printf("Failed to initialize. %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println("\nSuccessfully initialized the project. Run specifications with \"gauge specs/\"")
	} else if *addPlugin != "" {
		pluginName := *addPlugin
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
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		}

		loadGaugeEnvironment()

		specSource := flag.Arg(0)

		concepts, conceptParseResult := createConceptsDictionary(false)
		handleParseResult(conceptParseResult)

		specs, specParseResults := findSpecs(specSource, concepts)
		handleParseResult(specParseResults...)

		manifest := getProjectManifest()

		pluginHandler, warnings := startPluginsForExecution(manifest)
		handleWarningMessages(warnings)

		runnerConnection, runnerError := startRunnerAndMakeConnection(manifest)
		if runnerError != nil {
			fmt.Printf("Failed to start a runner. %s\n", runnerError.Error())
			os.Exit(1)
		}

		execution := newExecution(manifest, specs, runnerConnection, pluginHandler)
		validationErrors := execution.validate(concepts)
		if len(validationErrors) > 0 {
			fmt.Println("Validation failed. The following steps have errors")
			for _, stepValidationErrors := range validationErrors {
				for _, stepValidationError := range stepValidationErrors {
					s := stepValidationError.step
					fmt.Printf("\x1b[31;1m  %s:%d: %s. %s\n\x1b[0m", stepValidationError.fileName, s.lineNo, stepValidationError.message, s.lineText)
				}
			}
			err := execution.killProcess()
			if err != nil {
				fmt.Printf("Failed to kill the process. %s\n", err.Error())
			}
			os.Exit(1)
		} else {
			status := execution.start()
			exitCode := printExecutionStatus(status)
			os.Exit(exitCode)
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

func startRunnerAndMakeConnection(manifest *manifest) (net.Conn, error) {
	listener, listenerErr := newGaugeListener(common.GaugePortEnvName, 0)
	if listenerErr != nil {
		return nil, listenerErr
	}
	if err := common.SetEnvVariable(common.GaugeInternalPortEnvName, strconv.Itoa(listener.tcpListener.Addr().(*net.TCPAddr).Port)); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to set %s. %s", common.GaugePortEnvName, err.Error()))
	}

	testRunner, err := startRunner(manifest)
	if err != nil {
		return nil, err
	}

	runnerConnection, connectionError := listener.acceptConnection(runnerConnectionTimeOut)
	if connectionError != nil {
		testRunner.cmd.Process.Kill()
		return nil, connectionError
	}
	return runnerConnection, nil
}

func printExecutionStatus(status *testExecutionStatus) int {
	// Print out all the errors that happened during the execution
	// helps to view all the errors in one view
	noOfSpecificationsExecuted := len(status.specExecutionStatuses)
	noOfScenariosExecuted := 0
	noOfSpecificationsFailed := 0
	noOfScenariosFailed := 0
	exitCode := 0
	if status.isFailed() {
		fmt.Println("\nThe following failures occured:\n")
		exitCode = 1
	}

	for _, hookStatus := range status.hooksExecutionStatuses {
		if !hookStatus.GetPassed() {
			fmt.Printf("\x1b[31;1m%s\n\x1b[0m", hookStatus.GetErrorMessage())
			fmt.Printf("\x1b[31;1m%s\n\x1b[0m", hookStatus.GetStackTrace())
		}
	}

	for _, specExecStatus := range status.specExecutionStatuses {
		specFailing := false
		for _, hookStatus := range specExecStatus.hooksExecutionStatuses {
			if !hookStatus.GetPassed() {
				specFailing = true
				fmt.Printf("\x1b[31;1m%s\n\x1b[0m", specExecStatus.specification.fileName)
				fmt.Printf("\x1b[31;1m%s\n\x1b[0m", hookStatus.GetErrorMessage())
				fmt.Printf("\x1b[31;1m%s\n\x1b[0m", hookStatus.GetStackTrace())
			}
		}

		noOfScenariosExecuted += len(specExecStatus.scenariosExecutionStatuses[0])
		scenariosFailedInThisSpec := printScenarioExecutionStatus(specExecStatus.scenariosExecutionStatuses[0], specExecStatus.specification)
		if scenariosFailedInThisSpec > 0 {
			specFailing = true
			noOfScenariosFailed += scenariosFailedInThisSpec
		}

		if specFailing {
			noOfSpecificationsFailed += 1
		}
	}

	fmt.Printf("\n\n%d scenarios executed, %d failed\n", noOfScenariosExecuted, noOfScenariosFailed)
	fmt.Printf("%d specifications executed, %d failed\n", noOfSpecificationsExecuted, noOfSpecificationsFailed)
	return exitCode
}

func printScenarioExecutionStatus(scenariosExecStatuses []*scenarioExecutionStatus, specification *specification) int {
	noOfScenariosFailed := 0
	scenarioFailing := false
	for _, scenarioExecStatus := range scenariosExecStatuses {
		for _, hookStatus := range scenarioExecStatus.hooksExecutionStatuses {
			if !hookStatus.GetPassed() {
				scenarioFailing = true
				fmt.Printf("\x1b[31;1m%s:%s:%s\n\x1b[0m", specification.fileName,
					scenarioExecStatus.scenario.heading.value, hookStatus.GetErrorMessage())
			}
		}

		for _, stepExecStatus := range scenarioExecStatus.stepExecutionStatuses {
			for _, executionStatus := range stepExecStatus.executionStatus {
				if !executionStatus.GetPassed() {
					scenarioFailing = true
					fmt.Printf("\x1b[31;1m%s:%s\n\x1b[0m", specification.fileName, executionStatus.GetErrorMessage())
				}
			}
		}
		if scenarioFailing {
			noOfScenariosFailed += 1
		}
	}

	return noOfScenariosFailed
}

func findConceptFiles() []string {
	conceptsDir, err := common.GetDirInProject(common.SpecsDirectoryName)
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

func findSpecs(specSource string, conceptDictionary *conceptDictionary) ([]*specification, []*parseResult) {
	specFiles := make([]string, 0)
	parseResults := make([]*parseResult, 0)
	if common.DirExists(specSource) {
		specFiles = append(specFiles, findSpecsFilesIn(specSource)...)
	} else if common.FileExists(specSource) && isValidSpecExtension(specSource) {
		specFile, _ := filepath.Abs(specSource)
		specFiles = append(specFiles, specFile)
	} else {
		fmt.Printf("Spec file or directory does not exist: %s", specSource)
		os.Exit(1)
	}

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
	return common.FindFilesInDir(absRoot, isValidSpecExtension)
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
	fmt.Printf("%d.%d.%d\n", MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)
}
