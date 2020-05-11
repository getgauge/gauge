/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"fmt"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/plugin"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:     "docs [flags] <plugin> [args]",
	Short:   "Generate documentation using specified plugin",
	Long:    `Generate documentation using specified plugin.`,
	Example: "  gauge docs spectacle specs/",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.SetProjectRoot(args); err != nil {
			exit(err, cmd.UsageString())
		}
		loadEnvAndReinitLogger(cmd)
		if len(args) < 1 {
			exit(fmt.Errorf("Missing argument <plugin name>."), cmd.UsageString())
		}
		specDirs := getSpecsDir(args[1:])
		var startAPIFunc = func(specDirs []string) int {
			gaugeConnectionHandler := api.Start(specDirs)
			return gaugeConnectionHandler.ConnectionPortNumber()
		}
		plugin.GenerateDoc(args[0], specDirs, startAPIFunc)
	},
	DisableAutoGenTag: true,
}

func init() {
	GaugeCmd.AddCommand(docsCmd)
}
