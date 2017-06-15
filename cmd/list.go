package cmd

import (
	"github.com/getgauge/gauge/projectInit"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list-templates",
	Short:   "Lists all the Gauge templates available.",
	Long:    "Lists all the Gauge templates available.",
	Example: "  gauge list-templates",
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		track.ListTemplates()
		projectInit.ListTemplates()
	},
}

func init() {
	GaugeCmd.AddCommand(listCmd)
}
