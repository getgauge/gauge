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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/execution/stream"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/validation"
	"github.com/getgauge/gauge/version"

	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/projectInit"
	"github.com/getgauge/gauge/skel"
	"github.com/getgauge/gauge/util"
	flag "github.com/getgauge/mflag"
)

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "Run as a daemon")
var gaugeVersion = flag.Bool([]string{"v", "-version", "version"}, false, "Print the current version and exit. Eg: gauge --version")
var verbosity = flag.Bool([]string{"-verbose"}, false, "Enable step level reporting on console, default being scenario level. Eg: gauge --verbose specs")
var logLevel = flag.String([]string{"-log-level"}, "", "Set level of logging to debug, info, warning, error or critical")
var simpleConsoleOutput = flag.Bool([]string{"-simple-console"}, false, "Removes colouring and simplifies from the console output")
var initialize = flag.String([]string{"-init"}, "", "Initializes project structure in the current directory. Eg: gauge --init java")
var installPlugin = flag.String([]string{"-install"}, "", "Downloads and installs a plugin. Eg: gauge --install java")
var uninstallPlugin = flag.String([]string{"-uninstall"}, "", "Uninstalls a plugin. Eg: gauge --uninstall java")
var installAll = flag.Bool([]string{"-install-all"}, false, "Installs all the plugins specified in project manifest, if not installed. Eg: gauge --install-all")
var update = flag.String([]string{"-update"}, "", "Updates a plugin. Eg: gauge --update java")
var pluginVersion = flag.String([]string{"-plugin-version"}, "", "Version of plugin to be installed. This is used with --install or --uninstall flag.")
var installZip = flag.String([]string{"-file", "f"}, "", "Installs the plugin from zip file. This is used with --install. Eg: gauge --install java -f ZIP_FILE")
var currentEnv = flag.String([]string{"-env"}, "default", "Specifies the environment. If not specified, default will be used")
var addPlugin = flag.String([]string{"-add-plugin"}, "", "Adds the specified non-language plugin to the current project")
var pluginArgs = flag.String([]string{"-plugin-args"}, "", "Specified additional arguments to the plugin. This is used together with --add-plugin")
var specFilesToFormat = flag.String([]string{"-format"}, "", "Formats the specified spec files")
var executeTags = flag.String([]string{"-tags"}, "", "Executes the specs and scenarios tagged with given tags. Eg: gauge --tags tag1,tag2 specs")
var tableRows = flag.String([]string{"-table-rows"}, "", "Executes the specs and scenarios only for the selected rows. It can be specified by range as 2-4 or as list 2,4. Eg: gauge --table-rows \"1-3\" specs/hello.spec")
var apiPort = flag.String([]string{"-api-port"}, "", "Specifies the api port to be used. Eg: gauge --daemonize --api-port 7777")
var refactorSteps = flag.String([]string{"-refactor"}, "", "Refactor steps. Eg: gauge --refactor <old step> <new step> [[spec directories]]")
var parallel = flag.Bool([]string{"-parallel", "p"}, false, "Execute specs in parallel")
var numberOfExecutionStreams = flag.Int([]string{"n"}, util.NumberOfCores(), "Specify number of parallel execution streams")
var distribute = flag.Int([]string{"g", "-group"}, -1, "Specify which group of specification to execute based on -n flag")
var workingDir = flag.String([]string{"-dir"}, ".", "Set the working directory for the current command, accepts a path relative to current directory.")
var strategy = flag.String([]string{"-strategy"}, "lazy", "Set the parallelization strategy for execution. Possible options are: `eager`, `lazy`. Ex: gauge -p --strategy=\"eager\"")
var sort = flag.Bool([]string{"-sort", "s"}, false, "Run specs in Alphabetical Order. Eg: gauge -s specs")
var validate = flag.Bool([]string{"-validate", "#-check"}, false, "Check for validation and parse errors. Eg: gauge --validate specs")
var updateAll = flag.Bool([]string{"-update-all"}, false, "Updates all the installed Gauge plugins. Eg: gauge --update-all")
var checkUpdates = flag.Bool([]string{"-check-updates"}, false, "Checks for Gauge and plugins updates. Eg: gauge --check-updates")
var listTemplates = flag.Bool([]string{"-list-templates"}, false, "Lists all the Gauge templates available. Eg: gauge --list-templates")
var machineReadable = flag.Bool([]string{"-machine-readable"}, false, "Used with `--version` to produce JSON output of currently installed Gauge and plugin versions. e.g: gauge --version --machine-readable")
var runFailed = flag.Bool([]string{"-failed"}, false, "Run only the scenarios failed in previous run. Eg: gauge --failed")
var docs = flag.String([]string{"-docs"}, "", "Generate documenation using specified plugin. Eg: gauge --docs <plugin name> specs/")

func main() {
	skel.CreateSkelFilesIfRequired()
	flag.Parse()
	util.SetWorkingDir(*workingDir)
	initPackageFlags()
	validGaugeProject := true
	err := config.SetProjectRoot(flag.Args())
	if err != nil {
		validGaugeProject = false
	}
	if rerunErr := rerun.Initialize(); rerunErr != nil {
		fmt.Println(rerunErr)
		os.Exit(0)
	}
	if e := env.LoadEnv(*currentEnv); e != nil {
		logger.Fatalf(e.Error())
	}
	logger.Initialize(*logLevel)
	if *gaugeVersion && *machineReadable {
		printJSONVersion()
	} else if *machineReadable {
		fmt.Printf("flag '--machine-readable' can only be used with '--version' or '-v'\n\n")
		fmt.Printf("Usage:\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	} else if *gaugeVersion {
		printVersion()
	} else if *initialize != "" {
		projectInit.InitializeProject(*initialize)
	} else if *installZip != "" && *installPlugin != "" {
		install.HandleInstallResult(install.InstallPluginFromZipFile(*installZip, *installPlugin), *installPlugin, true)
	} else if *installPlugin != "" {
		install.HandleInstallResult(install.InstallPlugin(*installPlugin, *pluginVersion), *installPlugin, true)
	} else if *uninstallPlugin != "" {
		install.UninstallPlugin(*uninstallPlugin, *pluginVersion)
	} else if *installAll {
		install.InstallAllPlugins()
	} else if *update != "" {
		install.HandleUpdateResult(install.InstallPlugin(*update, *pluginVersion), *update, true)
	} else if *updateAll {
		install.UpdatePlugins()
	} else if *checkUpdates {
		install.PrintUpdateInfoWithDetails()
	} else if *addPlugin != "" {
		install.AddPluginToProject(*addPlugin, *pluginArgs)
	} else if *listTemplates {
		projectInit.ListTemplates()
	} else if flag.NFlag() == 0 && len(flag.Args()) == 0 {
		printUsage()
		os.Exit(0)
	} else if validGaugeProject {
		var specDirs = []string{common.SpecsDirectoryName}
		if len(flag.Args()) > 0 {
			specDirs = flag.Args()
		}
		if *refactorSteps != "" {
			refactorInit(flag.Args())
		} else if *daemonize {
			stream.Start()
			api.RunInBackground(*apiPort, specDirs)
		} else if *specFilesToFormat != "" {
			formatter.FormatSpecFilesIn(*specFilesToFormat)
		} else if *validate {
			validation.Validate(flag.Args())
		} else if *docs != "" {
			gaugeConnectionHandler := api.Start(specDirs)
			plugin.GenerateDoc(*docs, specDirs, gaugeConnectionHandler.ConnectionPortNumber())
		} else {
			exitCode := execution.ExecuteSpecs(specDirs)
			os.Exit(exitCode)
		}
	} else {
		logger.Fatalf(err.Error())
	}
}

func refactorInit(args []string) {
	if len(args) < 1 {
		logger.Fatalf("Flag needs at least two arguments: --refactor\nUsage : gauge --refactor <old step> <new step> [[spec directories]]")
	}
	var specDirs = []string{common.SpecsDirectoryName}
	if len(args) > 1 {
		specDirs = args[1:]
	}
	startChan := api.StartAPI(false)
	refactor.RefactorSteps(*refactorSteps, args[0], startChan, specDirs)
}

func printJSONVersion() {
	type pluginJSON struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type versionJSON struct {
		Version string        `json:"version"`
		Plugins []*pluginJSON `json:"plugins"`
	}
	gaugeVersion := versionJSON{version.FullVersion(), make([]*pluginJSON, 0)}
	allPluginsWithVersion, err := plugin.GetAllInstalledPluginsWithVersion()
	for _, pluginInfo := range allPluginsWithVersion {
		gaugeVersion.Plugins = append(gaugeVersion.Plugins, &pluginJSON{pluginInfo.Name, filepath.Base(pluginInfo.Path)})
	}
	b, err := json.MarshalIndent(gaugeVersion, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(fmt.Sprintf("%s\n", string(b)))
}

func printVersion() {
	fmt.Printf("Gauge version: %s\n\n", version.FullVersion())
	fmt.Println("Plugins\n-------")
	allPluginsWithVersion, err := plugin.GetAllInstalledPluginsWithVersion()
	if err != nil {
		fmt.Println("No plugins found")
		fmt.Println("Plugins can be installed with `gauge --install {plugin-name}`")
		os.Exit(0)
	}
	for _, pluginInfo := range allPluginsWithVersion {
		fmt.Printf("%s (%s)\n", pluginInfo.Name, filepath.Base(pluginInfo.Path))
	}
}

func printUsage() {
	fmt.Printf("Gauge version %s\n", version.FullVersion())
	fmt.Printf("Copyright %d ThoughtWorks, Inc.\n\n", time.Now().Year())
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}

func initPackageFlags() {
	if *parallel {
		*simpleConsoleOutput = true
		reporter.IsParallel = true
	}
	reporter.SimpleConsoleOutput = *simpleConsoleOutput
	reporter.Verbose = *verbosity
	execution.ExecuteTags = *executeTags
	execution.SetTableRows(*tableRows)
	validation.TableRows = *tableRows
	execution.NumberOfExecutionStreams = *numberOfExecutionStreams
	execution.InParallel = *parallel
	execution.Strategy = *strategy
	filter.ExecuteTags = *executeTags
	order.Sorted = *sort
	filter.Distribute = *distribute
	filter.NumberOfExecutionStreams = *numberOfExecutionStreams
	reporter.NumberOfExecutionStreams = *numberOfExecutionStreams
	rerun.RunFailed = *runFailed
	if *distribute != -1 {
		execution.Strategy = execution.Eager
	}
}
