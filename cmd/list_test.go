/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"reflect"
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/runner"
)

func TestOnlyUniqueTagsAreReturned(t *testing.T) {
	listTags([]*gauge.Specification{buildTestSpecification()}, func(res []string) {
		verifyUniqueness(res, []string{"bar", "foo"}, t)
	})
}

func TestOnlyUniqueSpecsAreReturned(t *testing.T) {
	specs := []*gauge.Specification{
		buildTestSpecification(),
		buildTestSpecification(),
	}
	listSpecifications(specs, func(res []string) {
		verifyUniqueness(res, []string{"Spec1"}, t)
	})
}

func TestOnlyUniqueSceanriosAreReturned(t *testing.T) {
	listScenarios([]*gauge.Specification{buildTestSpecification()}, func(res []string) {
		verifyUniqueness(res, []string{"scenario1"}, t)
	})
}

func TestOnlyUniqueStepsAreReturned(t *testing.T) {
	connectToRunner = func() (runner.Runner, error) {
		return &mockRunner{
			response: &gauge_messages.Message{
				StepNamesResponse: &gauge_messages.StepNamesResponse{
					Steps: []string{
						"Vowels are <vowelString>.",
						"Vowels in English language are <vowelString>.",
						"The word <word> has <expectedCount> vowels.",
						"Almost all words have vowels <wordsTable>",
					},
				},
			},
		}, nil
	}
	expected := []string{
		"Almost all words have vowels <wordsTable>",
		"The word <word> has <expectedCount> vowels.",
		"Vowels are <vowelString>.",
		"Vowels in English language are <vowelString>.",
	}
	listSteps([]*gauge.Specification{buildTestSpecification()}, func(res []string) {
		verifyUniqueness(res, expected, t)
	})
}

func buildTestSpecification() *gauge.Specification {
	return &gauge.Specification{
		Heading: &gauge.Heading{
			Value: "Spec1",
		},
		Scenarios: []*gauge.Scenario{
			{
				Heading: &gauge.Heading{
					Value: "scenario1",
				},
				Tags: &gauge.Tags{
					RawValues: [][]string{{"foo"}, {"bar"}},
				},
				Steps: []*gauge.Step{
					{
						Value:     "scenario1#step1",
						LineText:  "not important",
						IsConcept: false,
					},
				}},
			{
				Heading: &gauge.Heading{
					Value: "scenario1",
				},
				Tags: &gauge.Tags{
					RawValues: [][]string{{"foo"}},
				},
				Steps: []*gauge.Step{
					{
						Value:     "scenario2#step1",
						LineText:  "not important",
						IsConcept: false,
					},
				}},
		},
		TearDownSteps: []*gauge.Step{},
	}
}

func verifyUniqueness(actual []string, wanted []string, t *testing.T) {
	if !reflect.DeepEqual(actual, wanted) {
		t.Errorf("wanted: `%s`,\n got: `%s` ", wanted, actual)
	}
}
