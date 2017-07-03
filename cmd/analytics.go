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
	analyticsCmd = &cobra.Command{
		Use:   "analytics [command]",
		Short: "Turn analytics on/off",
		Long:  "Turn analytics on/off.",
		Example: `  gauge analytics on
  gauge analytics off
  gauge analytics`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) != 0 {
				logger.Fatalf(cmd.UsageString())
			}
			fmt.Println(map[bool]string{true: "on", false: "off"}[config.AnalyticsEnabled()])
		},
	}

	onCmd = &cobra.Command{
		Use:     "on",
		Short:   "Turn analytics on",
		Long:    "Turn analytics on.",
		Example: "  gauge analytics on",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := config.UpdateAnalytics("true"); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}

	offCmd = &cobra.Command{
		Use:     "off",
		Short:   "Turn analytics off",
		Long:    "Turn analytics off.",
		Example: "  gauge analytics off",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := config.UpdateAnalytics("false"); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}
)

func init() {
	analyticsCmd.AddCommand(onCmd)
	analyticsCmd.AddCommand(offCmd)
	GaugeCmd.AddCommand(analyticsCmd)
}
