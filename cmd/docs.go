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
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:     "docs [flags] <plugin> [args]",
	Short:   "Generate documenation using specified plugin.",
	Long:    "Generate documenation using specified plugin.",
	Example: "  gauge docs spectacle specs/",
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		if err := isValidGaugeProject(args); err != nil {
			logger.Fatalf(err.Error())
		}
		if len(args) < 1 {
			logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
		}
		track.Docs(args[0])
		specDirs := getSpecsDir(args[1:])
		gaugeConnectionHandler := api.Start(specDirs)
		plugin.GenerateDoc(args[0], specDirs, gaugeConnectionHandler.ConnectionPortNumber())
	},
}

func init() {
	GaugeCmd.AddCommand(docsCmd)
}
