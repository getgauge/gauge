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

var uninstallCmd = &cobra.Command{
	Use:     "uninstall [flags] <plugin>",
	Short:   "Uninstalls a plugin",
	Long:    `Uninstalls a plugin.`,
	Example: `  gauge uninstall java`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			exit(fmt.Errorf("Missing argument <plugin name>."), cmd.UsageString())
		}
		install.UninstallPlugin(args[0], pVersion)
	},
	DisableAutoGenTag: true,
}

func init() {
	GaugeCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringVarP(&pVersion, "version", "v", "", "Version of plugin to be uninstalled")
}
