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
	"os"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:     "config [flags] [args]",
		Short:   "Change global configurations",
		Long:    "Change global configurations.",
		Example: `  gauge config check_updates false`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if list || machineReadable {
				exit(config.List(machineReadable))
			}
			if len(args) == 0 {
				logger.Fatalf("Error: Config command needs argument(s).\n%s", cmd.UsageString())
			}
			if len(args) == 1 {
				exit(config.GetProperty(args[0]))
			}
			if err := config.Update(args[0], args[1]); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}
	list bool
)

func init() {
	GaugeCmd.AddCommand(configCmd)
	configCmd.Flags().BoolVarP(&list, "list", "", false, "List all global properties")
	configCmd.Flags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Print all properties in JSON format")
}

func exit(text string, err error) {
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Infof(text)
	os.Exit(0)
}
