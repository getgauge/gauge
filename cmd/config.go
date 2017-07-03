package cmd

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

var (
	propertyCmd = &cobra.Command{
		Use:     "config [flags] [args]",
		Short:   "Manage global config properties",
		Long:    "Manage global config properties.",
		Example: `  gauge config check_updates false`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if list || machineReadable {
				var f config.Formatter
				f = config.TextFormatter{}
				if machineReadable {
					f = &config.JsonFormatter{}
				}
				s, err := config.MergedProperties().Format(f)
				if err != nil {
					logger.Fatalf(err.Error())
				}
				logger.Info(s)
				return
			}
			if len(args) == 0 {
				logger.Fatalf("Error: Config command needs argument(s).\n%s", cmd.UsageString())
			}
			if len(args) == 1 {
				v, err := config.GetProperty(args[0])
				if err != nil {
					logger.Fatalf(err.Error())
				}
				logger.Info(v.Value)
				return
			}
			if err := config.Update(args[0], args[1]); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}
	list bool
)

func init() {
	GaugeCmd.AddCommand(propertyCmd)
	propertyCmd.Flags().BoolVarP(&list, "list", "", false, "List all global properties")
	propertyCmd.Flags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Print all properties in JSON format")
}
