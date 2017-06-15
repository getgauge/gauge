package cmd

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall [flags] <plugin>",
	Short:   "Uninstalls a plugin.",
	Long:    "Uninstalls a plugin.",
	Example: `  gauge uninstall java`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
		}
		setGlobalFlags()
		track.UninstallPlugin(args[0])
		install.UninstallPlugin(args[0], pVersion)
	},
}

func init() {
	GaugeCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringVarP(&pVersion, "plugin-version", "v", "", "Version of plugin to be uninstalled.")
}
