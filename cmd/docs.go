package cmd

import (
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:     "docs [flags] <plugin> [args]",
	Short:   "Generate documenation using specified plugin.",
	Long:    "Generate documenation using specified plugin.",
	Example: "  gauge docs spectacle specs/",
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		if err := isValidGaugeProject(args); err != nil {
			logger.Fatalf(err.Error())
		}
		if len(args) < 1 {
			logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
		}
		track.Docs(args[0])
		specDirs := getSpecsDir(args[1:])
		gaugeConnectionHandler := api.Start(specDirs)
		plugin.GenerateDoc(args[0], specDirs, gaugeConnectionHandler.ConnectionPortNumber())
	},
}

func init() {
	GaugeCmd.AddCommand(docsCmd)
}
