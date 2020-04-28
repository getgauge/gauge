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
