// This file is part of twist
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	manifestFile = "manifest.json"
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

func getProjectManifest() *manifest {
	contents := readFileContents(manifestFile)
	dec := json.NewDecoder(strings.NewReader(contents))

	var m manifest
	for {
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Failed to read: %s. %s\n", manifestFile, err.Error())
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

		tokens, err := parse(readFileContents(scenarioFilePath))
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

func createProjectTemplate(projectName string) {
	if exist, _ := exists(projectName); !exist {
		fmt.Println("Creating directory ", projectName)
		err := os.Mkdir(projectName, 0766)
		if err != nil {
			panic(err)
		}
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Command line flags
var daemonize = flag.Bool("daemonize", false, "Run as a daemon")
var create = flag.String("create", "", "Create a template")
var wd = flag.String("wd", "", "the working directory from which the executable has to run")

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: twist [options] scenario\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()
	if *wd != "" {
		os.Chdir(*wd)
		value, _ := os.Getwd()
		fmt.Println("Current working dir", value)
	}

	if *daemonize {
		makeListOfAvailableSteps()
		startAPIService()
	} else if *create != "" {
		createProjectTemplate(*create)
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		}

		scenarioFile := flag.Arg(0)
		tokens, err := parse(readFileContents(scenarioFile))
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
