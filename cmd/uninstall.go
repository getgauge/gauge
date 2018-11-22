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
	"fmt"

	"github.com/getgauge/gauge/plugin/install"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall [flags] <plugin>",
	Short:   "Uninstalls a plugin",
	Long:    `Uninstalls a plugin.`,
	Example: `  gauge uninstall java`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			exit(fmt.Errorf("Missing argument <plugin name>."), cmd.UsageString())
		}
		install.UninstallPlugin(args[0], pVersion)
	},
	DisableAutoGenTag: true,
}

func init() {
	GaugeCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringVarP(&pVersion, "version", "v", "", "Version of plugin to be uninstalled")
}
