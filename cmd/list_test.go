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
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge"
)

func TestOnlyUniqueTagsAreReturned(t *testing.T) {
	listTags([]*gauge.Specification{buildTestSpecification("Spec1")}, "", func(res []string) {
		verifyUniqueness(res, []string{"bar", "foo"}, t)
	})
}

func TestOnlyUniqueSpecsAreReturned(t *testing.T) {
	specs := []*gauge.Specification{
		buildTestSpecification("Spec1"),
		buildTestSpecification("Spec2"),
	}
	listSpecifications(specs, "", func(res []string) {
		verifyUniqueness(res, []string{"Spec1", "Spec2"}, t)
	})
}

func TestOnlyUniqueSceanriosAreReturned(t *testing.T) {
	specs := []*gauge.Specification{
		buildTestSpecification("Spec1"),
		buildTestSpecification("Spec2"),
	}
	listScenarios(specs, "", func(res []string) {
		verifyUniqueness(res, []string{"scenario1", "scenario2"}, t)
	})
}

func TestFilteredScenariosWithTag(t *testing.T) {
	specs := []*gauge.Specification{
		buildTestSpecification("Spec1"),
		buildTestSpecification("Spec2"),
	}
	listScenarios(specs, "bar", func(res []string) {
		verifyUniqueness(res, []string{"scenario1"}, t)
	})
}

func TestFilteredSpecsWithNoTagsAndNoFiltering(t *testing.T) {
	specs := []*gauge.Specification{
		buildTestSpecification("Spec1"),
		buildTestSpecification("Spec2"),
	}
	listSpecifications(specs, "", func(res []string) {
		verifyUniqueness(res, []string{"Spec1", "Spec2"}, t)
	})
}

func TestFilteredSpecsWithNoTagsAndFiltering(t *testing.T) {
	specs := []*gauge.Specification{
		buildTestSpecification("Spec1"),
		buildTestSpecification("Spec2"),
	}
	listSpecifications(specs, "blub", func(res []string) {
		verifyUniqueness(res, []string{}, t)
	})
}

func buildTestSpecification(name string) *gauge.Specification {
	return &gauge.Specification{
		Heading: &gauge.Heading{
			Value: name,
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
					Value: "scenario2",
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
