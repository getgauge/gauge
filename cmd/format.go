/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/formatter"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:     "format [flags] [args]",
	Short:   "Formats the specified spec files",
	Long:    `Formats the specified spec files.`,
	Example: "  gauge format specs/",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.SetProjectRoot(args); err != nil {
			exit(err, cmd.UsageString())
		}
		loadEnvAndReinitLogger(cmd)
		formatter.FormatSpecFilesIn(getSpecsDir(args)[0])
	},
	DisableAutoGenTag: true,
}

func init() {
	GaugeCmd.AddCommand(formatCmd)
}
