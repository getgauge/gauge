/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version [flags]",
		Short: "Print Gauge and plugin versions",
		Long:  `Print Gauge and plugin versions.`,
		Example: `  gauge version
  gauge version -m`,
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) { /* noop */ },
		DisableAutoGenTag: true,
	}
)

func init() {
	GaugeCmd.AddCommand(versionCmd)
}

func printVersion() {
	if machineReadable {
		printJSONVersion()
		return
	}
	printTextVersion()
}

func printJSONVersion() {
	type pluginJSON struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type versionJSON struct {
		Version    string        `json:"version"`
		CommitHash string        `json:"commitHash"`
		Plugins    []*pluginJSON `json:"plugins"`
	}
	gaugeVersion := versionJSON{version.FullVersion(), version.CommitHash, make([]*pluginJSON, 0)}
	allPluginsWithVersion, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		logger.Errorf(false, "Error fetching plugins info: %s", err.Error())

	}
	for _, pluginInfo := range allPluginsWithVersion {
		gaugeVersion.Plugins = append(gaugeVersion.Plugins, &pluginJSON{pluginInfo.Name, filepath.Base(pluginInfo.Path)})
	}
	b, err := json.MarshalIndent(gaugeVersion, "", "    ")
	if err != nil {
		fmt.Println("Error fetching version info as JSON:", err.Error())
		return
	}
	// logger can not be used, since it breaks the json format,
	// The logger adds out, nessage as key which vscode plugin does not understand.
	fmt.Printf("%s\n\n", string(b))
}

func printTextVersion() {
	logger.Infof(true, "Gauge version: %s", version.FullVersion())
	v := version.CommitHash
	if v != "" {
		logger.Infof(true, "Commit Hash: %s\n", v)

	}
	logger.Infof(true, "Plugins\n-------")
	allPluginsWithVersion, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		logger.Infof(true, "No plugins found\nPlugins can be installed with `gauge install {plugin-name}`")
	}
	for _, pluginInfo := range allPluginsWithVersion {
		logger.Infof(true, "%s (%s)", pluginInfo.Name, filepath.Base(pluginInfo.Path))
	}
}
