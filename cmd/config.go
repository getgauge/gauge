/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"errors"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:     "config [flags] [args]",
		Short:   "Change global configurations",
		Long:    `Change global configurations.`,
		Example: `  gauge config check_updates false`,
		Run: func(cmd *cobra.Command, args []string) {
			if list || machineReadable {
				text, err := config.List(machineReadable)
				if err != nil {
					logger.Fatal(true, err.Error())
				}
				logger.Info(true, text)
				return
			}
			if len(args) == 0 {
				exit(errors.New("Config command needs argument(s)."), cmd.UsageString())
			}
			if len(args) == 1 {
				text, err := config.GetProperty(args[0])
				if err != nil {
					logger.Fatal(true, err.Error())
				}
				logger.Info(true, text)
				return
			}
			if err := config.Update(args[0], args[1]); err != nil {
				logger.Fatal(true, err.Error())
			}
		},
		DisableAutoGenTag: true,
	}
	list bool
)

func init() {
	GaugeCmd.AddCommand(configCmd)
	configCmd.Flags().BoolVarP(&list, "list", "", false, "List all global properties")
	configCmd.Flags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Print all properties in JSON format")
}
