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

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:     "config [flags] [args]",
		Short:   "Change global configurations",
		Long:    `Change global configurations.`,
		Example: `  gauge config check_updates false`,
		Run: func(cmd *cobra.Command, args []string) {
			if list || machineReadable {
				text, err := config.List(machineReadable)
				if err != nil {
					logger.Fatalf(err.Error())
				}
				logger.Infof(text)
				return
			}
			if len(args) == 0 {
				exit(fmt.Errorf("Config command needs argument(s)."), cmd.UsageString())
			}
			if len(args) == 1 {
				text, err := config.GetProperty(args[0])
				if err != nil {
					logger.Fatalf(err.Error())
				}
				logger.Infof(text)
				return
			}
			if err := config.Update(args[0], args[1]); err != nil {
				logger.Fatalf(err.Error())
			}
		},
		DisableAutoGenTag: true,
	}
	list bool
)

func init() {
	GaugeCmd.AddCommand(configCmd)
	configCmd.Flags().BoolVarP(&list, "list", "", false, "List all global properties")
	configCmd.Flags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Print all properties in JSON format")
}
