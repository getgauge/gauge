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

	"strconv"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

var (
	telemetryCmd = &cobra.Command{
		Use:   "telemetry [command]",
		Short: "Configure telemetry options",
		Long:  "Configure telemetry options.",
		Example: `  gauge telemetry on
  gauge telemetry off
  gauge telemetry`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) != 0 {
				logger.Fatalf(cmd.UsageString())
			}
			fmt.Println(map[bool]string{true: "on", false: "off"}[config.TelemetryEnabled()])
		},
	}

	onCmd = &cobra.Command{
		Use:     "on",
		Short:   "Turn telemetry on",
		Long:    "Turn telemetry on.",
		Example: "  gauge telemetry on",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := config.UpdateTelemetry("true"); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}

	offCmd = &cobra.Command{
		Use:     "off",
		Short:   "Turn telemetry off",
		Long:    "Turn telemetry off.",
		Example: "  gauge telemetry off",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := config.UpdateTelemetry("false"); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}

	logCmd = &cobra.Command{
		Use:   "log <value>",
		Short: "Enable/disable telemetry logging",
		Long:  "Enable/disable telemetry logging.",
		Example: `  gauge telemetry log true
  gauge telemetry log false`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) < 1 {
				fmt.Println(config.TelemetryLogEnabled())
				return
			}
			if _, err := strconv.ParseBool(args[0]); err != nil {
				logger.Fatalf("Error: Invalid argument. The valid options are true or false.")
			}
			config.UpdateTelemetryLoggging(args[0])
		},
	}
)

func init() {
	telemetryCmd.AddCommand(onCmd)
	telemetryCmd.AddCommand(offCmd)
	telemetryCmd.AddCommand(logCmd)
	GaugeCmd.AddCommand(telemetryCmd)
}
