package cmd

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/projectInit"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init [flags] <template>",
	Short:   "Initializes project structure in the current directory.",
	Long:    "Initializes project structure in the current directory.",
	Example: "  gauge init java",
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		if len(args) < 1 {
			logger.Fatalf("Error: Missing argument <template name>. To see all the templates, run 'gauge list-templates'.\n%s", cmd.UsageString())
		}
		track.ProjectInit()
		projectInit.InitializeProject(args[0])
	},
}

func init() {
	GaugeCmd.AddCommand(initCmd)
}
