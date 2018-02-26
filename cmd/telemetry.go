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
	"os"
	"strings"

	"strconv"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

const gaugeTelemetryEnabled = "GAUGE_TELEMETRY_ENABLED"

var telemetryEnv = os.Getenv(gaugeTelemetryEnabled)

func telemetryEnabled() bool {
	boolValue, err := strconv.ParseBool(strings.TrimSpace(telemetryEnv))
	if err != nil {
		return config.TelemetryEnabled()
	}
	return boolValue
}

var (
	telemetryCmd = &cobra.Command{
		Use:   "telemetry [command]",
		Short: "Configure options for sending anonymous usage stats",
		Long:  `Configure options for sending anonymous usage stats.`,
		Example: `  gauge telemetry on
  gauge telemetry off
  gauge telemetry`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				exit(nil, cmd.UsageString())
			}
			fmt.Println(map[bool]string{true: "on", false: "off"}[telemetryEnabled()])
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if v, err := strconv.ParseBool(strings.TrimSpace(telemetryEnv)); err == nil {
				logger.Infof(true, "ENV[%s]=%t. Overrides telemetry configuration.", gaugeTelemetryEnabled, v)
			}
		},
		DisableAutoGenTag: true,
	}

	onCmd = &cobra.Command{
		Use:     "on",
		Short:   "Turn telemetry on",
		Long:    "Turn telemetry on.",
		Example: "  gauge telemetry on",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.UpdateTelemetry("true"); err != nil {
				logger.Fatalf(true, err.Error())
			}
		},
		DisableAutoGenTag: true,
	}

	offCmd = &cobra.Command{
		Use:     "off",
		Short:   "Turn telemetry off",
		Long:    "Turn telemetry off.",
		Example: "  gauge telemetry off",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.UpdateTelemetry("false"); err != nil {
				logger.Fatalf(true, err.Error())
			}
		},
		DisableAutoGenTag: true,
	}

	logCmd = &cobra.Command{
		Use:   "log <value>",
		Short: "Enable/disable telemetry logging",
		Long:  "Enable/disable telemetry logging.",
		Example: `  gauge telemetry log true
  gauge telemetry log false`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println(config.TelemetryLogEnabled())
				return
			}
			if _, err := strconv.ParseBool(args[0]); err != nil {
<<<<<<< HEAD
				exit(fmt.Errorf("Invalid argument. The valid options are true or false."), cmd.UsageString())
=======
				logger.Fatalf(true, "Error: Invalid argument. The valid options are true or false.")
>>>>>>> hide api and lsp logs from consumers
			}
			config.UpdateTelemetryLoggging(args[0])
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
		DisableAutoGenTag: true,
	}
)

func init() {
	telemetryCmd.AddCommand(onCmd)
	telemetryCmd.AddCommand(offCmd)
	telemetryCmd.AddCommand(logCmd)
	GaugeCmd.AddCommand(telemetryCmd)
}
