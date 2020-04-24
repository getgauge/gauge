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
		Use:     "init [flags] <template>",
		Short:   "Initialize project structure in the current directory",
		Long:    `Initialize project structure in the current directory.`,
		Example: "  gauge init java",
		Run: func(cmd *cobra.Command, args []string) {
			if templates {
				l, err := template.All()
				if err != nil {
					exit(fmt.Errorf("Failed to get templates. %w", err), cmd.UsageString())
				}
				fmt.Println(l)
				return
			} else if url != "" {
				projectInit.InitializeProjectFromURL(url, machineReadable)
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
	url       string
)

func init() {
	GaugeCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&templates, "templates", "t", false, "Lists all available templates")
	initCmd.Flags().StringVarP(&url, "url", "u", "", "Initialize a project from given template URL")
}
