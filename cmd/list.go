/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"fmt"
	supersort "sort"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"

	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:     "list [flags] [args]",
		Short:   "List specifications, scenarios, steps or tags for a gauge project",
		Long:    `List specifications, scenarios, steps or tags for a gauge project`,
		Example: `  gauge list --tags specs`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.SetProjectRoot(args); err != nil {
				exit(err, cmd.UsageString())
			}
			loadEnvAndReinitLogger(cmd)
			specs, failed := parser.ParseSpecs(getSpecsDir(args), gauge.NewConceptDictionary(), gauge.NewBuildErrors())
			if failed {
				return
			}
			if specsFlag {
				logger.Info(true, "[Specifications]")
				listSpecifications(specs, print)
			}
			if scenariosFlag {
				logger.Info(true, "[Scenarios]")
				listScenarios(specs, print)
			}
			if tagsFlag {
				logger.Info(true, "[Tags]")
				listTags(specs, print)
			}
			if stepsFlag {
				logger.Info(true, "[Steps]")
				listSteps(specs, print)
			}
			if !specsFlag && !scenariosFlag && !tagsFlag && !stepsFlag {
				exit(fmt.Errorf("Missing flag, nothing to list"), cmd.UsageString())
			}
		},
		DisableAutoGenTag: true,
	}
	tagsFlag      bool
	specsFlag     bool
	scenariosFlag bool
	stepsFlag     bool
)

func init() {
	GaugeCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&tagsFlag, "tags", "", false, "List the tags in projects")
	listCmd.Flags().BoolVarP(&specsFlag, "specs", "", false, "List the specifications in projects")
	listCmd.Flags().BoolVarP(&scenariosFlag, "scenarios", "", false, "List the scenarios in projects")
	listCmd.Flags().BoolVarP(&stepsFlag, "steps", "", false, "List all the steps in projects (including concept steps). Does not include unused steps.")
}

type handleResult func([]string)

func print(res []string) {
	for _, element := range res {
		logger.Info(true, element)
	}
}

func listTags(s []*gauge.Specification, f handleResult) {
	allTags := []string{}
	for _, spec := range s {
		allTags = appendTags(allTags, spec.Tags)
		for _, scenario := range spec.Scenarios {
			allTags = appendTags(allTags, scenario.Tags)
		}
	}
	f(sortedDistinctElements(allTags))
}

func listScenarios(s []*gauge.Specification, f handleResult) {
	allScenarios := filter.GetAllScenarios(s)
	f(sortedDistinctElements(allScenarios))
}

func listSpecifications(s []*gauge.Specification, f handleResult) {
	allSpecs := []string{}
	for _, spec := range s {
		allSpecs = append(allSpecs, spec.Heading.Value)
	}
	f(sortedDistinctElements(allSpecs))
}

func listSteps(s []*gauge.Specification, f handleResult) {
	f(sortedDistinctElements(getImplementedStepsWithAliases()))
}

func sortedDistinctElements(s []string) []string {
	unique := uniqueNonEmptyElementsOf(s)
	supersort.Strings(unique)
	return unique
}

func appendTags(s []string, tags *gauge.Tags) []string {
	if tags != nil {
		s = append(s, tags.Values()...)
	}
	return s
}

func uniqueNonEmptyElementsOf(input []string) []string {
	unique := make(map[string]bool, len(input))
	us := make([]string, 0, len(input))
	for _, elem := range input {
		if len(elem) != 0 && !unique[elem] {
			us = append(us, elem)
			unique[elem] = true
		}
	}

	return us
}

func getImplementedStepsWithAliases() []string {
	r, err := connectToRunner()
	defer func() { _ = r.Kill() }()
	if err != nil {
		panic(err)
	}
	getAllStepsRequest := &gm.Message{MessageType: gm.Message_StepNamesRequest, StepNamesRequest: &gm.StepNamesRequest{}}
	response, err := r.ExecuteMessageWithTimeout(getAllStepsRequest)
	if err != nil {
		exit(fmt.Errorf("error while connecting to runner : %s", err.Error()), "unable to get steps from runner")
	}
	return response.GetStepNamesResponse().GetSteps()
}

var connectToRunner = func() (runner.Runner, error) {
	outFile, err := util.OpenFile(logger.ActiveLogFile)
	if err != nil {
		return nil, err
	}
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		return nil, err
	}

	return runner.StartGrpcRunner(manifest, outFile, outFile, config.IdeRequestTimeout(), false)
}
