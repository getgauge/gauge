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
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:     "format [flags] [args]",
	Short:   "Formats the specified spec files",
	Long:    `Formats the specified spec files.`,
	Example: "  gauge format specs/",
	Run: func(cmd *cobra.Command, args []string) {
		if e := env.LoadEnv(environment); e != nil {
			logger.Fatalf(e.Error())
		}
		if err := config.SetProjectRoot(args); err != nil {
			exit(err, cmd.UsageString())
		}
		track.Format()
		formatter.FormatSpecFilesIn(getSpecsDir(args)[0])
	},
	DisableAutoGenTag: true,
}

func init() {
	GaugeCmd.AddCommand(formatCmd)
}
