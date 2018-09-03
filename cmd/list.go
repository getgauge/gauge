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
	"fmt"
	supersort "sort"

	"github.com/getgauge/gauge/track"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:     "list [flags] [args]",
		Short:   "List specifications, scenarios or tags for a gauge project",
		Long:    `List specifications, scenarios or tags for a gauge project`,
		Example: `  gauge list --tags specs`,
		Run: func(cmd *cobra.Command, args []string) {
			specs, failed := parser.ParseSpecs(getSpecsDir(args), gauge.NewConceptDictionary(), gauge.NewBuildErrors())
			if failed {
				return
			}
			if specsFlag {
				logger.Info(true, "[Specifications]")
				track.ListSpecifications()
				listSpecifications(specs, taggedWithFlag, print)
			}
			if scenariosFlag {
				logger.Info(true, "[Scenarios]")
				track.ListScenarios()
				listScenarios(specs, taggedWithFlag, print)
			}
			if tagsFlag {
				logger.Info(true, "[Tags]")
				track.ListTags()
				listTags(specs, taggedWithFlag, print)
			}
			if !specsFlag && !scenariosFlag && !tagsFlag {
				exit(fmt.Errorf("Missing flag, nothing to list"), cmd.UsageString())
			}
		},
		DisableAutoGenTag: true,
	}
	tagsFlag       bool
	specsFlag      bool
	scenariosFlag  bool
	taggedWithFlag string
)

func init() {
	GaugeCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&tagsFlag, "tags", "", false, "List the tags in projects")
	listCmd.Flags().BoolVarP(&specsFlag, "specs", "", false, "List the specifications in projects")
	listCmd.Flags().BoolVarP(&scenariosFlag, "scenarios", "", false, "List the scenarios in projects")
	listCmd.Flags().StringVarP(&taggedWithFlag, "taggedWith", "", "", "Filters all elements that contain the given tag")
}

type handleResult func([]string)

func print(res []string) {
	for _, element := range res {
		logger.Infof(true, element)
	}
}

func listTags(s []*gauge.Specification, taggedWith string, f handleResult) {
	allTags := []string{}
	for _, spec := range s {
		allTags = appendTags(allTags, spec.Tags, taggedWith)
		for _, scenario := range spec.Scenarios {
			allTags = appendTags(allTags, scenario.Tags, taggedWith)
		}
	}
	f(sortedDistinctElements(allTags))
}

func listScenarios(s []*gauge.Specification, taggedWith string, f handleResult) {
	allScenarios := []string{}
	for _, spec := range s {
		for _, scenario := range spec.Scenarios {
			if contains(scenario.Tags, taggedWith) {
				allScenarios = append(allScenarios, scenario.Heading.Value)
			}
		}
	}
	f(sortedDistinctElements(allScenarios))
}

func listSpecifications(s []*gauge.Specification, taggedWith string, f handleResult) {
	allSpecs := []string{}
	for _, spec := range s {
		if contains(spec.Tags, taggedWith) {
			allSpecs = append(allSpecs, spec.Heading.Value)
		}
	}
	f(sortedDistinctElements(allSpecs))
}

func sortedDistinctElements(s []string) []string {
	unique := uniqueNonEmptyElementsOf(s)
	supersort.Strings(unique)
	return unique
}

func appendTags(s []string, tags *gauge.Tags, taggedWith string) []string {
	if tags != nil {
		if contains(tags, taggedWith) {
			s = append(s, tags.Values()...)
		}
	}
	return s
}

func contains(tags *gauge.Tags, filter string) bool {
	if len(filter) == 0 {
		return true
	} else if tags == nil {
		return false
	}
	for _, theTag := range tags.Values() {
		if theTag == filter {
			return true
		}
	}
	return false
}

func uniqueNonEmptyElementsOf(input []string) []string {
	unique := make(map[string]bool, len(input))
	us := make([]string, len(unique))
	for _, elem := range input {
		if len(elem) != 0 && !unique[elem] {
			us = append(us, elem)
			unique[elem] = true
		}
	}

	return us

}
