// This file is part of twist
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/twist2/common"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	manifestFile      = "manifest.json"
	specsDirName      = "specs"
	skelFileName      = "hello_world.spec"
	envDirName        = "env"
	envDefaultDirName = "default"
)

type step struct {
	File          string
	RawText       string
	ProcessedText string
	LineNo        int
	Args          []string
}

var availableSteps []*step

type manifest struct {
	Language string
}

// All the environment variables loaded from the
// current environments JSON files will live here
type environmentVariables struct {
	Variables map[string]string
}

func getProjectManifest() *manifest {
	contents := common.ReadFileContents(manifestFile)
	dec := json.NewDecoder(strings.NewReader(contents))

	var m manifest
	for {
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			common.PrintError(fmt.Sprintf("Failed to read: %s. %s\n", manifestFile, err.Error()))
			os.Exit(1)
		}
	}

	return &m
}

func findScenarioFiles(fileChan chan<- string) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	walkFn := func(filePath string, info os.FileInfo, err error) error {
		ext := path.Ext(info.Name())
		if strings.ToLower(ext) == ".scn" {
			fileChan <- filePath
		}
		return nil
	}

	filepath.Walk(pwd, walkFn)
	fileChan <- "done"
}

func parseScenarioFiles(fileChan <-chan string) {
	for {
		scenarioFilePath := <-fileChan
		if scenarioFilePath == "done" {
			break
		}

		tokens, err := parse(common.ReadFileContents(scenarioFilePath))
		if se, ok := err.(*syntaxError); ok {
			fmt.Printf("%s:%d:%d %s\n", scenarioFilePath, se.lineNo, se.colNo, se.message)
		} else {
			for _, token := range tokens {
				if token.kind == typeWorkflowStep {
					s := &step{File: scenarioFilePath, RawText: token.line, ProcessedText: token.value, LineNo: token.lineNo, Args: token.args}
					availableSteps = append(availableSteps, s)
				}
			}
		}
	}
}

func makeListOfAvailableSteps() {
	fileChan := make(chan string)
	go findScenarioFiles(fileChan)
	go parseScenarioFiles(fileChan)
}

func startAPIService() {
	http.HandleFunc("/steps", func(w http.ResponseWriter, r *http.Request) {
		js, err := json.Marshal(availableSteps)
		if err != nil {
			io.WriteString(w, err.Error())
		} else {
			w.Header()["Content-Type"] = []string{"application/json"}
			w.Write(js)
		}
	})
	log.Fatal(http.ListenAndServe(":8889", nil))
}

func showMessage(action, filename string) {
	fmt.Printf(" %s  %s\n", action, filename)
}

func createProjectTemplate(language string) error {
	if !common.IsASupportedLanguage(language) {
		return errors.New(fmt.Sprintf("%s is not a supported language", language))
	}

	// Create the project manifest
	showMessage("create", manifestFile)
	if common.FileExists(manifestFile) {
		showMessage("skip", manifestFile)
	}
	manifest := &manifest{Language: language}
	b, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	ioutil.WriteFile(manifestFile, b, common.NewFilePermissions)

	// creating the spec directory
	showMessage("create", specsDirName)
	if !common.DirExists(specsDirName) {
		err = os.Mkdir(specsDirName, common.NewDirectoryPermissions)
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
	showMessage("create", envDirName)
	if !common.DirExists(envDirName) {
		err = os.Mkdir(envDirName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", envDirName, err.Error()))
		}
	}
	defaultEnv := path.Join(envDirName, envDefaultDirName)
	showMessage("create", defaultEnv)
	if !common.DirExists(defaultEnv) {
		err = os.Mkdir(defaultEnv, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultEnv, err.Error()))
		}
	}
	defaultJson, err := common.GetSkeletonFilePath(path.Join(common.EnvDirectoryName, common.DefaultEnvJSONFileName))
	if err != nil {
		return err
	}
	defaultJsonDest := path.Join(defaultEnv, common.DefaultEnvJSONFileName)
	showMessage("create", defaultJsonDest)
	err = common.CopyFile(defaultJson, defaultJsonDest)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultJsonDest, err.Error()))
	}

	return executeInitHookForRunner(language)
}

// Loads all the json files available in the specified env directory
func loadEnvironment(env string) error {
	dirToRead := path.Join(common.EnvDirectoryName, env)
	if !common.DirExists(dirToRead) {
		return errors.New(fmt.Sprintf("%s is an invalid environment", env))
	}

	isJson := func(fileName string) bool {
		return filepath.Ext(fileName) == ".json"
	}

	err := filepath.Walk(dirToRead, func(path string, info os.FileInfo, err error) error {
		if isJson(path) {
			var e environmentVariables
			contents := common.ReadFileContents(path)
			err := json.Unmarshal([]byte(contents), &e)
			if err != nil {
				return errors.New(fmt.Sprintf("Failed to parse: %s. %s", path, err.Error()))
			}

			for k, v := range e.Variables {
				err := common.SetEnvVariable(k, string(v))
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
var daemonize = flag.Bool("daemonize", false, "Run as a daemon")
var initialize = flag.String("init", "", "Initializes project structure in the current directory")
var currentEnv = flag.String("env", "default", "Specifies the environment")

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: twist [options] scenario\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()

	if *daemonize {
		makeListOfAvailableSteps()
		startAPIService()
	} else if *initialize != "" {
		err := createProjectTemplate(*initialize)
		if err != nil {
			fmt.Printf("Failed to initialize. %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println("Successfully initialized the project")
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		}

		err := loadEnvironment(*currentEnv)
		if err != nil {
			fmt.Printf("Failed to load the environment. %s\n", err.Error())
			os.Exit(1)
		}

		scenarioFile := flag.Arg(0)
		tokens, err := parse(common.ReadFileContents(scenarioFile))
		if se, ok := err.(*syntaxError); ok {
			fmt.Printf("%s:%d:%d %s\n", scenarioFile, se.lineNo, se.colNo, se.message)
			os.Exit(1)
		}

		manifest := getProjectManifest()

		_, err = startRunner(manifest)
		if err != nil {
			fmt.Printf("Failed to start a runner. %s\n", err.Error())
			os.Exit(1)
		}

		conn, err := acceptConnection()
		if err != nil {
			fmt.Printf("Failed to get a runner. %s\n", err.Error())
			os.Exit(1)
		}

		execution := newExecution(manifest, tokens, conn)
		err = execution.start()
		if err != nil {
			fmt.Printf("Execution failed. %s\n", err.Error())
			os.Exit(1)
		}
	}
}
