package cmd

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update [flags] <plugin>",
		Short: "Updates a plugin.",
		Long:  "Updates a plugin.",
		Example: `  gauge update java
  gauge update -a
  gauge update -c`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if all {
				track.UpdateAll()
				install.UpdatePlugins()
				return
			} else if check {
				track.CheckUpdates()
				install.PrintUpdateInfoWithDetails()
				return
			}
			if len(args) < 1 {
				logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
			}
			track.Update(args[0])
			install.HandleUpdateResult(install.InstallPlugin(args[0], pVersion), args[0], true)
		},
	}
	check bool
)

func init() {
	GaugeCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&all, "all", "a", false, "Updates all the installed Gauge plugins.")
	updateCmd.Flags().BoolVarP(&check, "check", "c", false, "Checks for Gauge and plugins updates.")
}
