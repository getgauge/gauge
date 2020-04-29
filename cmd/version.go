/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		fmt.Println("error:", err.Error())
	}
	for _, pluginInfo := range allPluginsWithVersion {
		gaugeVersion.Plugins = append(gaugeVersion.Plugins, &pluginJSON{pluginInfo.Name, filepath.Base(pluginInfo.Path)})
	}
	b, err := json.MarshalIndent(gaugeVersion, "", "    ")
	if err != nil {
		fmt.Println("error:", err.Error())
	}
	fmt.Println(fmt.Sprintf("%s\n", string(b)))
}

func printTextVersion() {
	fmt.Printf("Gauge version: %s\n", version.FullVersion())
	v := version.CommitHash
	if v != "" {
		fmt.Printf("Commit Hash: %s\n", v)

	}
	fmt.Printf("\nPlugins\n-------\n")
	allPluginsWithVersion, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		fmt.Println("No plugins found")
		fmt.Println("Plugins can be installed with `gauge install {plugin-name}`")
		os.Exit(0)
	}
	for _, pluginInfo := range allPluginsWithVersion {
		fmt.Printf("%s (%s)\n", pluginInfo.Name, filepath.Base(pluginInfo.Path))
	}
}
