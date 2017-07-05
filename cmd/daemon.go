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
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/api/lang"
	"github.com/getgauge/gauge/execution/stream"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	daemonCmd = &cobra.Command{
		Use:     "daemon [flags] <port> [args]",
		Short:   "Run as a daemon",
		Long:    "Run as a daemon.",
		Example: "  gauge daemon 1234",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := isValidGaugeProject(args); err != nil {
				logger.Fatalf(err.Error())
			}
			if lsp {
				lang.Server(&infoGatherer.SpecInfoGatherer{SpecDirs: getSpecsDir(args)}).Start()
				return
			}
			track.Daemon()
			stream.Start()
			port := ""
			specs := args
			if len(args) > 0 {
				port = args[0]
				specs = getSpecsDir(args[1:])
			}
			api.RunInBackground(port, specs)
		},
	}
	lsp bool
)

func init() {
	GaugeCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().BoolVarP(&lsp, "lsp", "", false, "Start language server")
	daemonCmd.Flags().MarkHidden("lsp")
}
