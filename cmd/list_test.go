/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge"
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

func buildTestSpecification() *gauge.Specification {
	return &gauge.Specification{
		Heading: &gauge.Heading{
			Value: "Spec1",
		},
		Scenarios: []*gauge.Scenario{
			&gauge.Scenario{
				Heading: &gauge.Heading{
					Value: "scenario1",
				},
				Tags: &gauge.Tags{
					RawValues: [][]string{{"foo"}, {"bar"}},
				},
				Steps: []*gauge.Step{
					&gauge.Step{
						Value:     "scenario1#step1",
						LineText:  "not important",
						IsConcept: false,
					},
				}},
			&gauge.Scenario{
				Heading: &gauge.Heading{
					Value: "scenario1",
				},
				Tags: &gauge.Tags{
					RawValues: [][]string{{"foo"}},
				},
				Steps: []*gauge.Step{
					&gauge.Step{
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
