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
	"os"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/cmd"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/stream"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/projectInit"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/skel"
	"github.com/getgauge/gauge/track"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	flag "github.com/getgauge/mflag"
)

// Command line flags
var daemonize = flag.Bool([]string{"-daemonize"}, false, "[DEPRECATED] Use gauge daemon.")
var gaugeVersion = flag.Bool([]string{"v", "-version", "version"}, false, "[DEPRECATED] Use gauge version")
var verbosity = flag.Bool([]string{"-verbose"}, false, "[DEPRECATED] Use gauge run -v")
var logLevel = flag.String([]string{"-log-level"}, "", "Set level of logging to debug, info, warning, error or critical")
var simpleConsoleOutput = flag.Bool([]string{"-simple-console"}, false, "[DEPRECATED] gauge run --simple-console")
var initialize = flag.String([]string{"-init"}, "", "[DEPRECATED] gauge init <template name>")
var installPlugin = flag.String([]string{"-install"}, "", "[DEPRECATED] Use gauge install <plugin name>")
var uninstallPlugin = flag.String([]string{"-uninstall"}, "", "[DEPRECATED] Use gauge uninstall <plugin name>")
var installAll = flag.Bool([]string{"-install-all"}, false, "[DEPRECATED] Use gauge install --all")
var update = flag.String([]string{"-update"}, "", "[DEPRECATED] Use gauge update <plugin name>")
var pluginVersion = flag.String([]string{"-plugin-version"}, "", "[DEPRECATED] Use gauge [install|uninstall] <plugin name> -v <version>")
var installZip = flag.String([]string{"-file", "f"}, "", "[DEPRECATED] Use gauge install <plugin name> -f <zip file>")
var currentEnv = flag.String([]string{"-env"}, "default", "[DEPRECATED] Use gauge run -e <env name>")
var addPlugin = flag.String([]string{"-add-plugin"}, "", "[DEPRECATED] Use gauge add <plugin name>")
var pluginArgs = flag.String([]string{"-plugin-args"}, "", "[DEPRECATED] Use gauge add <plugin name> --args <args>")
var specFilesToFormat = flag.String([]string{"-format"}, "", "[DEPRECATED] Use gauge format specs/")
var executeTags = flag.String([]string{"-tags"}, "", "[DEPRECATED] Use gauge run --tags tag1,tag2 specs")
var tableRows = flag.String([]string{"-table-rows"}, "", "[DEPRECATED] gauge run --table-rows <rows>")
var apiPort = flag.String([]string{"-api-port"}, "", "[DEPRECATED] Use gauge daemon 7777")
var refactorSteps = flag.String([]string{"-refactor"}, "", "[DEPRECATED] Use gauge refactor <old step> <new step>")
var parallel = flag.Bool([]string{"-parallel", "p"}, false, "[DEPRECATED] guage run -p specs/")
var numberOfExecutionStreams = flag.Int([]string{"n"}, util.NumberOfCores(), "[DEPRECATED] Use guage run -p -n specs/")
var distribute = flag.Int([]string{"g", "-group"}, -1, "[DEPRECATED] Use gauge -n 5 -g 1 specs/")
var workingDir = flag.String([]string{"-dir"}, ".", "Set the working directory for the current command, accepts a path relative to current directory.")
var strategy = flag.String([]string{"-strategy"}, "lazy", "[DEPRECATED] Use gauge run -p --strategy=\"eager\"")
var sort = flag.Bool([]string{"-sort", "s"}, false, "[DEPRECATED] Use gauge run -s specs")
var validate = flag.Bool([]string{"-validate", "-check"}, false, "[DEPRECATED] Use gauge validate specs")
var updateAll = flag.Bool([]string{"-update-all"}, false, "[DEPRECATED] Use gauge update -a")
var checkUpdates = flag.Bool([]string{"-check-updates"}, false, "[DEPRECATED] Use gauge update -c")
var listTemplates = flag.Bool([]string{"-list-templates"}, false, "[DEPRECATED] Use gauge list-templates")
var machineReadable = flag.Bool([]string{"-machine-readable"}, false, "[DEPRECATED] Use gauge version -m")
var runFailed = flag.Bool([]string{"-failed"}, false, "[DEPRECATED] Use gauge run --failed")
var docs = flag.String([]string{"-docs"}, "", "[DEPRECATED] Use gauge docs <plugin name> specs/")

func main() {
	skel.CreateSkelFilesIfRequired()
	exit, err := cmd.Parse()
	if err == nil {
		os.Exit(exit)
	}
	os.Stderr.Write([]byte("[DEPRECATED] This usage will be removed soon. Run `gauge help --legacy` for more info.\n"))
	flag.Parse()
	logger.Initialize(*logLevel)
	util.SetWorkingDir(*workingDir)
	initPackageFlags()
	validGaugeProject := true
	err = config.SetProjectRoot(flag.Args())
	if err != nil {
		validGaugeProject = false
	}
	if *runFailed {
		logger.Fatalf("Rerun is not supported via the old usage. Use 'gauge run -f'")
	}
	if e := env.LoadEnv(*currentEnv); e != nil {
		logger.Fatalf(e.Error())
	}
	logger.Debugf("Gauge Install ID: %s", config.UniqueID())
	if *gaugeVersion && *machineReadable {
		printJSONVersion()
	} else if *machineReadable {
		fmt.Println("flag '--machine-readable' can only be used with 'version' subcommand")
		os.Exit(1)
	} else if *gaugeVersion {
		printVersion()
	} else if *initialize != "" {
		track.ProjectInit()
		projectInit.InitializeProject(*initialize)
	} else if *installZip != "" && *installPlugin != "" {
		track.Install(*installPlugin, true)
		install.HandleInstallResult(install.InstallPluginFromZipFile(*installZip, *installPlugin), *installPlugin, true)
	} else if *installPlugin != "" {
		track.Install(*installPlugin, false)
		install.HandleInstallResult(install.InstallPlugin(*installPlugin, *pluginVersion), *installPlugin, true)
	} else if *uninstallPlugin != "" {
		track.UninstallPlugin(*uninstallPlugin)
		install.UninstallPlugin(*uninstallPlugin, *pluginVersion)
	} else if *installAll {
		track.InstallAll()
		install.InstallAllPlugins()
	} else if *update != "" {
		track.Update(*update)
		install.HandleUpdateResult(install.InstallPlugin(*update, *pluginVersion), *update, true)
	} else if *updateAll {
		track.UpdateAll()
		install.UpdatePlugins()
	} else if *checkUpdates {
		track.CheckUpdates()
		install.PrintUpdateInfoWithDetails()
	} else if *addPlugin != "" {
		track.AddPlugins(*addPlugin)
		install.AddPluginToProject(*addPlugin, *pluginArgs)
	} else if *listTemplates {
		track.ListTemplates()
		projectInit.ListTemplates()
	} else if validGaugeProject {
		var specDirs = []string{common.SpecsDirectoryName}
		if len(flag.Args()) > 0 {
			specDirs = flag.Args()
		}
		if *refactorSteps != "" {
			track.Refactor()
			refactorInit(flag.Args())
		} else if *daemonize {
			track.Daemon()
			stream.Start()
			api.RunInBackground(*apiPort, specDirs)
		} else if *specFilesToFormat != "" {
			track.Format()
			formatter.FormatSpecFilesIn(*specFilesToFormat)
		} else if *validate {
			track.Validation()
			validation.Validate(flag.Args())
		} else if *docs != "" {
			track.Docs(*docs)
			gaugeConnectionHandler := api.Start(specDirs)
			plugin.GenerateDoc(*docs, specDirs, gaugeConnectionHandler.ConnectionPortNumber())
		} else {
			track.Execution(*parallel, *executeTags != "", *sort, *simpleConsoleOutput, *verbosity, *strategy)
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
	cmd.PrintJSONVersion()
}

func printVersion() {
	cmd.PrintVersion()
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
	if *distribute != -1 {
		execution.Strategy = execution.Eager
	}
}
