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
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	installCmd = &cobra.Command{
		Use:   "install [flags] [plugin]",
		Short: "Download and install plugin(s)",
		Long:  "Download and install specified plugin or all plugins in the project's `manifest.json` file.",
		Example: `  gauge install
  gauge install java
  gauge install java -f gauge-java-0.6.3-darwin.x86_64.zip`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) < 1 {
				track.InstallAll()
				install.InstallAllPlugins()
				return
			}
			if zip != "" {
				track.Install(args[0], true)
				install.HandleInstallResult(install.InstallPluginFromZipFile(zip, args[0]), args[0], true)
			} else {
				track.Install(args[0], false)
				install.HandleInstallResult(install.InstallPlugin(args[0], pVersion), args[0], true)
			}
		},
	}
	zip      string
	pVersion string
)

func init() {
	GaugeCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&zip, "file", "f", "", "Installs the plugin from zip file")
	installCmd.Flags().StringVarP(&pVersion, "version", "v", "", "Version of plugin to be installed")
}
