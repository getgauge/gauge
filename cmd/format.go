package cmd

import (
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:     "format [flags] [args]",
	Short:   "Formats the specified spec files.",
	Long:    "Formats the specified spec files.",
	Example: "  gauge format specs/",
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		if err := isValidGaugeProject(args); err != nil {
			logger.Fatalf(err.Error())
		}
		track.Format()
		formatter.FormatSpecFilesIn(getSpecsDir(args)[0])
	},
}

func init() {
	GaugeCmd.AddCommand(formatCmd)
}
