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

	"github.com/getgauge/gauge/projectInit"
	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:     "init [flags] <template>",
		Short:   "Initialize project structure in the current directory",
		Long:    `Initialize project structure in the current directory.`,
		Example: "  gauge init java",
		Run: func(cmd *cobra.Command, args []string) {
			if templates {
				projectInit.ListTemplates()
				return
			}
			if len(args) < 1 {
				exit(fmt.Errorf("Missing argument <template name>. To see all the templates, run 'gauge init -t'"), cmd.UsageString())
			}
			projectInit.InitializeProject(args[0], machineReadable)
		},
		DisableAutoGenTag: true,
	}
	templates bool
)

func init() {
	GaugeCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&templates, "templates", "t", false, "Lists all available templates")
}
