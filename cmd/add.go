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
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:     "add [flags] <plugin>",
		Short:   "Adds the specified non-language plugin to the current project",
		Long:    `Adds the specified non-language plugin to the current project.`,
		Example: "  gauge add xml-report",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) < 1 {
				logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
			}
			track.AddPlugins(args[0])
			install.AddPluginToProject(args[0], pArgs)
		},
	}
	pArgs string
)

func init() {
	GaugeCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&pArgs, "args", "", "", "Specified additional arguments to the plugin")
}
