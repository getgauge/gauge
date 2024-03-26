/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/validation"
	"github.com/spf13/cobra"
)

const (
	hideSuggestionDefault = false
	hideSuggestionName    = "hide-suggestion"
)

var (
	validateCmd = &cobra.Command{
		Use:   "validate [flags] [args]",
		Short: "Check for validation and parse errors",
		Long:  `Check for validation and parse errors.`,
		Example: `  gauge validate specs/
  gauge validate --env test specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			validation.HideSuggestion = hideSuggestion
			if err := config.SetProjectRoot(args); err != nil {
				exit(err, cmd.UsageString())
			}
			loadEnvAndReinitLogger(cmd)
			installMissingPlugins(installPlugins, true)
			validation.Validate(args)
		},
		DisableAutoGenTag: true,
	}
	hideSuggestion bool
)

func init() {
	GaugeCmd.AddCommand(validateCmd)
	flags := validateCmd.Flags()
	flags.BoolVarP(&hideSuggestion, "hide-suggestion", "", false, "Prints a step implementation stub for every unimplemented step")
	flags.StringVarP(&environment, environmentName, "e", environmentDefault, "Specifies the environment to use")
}
