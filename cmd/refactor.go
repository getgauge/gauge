package cmd

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var refactorCmd = &cobra.Command{
	Use:     "refactor",
	Short:   "Refactor steps.",
	Long:    "Refactor steps.",
	Example: `  gauge refactor "old step" "new step"`,
	Run: func(cmd *cobra.Command, args []string) {
		setGlobalFlags()
		if len(args) < 2 {
			logger.Fatalf("Error: Refactor command needs at least two arguments.\n%s", cmd.UsageString())

		}
		if err := isValidGaugeProject(args); err != nil {
			logger.Fatalf(err.Error())
		}
		track.Refactor()
		refactorInit(args)
	},
}

func init() {
	GaugeCmd.AddCommand(refactorCmd)
}

func refactorInit(args []string) {
	var specDirs = []string{common.SpecsDirectoryName}
	if len(args) > 1 {
		specDirs = args[1:]
	}
	startChan := api.StartAPI(false)
	refactor.RefactorSteps(args[0], args[1], startChan, specDirs)
}
