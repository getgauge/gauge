// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/getgauge/gauge/validation"
	"github.com/spf13/cobra"
)

var (
	validateCmd = &cobra.Command{
		Use:     "validate [flags] [args]",
		Short:   "Check for validation and parse errors",
		Long:    `Check for validation and parse errors.`,
		Example: "  gauge validate specs/",
		Run: func(cmd *cobra.Command, args []string) {
			validation.HideSuggestion = hideSuggestion
			if err := config.SetProjectRoot(args); err != nil {
				logger.Fatalf(err.Error())
			}
			track.Validation(hideSuggestion)
			validation.Validate(args)
		},
		DisableAutoGenTag: true,
	}
	hideSuggestion bool
)

func init() {
	GaugeCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&hideSuggestion, "hide-suggestion", "", false, "Prints a step implementation stub for every unimplemented step")

}
