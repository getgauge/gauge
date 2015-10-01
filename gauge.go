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
	"fmt"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/logger/execLogger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/project_init"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/version"
	flag "github.com/getgauge/mflag"
	"os"
	"time"
)

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "Run as a daemon")
var gaugeVersion = flag.Bool([]string{"v", "-version", "version"}, false, "Print the current version and exit. Eg: gauge --version")
var verbosity = flag.Bool([]string{"-verbose"}, false, "Enable verbose logging for debugging")
var logLevel = flag.String([]string{"-log-level"}, "", "Set level of logging to debug, info, warning, error or critical")
var simpleConsoleOutput = flag.Bool([]string{"-simple-console"}, false, "Removes colouring and simplifies from the console output")
var initialize = flag.String([]string{"-init"}, "", "Initializes project structure in the current directory. Eg: gauge --init java")
var installPlugin = flag.String([]string{"-install"}, "", "Downloads and installs a plugin. Eg: gauge --install java")
var uninstallPlugin = flag.String([]string{"-uninstall"}, "", "Uninstalls a plugin. Eg: gauge --uninstall java")
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
var refactorSteps = flag.String([]string{"-refactor"}, "", "Refactor steps")
var parallel = flag.Bool([]string{"-parallel", "p"}, false, "Execute specs in parallel")
var numberOfExecutionStreams = flag.Int([]string{"n"}, util.NumberOfCores(), "Specify number of parallel execution streams")
var distribute = flag.Int([]string{"g", "-group"}, -1, "Specify which group of specification to execute based on -n flag")
var workingDir = flag.String([]string{"-dir"}, ".", "Set the working directory for the current command, accepts a path relative to current directory.")
var strategy = flag.String([]string{"-strategy"}, "eager", "Set the parallelization strategy for execution. This is used along with -p flag. Ex: gauge -p --strategy=\"eager\" ")
var doNotRandomize = flag.Bool([]string{"-sort", "s"}, false, "Run specs in Alphabetical Order. Eg: gauge -s specs")
var check = flag.Bool([]string{"-check"}, false, "Checks for parse and validation errors. Eg: gauge --check specs")

func main() {
	flag.Parse()
	project_init.SetWorkingDir(*workingDir)
	initPackageFlags()
	validGaugeProject := true
	err := config.SetProjectRoot(flag.Args())
	if err != nil {
		validGaugeProject = false
	}
	env.LoadEnv(true)
	logger.Initialize(*verbosity, *logLevel, *simpleConsoleOutput)
	if *gaugeVersion {
		version.PrintVersion()
	} else if *daemonize {
		if validGaugeProject {
			api.RunInBackground(*apiPort)
		} else {
			logger.Log.Error(err.Error())
		}
	} else if *specFilesToFormat != "" {
		if validGaugeProject {
			formatter.FormatSpecFilesIn(*specFilesToFormat)
		} else {
			logger.Log.Error(err.Error())
		}
	} else if *initialize != "" {
		project_init.InitializeProject(*initialize)
	} else if *installZip != "" && *installPlugin != "" {
		install.InstallPluginZip(*installZip, *installPlugin)
	} else if *installPlugin != "" {
		install.DownloadAndInstallPlugin(*installPlugin, *installVersion)
	} else if *uninstallPlugin != "" {
		install.UninstallPlugin(*uninstallPlugin)
	} else if *installAll {
		install.InstallAllPlugins()
	} else if *update != "" {
		install.UpdatePlugin(*update)
	} else if *addPlugin != "" {
		install.AddPluginToProject(*addPlugin, *pluginArgs)
	} else if !validGaugeProject {
		logger.Log.Error(err.Error())
		os.Exit(1)
	} else if *refactorSteps != "" {
		startChan := api.StartAPI()
		refactor.RefactorSteps(*refactorSteps, newStepName(), startChan)
	} else if *check {
		execution.CheckSpecs(flag.Args())
	} else {
		if len(flag.Args()) == 0 {
			printUsage()
		} else {
			if *distribute != -1 {
				*doNotRandomize = true
			}
			execution.ExecuteSpecs(*parallel, flag.Args())
		}
	}
}

func printUsage() {
	fmt.Printf("gauge -version %s\n", version.CurrentGaugeVersion.String())
	fmt.Printf("Copyright %d Thoughtworks\n\n", time.Now().Year())
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	os.Exit(2)
}

func newStepName() string {
	if len(flag.Args()) != 1 {
		printUsage()
	}
	return flag.Args()[0]
}

func initPackageFlags() {
	if util.IsWindows() {
		*simpleConsoleOutput = true
	}
	execLogger.SimpleConsoleOutput = *simpleConsoleOutput
	env.ProjectEnv = *currentEnv
	execution.ExecuteTags = *executeTags
	execution.TableRows = *tableRows
	execution.NumberOfExecutionStreams = *numberOfExecutionStreams
	filter.ExecuteTags = *executeTags
	filter.DoNotRandomize = *doNotRandomize
	filter.Distribute = *distribute
	filter.NumberOfExecutionStreams = *numberOfExecutionStreams
	execution.Strategy = *strategy
}
