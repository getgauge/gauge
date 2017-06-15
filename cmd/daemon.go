package cmd

import (
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/execution/stream"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	daemonCmd = &cobra.Command{
		Use:     "daemon",
		Short:   "Run as a daemon.",
		Long:    "Run as a daemon.",
		Example: "  gauge daemon 1234",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := isValidGaugeProject(args); err != nil {
				logger.Fatalf(err.Error())
			}
			track.Daemon()
			stream.Start()
			port := ""
			if len(args) > 0 {
				port = args[0]
			}
			api.RunInBackground(port, getSpecsDir(args))
		},
	}
)

func init() {
	GaugeCmd.AddCommand(daemonCmd)
}
