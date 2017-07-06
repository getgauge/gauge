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

package parser

import (
	"path/filepath"

	"strings"

	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestSimpleStepAfterStepValueExtraction(c *C) {
	stepText := "a simple step"
	stepValue, err := ExtractStepValueAndParams(stepText, false)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 0)
	c.Assert(stepValue.StepValue, Equals, "a simple step")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a simple step")
}

func (s *MySuite) TestStepWithColonAfterStepValueExtraction(c *C) {
	stepText := "a : simple step \"hello\""
	stepValue, err := ExtractStepValueAndParams(stepText, false)
	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(stepValue.StepValue, Equals, "a : simple step {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a : simple step <hello>")
}

func (s *MySuite) TestSimpleStepAfterStepValueExtractionForStepWithAParam(c *C) {
	stepText := "Comment <a>"
	stepValue, err := ExtractStepValueAndParams(stepText, false)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(stepValue.StepValue, Equals, "Comment {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "Comment <a>")
}

func (s *MySuite) TestAddingTableParamAfterStepValueExtraction(c *C) {
	stepText := "a simple step"
	stepValue, err := ExtractStepValueAndParams(stepText, true)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(args[0], Equals, string(gauge.TableArg))
	c.Assert(stepValue.StepValue, Equals, "a simple step {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a simple step <table>")
}

func (s *MySuite) TestAddingTableParamAfterStepValueExtractionForStepWithExistingParam(c *C) {
	stepText := "a \"param1\" step with multiple params <param2> <file:specialParam>"
	stepValue, err := ExtractStepValueAndParams(stepText, true)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 4)
	c.Assert(args[0], Equals, "param1")
	c.Assert(args[1], Equals, "param2")
	c.Assert(args[2], Equals, "file:specialParam")
	c.Assert(args[3], Equals, "table")
	c.Assert(stepValue.StepValue, Equals, "a {} step with multiple params {} {} {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a <param1> step with multiple params <param2> <file:specialParam> <table>")
}

func (s *MySuite) TestAfterStepValueExtractionForStepWithExistingParam(c *C) {
	stepText := "a \"param1\" step with multiple params <param2> <file:specialParam>"
	stepValue, err := ExtractStepValueAndParams(stepText, false)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 3)
	c.Assert(args[0], Equals, "param1")
	c.Assert(args[1], Equals, "param2")
	c.Assert(args[2], Equals, "file:specialParam")
	c.Assert(stepValue.StepValue, Equals, "a {} step with multiple params {} {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a <param1> step with multiple params <param2> <file:specialParam>")
}

func (s *MySuite) TestCreateStepValueFromStep(c *C) {
	step := &gauge.Step{Value: "simple step with {} and {}", Args: []*gauge.StepArg{staticArg("hello"), dynamicArg("desc")}}
	stepValue := CreateStepValue(step)

	args := stepValue.Args
	c.Assert(len(args), Equals, 2)
	c.Assert(args[0], Equals, "hello")
	c.Assert(args[1], Equals, "desc")
	c.Assert(stepValue.StepValue, Equals, "simple step with {} and {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "simple step with <hello> and <desc>")
}

func (s *MySuite) TestCreateStepValueFromStepWithSpecialParams(c *C) {
	step := &gauge.Step{Value: "a step with {}, {} and {}", Args: []*gauge.StepArg{specialTableArg("hello"), specialStringArg("file:user.txt"), tableArgument()}}
	stepValue := CreateStepValue(step)

	args := stepValue.Args
	c.Assert(len(args), Equals, 3)
	c.Assert(args[0], Equals, "hello")
	c.Assert(args[1], Equals, "file:user.txt")
	c.Assert(args[2], Equals, "table")
	c.Assert(stepValue.StepValue, Equals, "a step with {}, {} and {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a step with <hello>, <file:user.txt> and <table>")
}

func (s *MySuite) TestSpecsFormArgsForMultipleIndexedArgsForOneSpec(c *C) {
	specs, _ := parseSpecsInDirs(gauge.NewConceptDictionary(), []string{filepath.Join("testdata", "sample.spec:3"), filepath.Join("testdata", "sample.spec:6")}, gauge.NewBuildErrors())

	c.Assert(len(specs), Equals, 1)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
}

func (s *MySuite) TestSpecsFromArgsMaintainsOrderOfSpecsPassed(c *C) {
	sampleSpec := filepath.Join("testdata", "sample.spec")
	sample2Spec := filepath.Join("testdata", "sample2.spec")
	specs, _ := parseSpecsInDirs(gauge.NewConceptDictionary(), []string{sample2Spec, sampleSpec}, gauge.NewBuildErrors())

	c.Assert(len(specs), Equals, 2)
	c.Assert(specs[0].Heading.Value, Equals, "Sample 2")
	c.Assert(specs[1].Heading.Value, Equals, "Sample")
}

func (s *MySuite) TestGetAllSpecsMaintainsOrderOfSpecs(c *C) {
	sample2Spec := filepath.Join("testdata", "sample2.spec")
	sampleSpec := filepath.Join("testdata", "sample.spec")
	givenSpecs, indexedSpecs := getAllSpecFiles([]string{sample2Spec, sampleSpec})

	c.Assert(len(givenSpecs), Equals, 2)
	c.Assert(len(indexedSpecs), Equals, 2)

	if !strings.HasSuffix(givenSpecs[0], sample2Spec) {
		c.Fatalf("%s file order has changed", sample2Spec)
	}
	if !strings.HasSuffix(givenSpecs[1], sampleSpec) {
		c.Fatalf("%s file order has changed", sampleSpec)
	}

	if !strings.HasSuffix(indexedSpecs[0].filePath, sample2Spec) {
		c.Fatalf("%s file order has changed", sample2Spec)
	}
	c.Assert(len(indexedSpecs[0].indices), Equals, 0)

	if !strings.HasSuffix(indexedSpecs[1].filePath, sampleSpec) {
		c.Fatalf("%s file order has changed", sampleSpec)
	}
	c.Assert(len(indexedSpecs[1].indices), Equals, 0)
}

func (s *MySuite) TestGetAllSpecsAddIndicesForIndexedSpecs(c *C) {
	file := filepath.Join("testdata", "sample.spec")
	_, indexedSpecs := getAllSpecFiles([]string{file + ":1", file + ":5"})

	c.Assert(len(indexedSpecs), Equals, 1)

	if !strings.HasSuffix(indexedSpecs[0].filePath, file) {
		c.Fatalf("%s file order has changed", file)
	}
	c.Assert(len(indexedSpecs[0].indices), Equals, 2)
	c.Assert(indexedSpecs[0].indices[0], Equals, 1)
	c.Assert(indexedSpecs[0].indices[1], Equals, 5)
}

func (s *MySuite) TestGetAllSpecsShouldDeDuplicateSpecs(c *C) {
	sampleSpec := filepath.Join("testdata", "sample.spec")
	sample2Spec := filepath.Join("testdata", "sample2.spec")

	_, indexedSpecs := getAllSpecFiles([]string{sampleSpec, sample2Spec, sampleSpec, sample2Spec + ":2"})

	c.Assert(len(indexedSpecs), Equals, 2)

	if !strings.HasSuffix(indexedSpecs[0].filePath, sampleSpec) {
		c.Fatalf("%s file order has changed", sampleSpec)
	}

	if !strings.HasSuffix(indexedSpecs[1].filePath, sample2Spec) {
		c.Fatalf("%s file order has changed", sample2Spec)
	}
}

func (s *MySuite) TestGetAllSpecsShouldDeDuplicateIndexedSpecs(c *C) {
	sampleSpec := filepath.Join("testdata", "sample.spec")
	sample2Spec := filepath.Join("testdata", "sample2.spec")

	_, indexedSpecs := getAllSpecFiles([]string{sampleSpec + ":2", sample2Spec, sampleSpec})

	c.Assert(len(indexedSpecs), Equals, 2)

	if !strings.HasSuffix(indexedSpecs[0].filePath, sampleSpec) {
		c.Fatalf("%s file order has changed", sampleSpec)
	}

	if !strings.HasSuffix(indexedSpecs[1].filePath, sample2Spec) {
		c.Fatalf("%s file order has changed", sample2Spec)
	}
}

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
	specName, scenarioNum := getIndexedSpecName("specs/hello_world.spec:67")
	c.Assert(specName, Equals, "specs/hello_world.spec")
	c.Assert(scenarioNum, Equals, 67)
}
func (s *MySuite) TestToObtainIndexedSpecName1(c *C) {
	specName, scenarioNum := getIndexedSpecName("hello_world.spec:67342")
	c.Assert(specName, Equals, "hello_world.spec")
	c.Assert(scenarioNum, Equals, 67342)
}

func (s *MySuite) TestGetIndex(c *C) {
	c.Assert(getIndex("hello.spec:67"), Equals, 10)
	c.Assert(getIndex("specs/hello.spec:67"), Equals, 16)
	c.Assert(getIndex("specs\\hello.spec:67"), Equals, 16)
	c.Assert(getIndex(":67"), Equals, 0)
	c.Assert(getIndex(""), Equals, 0)
	c.Assert(getIndex("foo"), Equals, 0)
	c.Assert(getIndex(":"), Equals, 0)
	c.Assert(getIndex("f:7a.spec:9"), Equals, 9)
}

func staticArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.Static, Value: val}
}

func dynamicArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.Dynamic, Value: val}
}

func tableArgument() *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.TableArg}
}

func specialTableArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.SpecialTable, Name: val}
}

func specialStringArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.SpecialString, Name: val}
}
