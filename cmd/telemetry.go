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

func init() {
	// This command is deprecated, it's going to be removed in the future
	telemetryCmd := &cobra.Command{
		Use: "telemetry [command]",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Gauge does not gather telemetry data. This command exists to notify this deprecation and shall be removed in the future")
		},
		Hidden: true,
	}

	GaugeCmd.AddCommand(telemetryCmd)
}
