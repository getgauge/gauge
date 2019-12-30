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

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/plugin"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:     "docs [flags] <plugin> [args]",
	Short:   "Generate documentation using specified plugin",
	Long:    `Generate documentation using specified plugin.`,
	Example: "  gauge docs spectacle specs/",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.SetProjectRoot(args); err != nil {
			exit(err, cmd.UsageString())
		}
		loadEnvAndReinitLogger(cmd)
		if len(args) < 1 {
			exit(fmt.Errorf("Missing argument <plugin name>."), cmd.UsageString())
		}
		specDirs := getSpecsDir(args[1:])
		var startAPIFunc = func(specDirs []string) int {
			gaugeConnectionHandler := api.Start(specDirs)
			return gaugeConnectionHandler.ConnectionPortNumber()
		}
		plugin.GenerateDoc(args[0], specDirs, startAPIFunc)
	},
	DisableAutoGenTag: true,
}

func init() {
	GaugeCmd.AddCommand(docsCmd)
}
