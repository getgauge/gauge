package cmd

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

var (
	propertyCmd = &cobra.Command{
		Use:     "property [flags] [args]",
		Short:   "Manage global properties.",
		Long:    "Manage global properties.",
		Example: `  gauge property updates false`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if len(args) < 2 {
				logger.Fatalf("Error: Property command needs at least two arguments.\n%s", cmd.UsageString())
			}
			if err := config.Update(args[0], args[1]); err != nil {
				logger.Fatalf(err.Error())
			}
		},
	}
)

func init() {
	GaugeCmd.AddCommand(propertyCmd)
	propertyCmd.Flags().BoolVarP(&all, "all", "a", false, "List all global properties")
	propertyCmd.Flags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Print all properties in JSON format")
}
