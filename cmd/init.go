/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"fmt"

	"github.com/getgauge/gauge/template"

	"github.com/getgauge/gauge/projectInit"
	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init <template> [flags]",
		Short: "Initialize project structure in the current directory",
		Long:  `Initialize project structure in the current directory.`,
		Example: `  gauge init java
  gauge init https://template.org/java.zip
  gauge init /Users/user/Download/foo.zip`,
		Run: func(cmd *cobra.Command, args []string) {
			if templates {
				l, err := template.All()
				if err != nil {
					exit(fmt.Errorf("failed to get templates. %w", err), cmd.UsageString())
				}
				fmt.Println(l)
			} else {
				if len(args) < 1 {
					exit(fmt.Errorf("missing argument <template name, URL or filepath>. To see all the templates, run 'gauge init -t'"), cmd.UsageString())
				}
				projectInit.Template(args[0], machineReadable)
			}
		},
		DisableAutoGenTag: true,
	}
	templates bool
)

func init() {
	GaugeCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&templates, "templates", "t", false, "Lists all available templates")
}
