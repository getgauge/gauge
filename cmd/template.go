/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"fmt"

	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/template"
	"github.com/spf13/cobra"
)

var (
	templateCmd = &cobra.Command{
		Use:     "template [flags] [args]",
		Short:   "Change template configurations",
		Long:    `Change template configurations.`,
		Example: `  gauge template java getgauge/java-template`,
		Run: func(cmd *cobra.Command, args []string) {
			if templateList || machineReadable {
				text, err := template.List(machineReadable)
				if err != nil {
					logger.Fatalf(true, err.Error())
				}
				logger.Infof(true, text)
				return
			}
			if len(args) == 0 {
				exit(fmt.Errorf("template command needs argument(s)"), cmd.UsageString())
			}
			if len(args) == 1 {
				text, err := template.Get(args[0])
				if err != nil {
					logger.Fatalf(true, err.Error())
				}
				logger.Infof(true, text)
				return
			}
			err := template.Update(args[0], args[1])
			if err != nil {
				exit(fmt.Errorf("template URL should be a valid link of zip"), cmd.UsageString())
			}
		},
		DisableAutoGenTag: true,
	}
	templateList bool
)

func init() {
	GaugeCmd.AddCommand(templateCmd)
	templateCmd.Flags().BoolVarP(&templateList, "list", "", false, "List all template properties")
}
