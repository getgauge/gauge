package cmd

import (
	"errors"
	"fmt"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/spf13/cobra"
)

var (
	GaugeCmd = &cobra.Command{
		Use:   "gauge",
		Short: "Gauge is a light-weight open-source cross-platform test automation tool.",
		Long: `Gauge is a light-weight cross-platform test automation tool with the ability to author test cases in the business language.

Complete documentation is available at https://docs.getgauge.io/.`,
		Example: `  gauge run specs/
  gauge run -p specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cmd.Help()
			}
		},
	}
	logLevel string
	dir      string
)

func init() {
	GaugeCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set level of logging to debug, info, warning, error or critical.")
	GaugeCmd.PersistentFlags().StringVarP(&dir, "dir", "d", ".", "Set the working directory for the current command, accepts a path relative to current directory.")
}

type commandWriter struct {
	bytes []byte
}

func (w *commandWriter) Write(p []byte) (int, error) {
	w.bytes = append(w.bytes, p...)
	return len(p), nil
}

func (w *commandWriter) display() {
	fmt.Print(string(w.bytes))
}

func Parse() (int, error) {
	writer := &commandWriter{}
	GaugeCmd.SetOutput(writer)
	if cmd, err := GaugeCmd.ExecuteC(); err != nil {
		if cmd == GaugeCmd {
			return 1, errors.New("Failed parsing using the new gauge command structure, falling back to the old usage.")
		}
		writer.display()
		return 1, nil
	}
	writer.display()
	return 0, nil
}

func getSpecsDir(args []string) []string {
	if len(args) > 0 {
		return args
	}
	return []string{common.SpecsDirectoryName}
}

func isValidGaugeProject(args []string) error {
	return config.SetProjectRoot(args)
}

func setGlobalFlags() {
	logger.Initialize(logLevel)
	logger.Debug("Gauge Install ID: %s", config.UniqueID())
	util.SetWorkingDir(dir)
}
