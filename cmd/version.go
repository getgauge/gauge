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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getgauge/gauge/plugin"
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
			setGlobalFlags()
			if machineReadable {
				PrintJSONVersion()
				return
			}
			PrintVersion()
		},
		DisableAutoGenTag: true,
	}
)

func init() {
	GaugeCmd.AddCommand(versionCmd)
}

func PrintJSONVersion() {
	type pluginJSON struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type versionJSON struct {
		Version    string        `json:"version"`
		CommitHash string        `json:"commitHash"`
		Plugins    []*pluginJSON `json:"plugins"`
	}
	gaugeVersion := versionJSON{version.FullVersion(), version.GetCommitHash(), make([]*pluginJSON, 0)}
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

func PrintVersion() {
	fmt.Printf("Gauge version: %s\n", version.FullVersion())
	v := version.GetCommitHash()
	if v != "" {
		fmt.Printf("Commit Hash: %s\n", v)

	}
	fmt.Printf("\nPlugins\n-------\n")
	allPluginsWithVersion, err := plugin.GetAllInstalledPluginsWithVersion()
	if err != nil {
		fmt.Println("No plugins found")
		fmt.Println("Plugins can be installed with `gauge install {plugin-name}`")
		os.Exit(0)
	}
	for _, pluginInfo := range allPluginsWithVersion {
		fmt.Printf("%s (%s)\n", pluginInfo.Name, filepath.Base(pluginInfo.Path))
	}
}
