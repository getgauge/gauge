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

var (
	updateCmd = &cobra.Command{
		Use:   "update [flags] <plugin>",
		Short: "Updates a plugin",
		Long:  `Updates a plugin.`,
		Example: `  gauge update java
  gauge update -a
  gauge update -c`,
		Run: func(cmd *cobra.Command, args []string) {
			if all {
				install.UpdatePlugins(machineReadable)
				return
			} else if check {
				install.PrintUpdateInfoWithDetails()
				return
			}
			if len(args) < 1 {
				exit(fmt.Errorf("missing argument <plugin name>"), cmd.UsageString())
			}
			install.HandleUpdateResult(install.Plugin(args[0], pVersion, machineReadable), args[0], true)
		},
		DisableAutoGenTag: true,
	}
	all   bool
	check bool
)

func init() {
	GaugeCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&all, "all", "a", false, "Updates all the installed Gauge plugins")
	updateCmd.Flags().BoolVarP(&check, "check", "c", false, "Checks for Gauge and plugins updates")
}
