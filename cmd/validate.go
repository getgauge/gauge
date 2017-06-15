package cmd

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/getgauge/gauge/validation"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Check for validation and parse errors.",
	Long:    "Check for validation and parse errors.",
	Example: "  gauge validate specs/",
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		if err := isValidGaugeProject(args); err != nil {
			logger.Fatalf(err.Error())
		}
		track.Validation()
		validation.Validate(args)
	},
}

func init() {
	GaugeCmd.AddCommand(validateCmd)
}
