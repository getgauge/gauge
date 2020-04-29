/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"os"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/api/lang"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/spf13/cobra"
)

const (
	isDaemon = "IS_DAEMON"
)

var (
	daemonCmd = &cobra.Command{
		Use:     "daemon [flags] <port> [args]",
		Short:   "Run as a daemon",
		Long:    `Run as a daemon.`,
		Example: "  gauge daemon 1234",
		Run: func(cmd *cobra.Command, args []string) {
			err := os.Setenv(isDaemon, "true")
			if err != nil {
				exit(err, "Unable to set Daemon=true")
			}
			if err := config.SetProjectRoot(args); err != nil {
				exit(err, cmd.UsageString())
			}
			loadEnvAndReinitLogger(cmd)
			if lsp {
				lang.Start(&infoGatherer.SpecInfoGatherer{SpecDirs: getSpecsDir(args)}, logLevel)
				return
			}
			port := ""
			specs := util.GetSpecDirs()
			if len(args) > 0 {
				port = args[0]
				specs = getSpecsDir(args[1:])
			}
			api.RunInBackground(port, specs)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) { /* noop */ },
		DisableAutoGenTag: true,
	}
	lsp bool
)

func init() {
	GaugeCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().BoolVarP(&lsp, "lsp", "", false, "Start language server")
	err := daemonCmd.Flags().MarkHidden("lsp")
	if err != nil {
		logger.Fatalf(true, "Unable to hide `--lsp` flag: %s", err.Error())
	}
}
