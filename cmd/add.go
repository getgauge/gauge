package cmd

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:     "add [flags] <plugin>",
		Short:   "Adds the specified non-language plugin to the current project.",
		Long:    `Adds the specified non-language plugin to the current project.`,
		Example: "  gauge add xml-report",
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) < 1 {
				logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
			}
			track.AddPlugins(args[0])
			install.AddPluginToProject(args[0], pArgs)
		},
	}
	pArgs string
)

func init() {
	GaugeCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&pArgs, "plugin-args", "", "", "Specified additional arguments to the plugin.")
}
