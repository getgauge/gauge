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
	"os"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/api/lang"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
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
			if e := env.LoadEnv(environment); e != nil {
				logger.Fatalf(true, e.Error())
			}
			os.Setenv(isDaemon, "true")
			if err := config.SetProjectRoot(args); err != nil {
				exit(err, cmd.UsageString())
			}
			if lsp {
				track.Lsp()
				lang.Start(&infoGatherer.SpecInfoGatherer{SpecDirs: getSpecsDir(args)}, logLevel)
				return
			}
			track.Daemon()
			port := ""
			specs := util.GetSpecDirs()
			if len(args) > 0 {
				port = args[0]
				specs = getSpecsDir(args[1:])
			}
			api.RunInBackground(port, specs)
		},
		DisableAutoGenTag: true,
	}
	lsp bool
)

func init() {
	GaugeCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().BoolVarP(&lsp, "lsp", "", false, "Start language server")
	daemonCmd.Flags().MarkHidden("lsp")
}
