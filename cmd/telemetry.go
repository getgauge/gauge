/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	telemetryCmd = &cobra.Command{
		Use:   "telemetry [command]",
		Short: "Configure options for sending anonymous usage stats",
		Long:  `Configure options for sending anonymous usage stats.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This command is deprecated, since Gauge no longer collects telemetry.")
		},
		DisableAutoGenTag: false,
	}

	onCmd = &cobra.Command{
		Use:   "on",
		Short: "Turn telemetry on",
		Long:  "Turn telemetry on.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This command is deprecated, since Gauge no longer collects telemetry.")
		},
		DisableAutoGenTag: false,
	}

	offCmd = &cobra.Command{
		Use:   "off",
		Short: "Turn telemetry off",
		Long:  "Turn telemetry off.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This command is deprecated, since Gauge no longer collects telemetry.")
		},
		DisableAutoGenTag: false,
	}

	logCmd = &cobra.Command{
		Use:   "log <value>",
		Short: "Enable/disable telemetry logging",
		Long:  "Enable/disable telemetry logging.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This command is deprecated, since Gauge no longer collects telemetry.")
		},
		DisableAutoGenTag: false,
	}
)

func init() {
	telemetryCmd.AddCommand(onCmd)
	telemetryCmd.AddCommand(offCmd)
	telemetryCmd.AddCommand(logCmd)
	GaugeCmd.AddCommand(telemetryCmd)
}
