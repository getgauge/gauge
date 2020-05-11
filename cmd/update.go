/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"fmt"

	"github.com/getgauge/gauge/plugin/install"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update [flags] <plugin>",
		Short: "Updates a plugin",
		Long:  `Updates a plugin.`,
		Example: `  gauge update java
  gauge update -a
  gauge update -c`,
		Run: func(cmd *cobra.Command, args []string) {
			if all {
				install.UpdatePlugins(machineReadable)
				return
			} else if check {
				install.PrintUpdateInfoWithDetails()
				return
			}
			if len(args) < 1 {
				exit(fmt.Errorf("missing argument <plugin name>"), cmd.UsageString())
			}
			install.HandleUpdateResult(install.Plugin(args[0], pVersion, machineReadable), args[0], true)
		},
		DisableAutoGenTag: true,
	}
	all   bool
	check bool
)

func init() {
	GaugeCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&all, "all", "a", false, "Updates all the installed Gauge plugins")
	updateCmd.Flags().BoolVarP(&check, "check", "c", false, "Checks for Gauge and plugins updates")
}
