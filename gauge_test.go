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

package main

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestToCheckIfItsIndexedSpec(c *C) {
	c.Assert(isIndexedSpec("specs/hello_world:as"), Equals, false)
	c.Assert(isIndexedSpec("specs/hello_world.spec:0"), Equals, true)
	c.Assert(isIndexedSpec("specs/hello_world.spec:78809"), Equals, true)
	c.Assert(isIndexedSpec("specs/hello_world.spec:09"), Equals, true)
	c.Assert(isIndexedSpec("specs/hello_world.spec:09sa"), Equals, false)
	c.Assert(isIndexedSpec("specs/hello_world.spec:09090"), Equals, true)
	c.Assert(isIndexedSpec("specs/hello_world.spec"), Equals, false)
	c.Assert(isIndexedSpec("specs/hello_world.spec:"), Equals, false)
	c.Assert(isIndexedSpec("specs/hello_world.md"), Equals, false)
}

func (s *MySuite) TestToObtainIndexedSpecName(c *C) {
	specName, scenarioNum := GetIndexedSpecName("specs/hello_world.spec:67")
	c.Assert(specName, Equals, "specs/hello_world.spec")
	c.Assert(scenarioNum, Equals, 67)
}
func (s *MySuite) TestToObtainIndexedSpecName1(c *C) {
	specName, scenarioNum := GetIndexedSpecName("hello_world.spec:67342")
	c.Assert(specName, Equals, "hello_world.spec")
	c.Assert(scenarioNum, Equals, 67342)
}

func (s *MySuite) TestToCheckTagsInSpecLevel(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: tagKind, args: []string{"tag1", "tag2"}, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
	}

	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))

	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.tags.values), Equals, 2)
	c.Assert(spec.tags.values[0], Equals, "tag1")
	c.Assert(spec.tags.values[1], Equals, "tag2")
}

func (s *MySuite) TestToCheckTagsInScenarioLevel(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1", "tag2"}, lineNo: 3},
	}

	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))

	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.scenarios[0].tags.values), Equals, 2)
	c.Assert(spec.scenarios[0].tags.values[0], Equals, "tag1")
	c.Assert(spec.scenarios[0].tags.values[1], Equals, "tag2")
}

func (s *MySuite) TestToSplitTagNames(c *C) {
	allTags := splitAndTrimTags("tag1 , tag2,   tag3")
	c.Assert(allTags[0], Equals, "tag1")
	c.Assert(allTags[1], Equals, "tag2")
	c.Assert(allTags[2], Equals, "tag3")
}

func (s *MySuite) TestToFilterScenariosByTag(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 5},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.scenarios), Equals, 3)
	c.Assert(len(spec.scenarios[1].tags.values), Equals, 2)

	var specs []*specification
	specs = append(specs, spec)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].scenarios), Equals, 1)
	c.Assert(specs[0].scenarios[0].heading.value, Equals, "Scenario Heading 2")
}

func (s *MySuite) TestToFilterScenariosByUnavailableTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 5},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.scenarios), Equals, 3)
	c.Assert(len(spec.scenarios[1].tags.values), Equals, 2)

	var specs []*specification
	specs = append(specs, spec)
	specs = filterSpecsByTags(specs, "tag3")
	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestToFilterMultipleScenariosByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 5},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 6},
		&token{kind: tagKind, args: myTags, lineNo: 7},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec)
	c.Assert(len(specs[0].scenarios), Equals, 3)
	c.Assert(len(specs[0].scenarios[0].tags.values), Equals, 1)
	c.Assert(len(specs[0].scenarios[1].tags.values), Equals, 2)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].scenarios), Equals, 2)
	c.Assert(specs[0].scenarios[0].heading.value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].scenarios[1].heading.value, Equals, "Scenario Heading 3")
}

func (s *MySuite) TestToFilterMultipleScenariosByMultipleTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 5},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 6},
		&token{kind: tagKind, args: myTags, lineNo: 7},
		&token{kind: scenarioKind, value: "Scenario Heading 4", lineNo: 8},
		&token{kind: tagKind, args: []string{"prod", "tag7", "tag1", "tag2"}, lineNo: 9},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec)

	c.Assert(len(specs[0].scenarios), Equals, 4)
	c.Assert(len(specs[0].scenarios[0].tags.values), Equals, 1)
	c.Assert(len(specs[0].scenarios[1].tags.values), Equals, 2)
	c.Assert(len(specs[0].scenarios[2].tags.values), Equals, 2)
	c.Assert(len(specs[0].scenarios[3].tags.values), Equals, 4)

	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].scenarios), Equals, 3)
	c.Assert(specs[0].scenarios[0].heading.value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].scenarios[1].heading.value, Equals, "Scenario Heading 3")
	c.Assert(specs[0].scenarios[2].heading.value, Equals, "Scenario Heading 4")
}

func (s *MySuite) TestToFilterScenariosByTagsAtSpecLevel(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 5},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec)

	c.Assert(len(specs[0].scenarios), Equals, 3)
	c.Assert(len(specs[0].tags.values), Equals, 2)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].scenarios), Equals, 3)
	c.Assert(specs[0].scenarios[0].heading.value, Equals, "Scenario Heading 1")
	c.Assert(specs[0].scenarios[1].heading.value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].scenarios[2].heading.value, Equals, "Scenario Heading 3")
}

func (s *MySuite) TestToFilterSpecsByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 2},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens2 := []*token{
		&token{kind: specKind, value: "Spec Heading3", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec3, result := new(specParser).createSpecification(tokens2, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec1.scenarios), Equals, 2)
	c.Assert(len(spec1.scenarios[0].tags.values), Equals, 2)
	c.Assert(len(spec2.scenarios), Equals, 2)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 2)
	c.Assert(len(specs[0].scenarios), Equals, 1)
	c.Assert(len(specs[1].scenarios), Equals, 1)
	c.Assert(specs[0].heading.value, Equals, "Spec Heading1")
	c.Assert(specs[1].heading.value, Equals, "Spec Heading3")
}

func (s *MySuite) TestToSortSpecs(c *C) {
	spec1 := &specification{fileName: "ab"}
	spec2 := &specification{fileName: "b"}
	spec3 := &specification{fileName: "c"}
	var specs []*specification
	specs = append(specs, spec3)
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	specs = sortSpecsList(specs)

	c.Assert(specs[0].fileName, Equals, spec1.fileName)
	c.Assert(specs[1].fileName, Equals, spec2.fileName)
	c.Assert(specs[2].fileName, Equals, spec3.fileName)
}

func (s *MySuite) TestToFilterSpecsByTagExpOfTwoTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].tags.values[0], Equals, myTags[0])
	c.Assert(specs[0].tags.values[1], Equals, myTags[1])
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 1)
}

func (s *MySuite) TestToEvaluateTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 01", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 02", lineNo: 4},
		&token{kind: tagKind, args: []string{"tag3"}, lineNo: 5},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	specs = filterSpecsByTags(specs, "tag1 & !(tag1 & tag4) & (tag2 | tag3)")
	c.Assert(len(specs), Equals, 1)
	c.Assert(len(specs[0].scenarios), Equals, 2)
	c.Assert(specs[0].scenarios[0].heading.value, Equals, "Scenario Heading 1")
	c.Assert(specs[0].scenarios[1].heading.value, Equals, "Scenario Heading 2")
}

func (s *MySuite) TestToFilterSpecsByWrongTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].tags.values[0], Equals, myTags[0])
	c.Assert(specs[0].tags.values[1], Equals, myTags[1])
	specs = filterSpecsByTags(specs, "(tag1 & tag2")
	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestToShuffleSpecsToRandomize(c *C) {
	var specs []*specification
	specs = append(specs, &specification{fileName: "a"}, &specification{fileName: "b"}, &specification{fileName: "c"}, &specification{fileName: "d"},
		&specification{fileName: "e"}, &specification{fileName: "f"}, &specification{fileName: "g"}, &specification{fileName: "h"})
	shuffledSpecs := shuffleSpecs(specs)
	for i, spec := range shuffledSpecs {
		if spec.fileName != specs[i].fileName {
			c.Succeed()
		}
	}
}

func (s *MySuite) TestToRunSpecificSetOfSpecs(c *C) {
	spec1 := &specification{heading: &heading{value: "SPECHEADING1"}}
	spec2 := &specification{heading: &heading{value: "SPECHEADING2"}}
	heading3 := &heading{value: "SPECHEADING3"}
	spec3 := &specification{heading: heading3}
	spec4 := &specification{heading: &heading{value: "SPECHEADING4"}}
	spec5 := &specification{heading: &heading{value: "SPECHEADING5"}}
	spec6 := &specification{heading: &heading{value: "SPECHEADING6"}}
	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = append(specs, spec4)
	specs = append(specs, spec5)
	specs = append(specs, spec6)

	value := 6
	value1 := 3
	numberOfExecutionStreams = &value
	distribute = &value1

	specsToExecute := applyFlags(specs)

	c.Assert(len(specsToExecute), Equals, 1)
	c.Assert(specsToExecute[0].heading, Equals, heading3)

}

func (s *MySuite) TestToRunSpecificSetOfSpecsGivesSameSpecsEverytime(c *C) {
	spec1 := &specification{heading: &heading{value: "SPECHEADING1"}}
	spec2 := &specification{heading: &heading{value: "SPECHEADING2"}}
	spec3 := &specification{heading: &heading{value: "SPECHEADING3"}}
	spec4 := &specification{heading: &heading{value: "SPECHEADING4"}}
	heading5 := &heading{value: "SPECHEADING5"}
	spec5 := &specification{heading: heading5}
	heading6 := &heading{value: "SPECHEADING6"}
	spec6 := &specification{heading: heading6}
	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = append(specs, spec4)
	specs = append(specs, spec5)
	specs = append(specs, spec6)

	value := 3
	numberOfExecutionStreams = &value
	distribute = &value
	specsToExecute1 := applyFlags(specs)
	c.Assert(len(specsToExecute1), Equals, 2)

	specsToExecute2 := applyFlags(specs)
	c.Assert(len(specsToExecute2), Equals, 2)

	specsToExecute3 := applyFlags(specs)
	c.Assert(len(specsToExecute3), Equals, 2)

	c.Assert(specsToExecute2[0].heading, Equals, specsToExecute1[0].heading)
	c.Assert(specsToExecute2[1].heading, Equals, specsToExecute1[1].heading)
	c.Assert(specsToExecute3[0].heading, Equals, specsToExecute1[0].heading)
	c.Assert(specsToExecute3[1].heading, Equals, specsToExecute1[1].heading)
}

func (s *MySuite) TestToRunSpecificSetOfSpecsGivesEmptySpecsIfDistributableNumberIsNotValid(c *C) {
	spec1 := &specification{heading: &heading{value: "SPECHEADING1"}}
	var specs []*specification
	specs = append(specs, spec1)

	value := 1
	value1 := 3
	numberOfExecutionStreams = &value
	distribute = &value1
	specsToExecute1 := applyFlags(specs)
	c.Assert(len(specsToExecute1), Equals, 0)

	value = 1
	value1 = -3
	numberOfExecutionStreams = &value
	distribute = &value1
	specsToExecute1 = applyFlags(specs)
	c.Assert(len(specsToExecute1), Equals, 0)
}
