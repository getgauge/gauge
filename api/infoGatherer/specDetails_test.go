/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package infoGatherer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

const specDir = "specs"

var _ = Suite(&MySuite{})

var concept1 []byte
var concept2 []byte
var concept3 []byte
var concept4 []byte
var spec1 []byte
var spec2 []byte
var spec3 []byte
var specWithTags []byte
var spec2WithTags []byte
var specWithConcept []byte

type MySuite struct {
	specsDir   string
	projectDir string
}

func (s *MySuite) SetUpTest(c *C) {
	s.projectDir, _ = os.MkdirTemp("_testdata", "gaugeTest")
	s.specsDir, _ = createDirIn(s.projectDir, specDir)
	config.ProjectRoot = s.projectDir

	s.buildTestData()
}

func (s *MySuite) TearDownTest(c *C) {
	_ = os.RemoveAll(s.projectDir)
}

func (s *MySuite) buildTestData() {
	concept1 = make([]byte, 0)
	concept1 = append(concept1, `# foo bar
* first step with "foo"
* say "hello" to me
* a "final" step
`...)

	concept2 = make([]byte, 0)
	concept2 = append(concept2, `# bar
* first step with "foo"
* say "hello" to me
* a "final" step
`...)

	concept3 = make([]byte, 0)
	concept3 = append(concept3, `# foo bar with <param> having errors
* first step with "foo"
* say <param> to me
* a <final> step
`...)

	concept4 = make([]byte, 0)
	concept4 = append(concept4, `# foo bar with 1 step
* say hello
`...)

	spec1 = make([]byte, 0)
	spec1 = append(spec1, `Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
`...)

	spec2 = make([]byte, 0)
	spec2 = append(spec2, `Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
* say "bye" to me
`...)

	spec3 = make([]byte, 0)
	spec3 = append(spec3, `Specification Heading
=====================
|Col1|Col2|
|----|----|
|Val1|Val2|

Scenario with parse errors
----------
* say hello
* say "hello" to me
* say <bye> to me
`...)

	specWithTags = make([]byte, 0)
	specWithTags = append(specWithTags, `Specification Heading
=====================
tags:foo, bar, hello

Scenario with tags
----------
tags: simple, complex

* say hello
* say "hello" to me
* say <bye> to me
`...)

	spec2WithTags = make([]byte, 0)
	spec2WithTags = append(spec2WithTags, `Specification Heading
=====================
tags:foo, another

Scenario with tags
----------
tags: simple, complex

* say hello
* say "hello" to me
* say <bye> to me
`...)

	specWithConcept = make([]byte, 0)
	specWithConcept = append(specWithConcept, `Specification Heading
=====================
tags:foo, another

Scenario with tags
----------
tags: simple, complex

* say hello
* foo bar with 1 step
`...)

}

func (s *MySuite) TestGetParsedSpecs(c *C) {
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{specDir}}

	specFiles := util.FindSpecFilesIn(s.specsDir)
	details := specInfoGatherer.getParsedSpecs(specFiles)

	c.Assert(len(details), Equals, 1)
	c.Assert(details[0].Spec.Heading.Value, Equals, "Specification Heading")
}

func (s *MySuite) TestGetParsedSpecsForInvalidFile(c *C) {
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{specDir}}

	details := specInfoGatherer.getParsedSpecs([]string{"spec1.spec"})

	c.Assert(len(details), Equals, 1)
	c.Assert(len(details[0].Errs), Equals, 1)
	c.Assert(details[0].Errs[0].Message, Equals, "File spec1.spec doesn't exist.")
}

func (s *MySuite) TestGetParsedConcepts(c *C) {
	_, err := createFileIn(s.specsDir, "concept.cpt", concept1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.projectDir + string(filepath.Separator) + specDir}}

	conceptsMap := specInfoGatherer.getParsedConcepts()

	c.Assert(len(conceptsMap), Equals, 1)
	c.Assert(conceptsMap["foo bar"], NotNil)
	c.Assert(specInfoGatherer.conceptDictionary, NotNil)
}

func (s *MySuite) TestInitSpecsCache(c *C) {
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(1)

	specInfoGatherer.initSpecsCache()

	c.Assert(len(specInfoGatherer.specsCache.specDetails), Equals, 1)
}

func (s *MySuite) TestInitConceptsCache(c *C) {
	_, err := createFileIn(s.specsDir, "concept1.cpt", concept1)
	c.Assert(err, Equals, nil)
	_, err = createFileIn(s.specsDir, "concept2.cpt", concept2)
	c.Assert(err, Equals, nil)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.projectDir + string(filepath.Separator) + specDir}}
	specInfoGatherer.waitGroup.Add(1)

	specInfoGatherer.initConceptsCache()

	c.Assert(len(specInfoGatherer.conceptsCache.concepts), Equals, 2)
}

func (s *MySuite) TestInitStepsCache(c *C) {
	f, _ := createFileIn(s.specsDir, "spec1.spec", spec1)
	f, _ = filepath.Abs(f)
	f1, _ := createFileIn(s.specsDir, "concept2.cpt", concept2)
	f1, _ = filepath.Abs(f1)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(3)

	specInfoGatherer.initConceptsCache()
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()
	c.Assert(len(specInfoGatherer.stepsCache.steps[f]), Equals, 2)
	c.Assert(len(specInfoGatherer.stepsCache.steps[f1]), Equals, 3)

}

func (s *MySuite) TestInitTagsCache(c *C) {
	_, err := createFileIn(s.specsDir, "specWithTags.spec", specWithTags)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(2)

	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initTagsCache()
	c.Assert(len(specInfoGatherer.Tags()), Equals, 5)
}

func (s *MySuite) TestInitTagsCacheWithMultipleFiles(c *C) {
	_, err := createFileIn(s.specsDir, "specWithTags.spec", specWithTags)
	if err != nil {
		c.Error(err)
	}

	_, err = createFileIn(s.specsDir, "spec2WithTags.spec", spec2WithTags)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(2)

	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initTagsCache()
	c.Assert(len(specInfoGatherer.Tags()), Equals, 6)
}

func (s *MySuite) TestGetStepsFromCachedSpecs(c *C) {
	f, _ := createFileIn(s.specsDir, "spec1.spec", spec1)
	f, _ = filepath.Abs(f)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(3)
	specInfoGatherer.initSpecsCache()

	stepsFromSpecsMap := specInfoGatherer.getStepsFromCachedSpecs()
	c.Assert(len(stepsFromSpecsMap[f]), Equals, 2)
	c.Assert(stepsFromSpecsMap[f][0].Value, Equals, "say hello")
	c.Assert(stepsFromSpecsMap[f][1].Value, Equals, "say {} to me")
}

func (s *MySuite) TestGetStepsFromCachedConcepts(c *C) {
	f, _ := createFileIn(s.specsDir, "concept1.cpt", concept1)
	f, _ = filepath.Abs(f)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(3)
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initConceptsCache()

	stepsFromConceptsMap := specInfoGatherer.getStepsFromCachedConcepts()
	c.Assert(len(stepsFromConceptsMap[f]), Equals, 3)
	c.Assert(stepsFromConceptsMap[f][0].Value, Equals, "first step with {}")
	c.Assert(stepsFromConceptsMap[f][1].Value, Equals, "say {} to me")
	c.Assert(stepsFromConceptsMap[f][2].Value, Equals, "a {} step")
}

func (s *MySuite) TestGetAvailableSteps(c *C) {
	var steps []*gauge.Step
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(2)
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()

	steps = specInfoGatherer.Steps(true)
	c.Assert(len(steps), Equals, 2)
	if !hasStep(steps, "say hello") {
		c.Fatalf("Step value not found %s", "say hello")
	}
	if !hasStep(steps, "say {} to me") {
		c.Fatalf("Step value not found %s", "say {} to me")
	}
}

func (s *MySuite) TestGetAvailableStepsShouldFilterDuplicates(c *C) {
	var steps []*gauge.Step
	_, err := createFileIn(s.specsDir, "spec2.spec", spec2)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(2)
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()

	steps = specInfoGatherer.Steps(true)
	c.Assert(len(steps), Equals, 2)
	if !hasStep(steps, "say hello") {
		c.Fatalf("Step value not found %s", "say hello")
	}
	if !hasStep(steps, "say {} to me") {
		c.Fatalf("Step value not found %s", "say {} to me")
	}
}

func (s *MySuite) TestGetAvailableStepsShouldFilterConcepts(c *C) {
	var steps []*gauge.Step
	_, err := createFileIn(s.specsDir, "concept1.cpt", concept4)
	if err != nil {
		c.Error(err)
	}

	_, err = createFileIn(s.specsDir, "spec1.spec", specWithConcept)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(3)
	specInfoGatherer.initConceptsCache()
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()

	steps = specInfoGatherer.Steps(true)
	c.Assert(len(steps), Equals, 1)
	if hasStep(steps, "foo bar with 1 step") {
		c.Fatalf("Step value found %s", "foo bar with 1 step")
	}
	steps = specInfoGatherer.Steps(false)
	c.Assert(len(steps), Equals, 2)
	if !hasStep(steps, "foo bar with 1 step") {
		c.Fatalf("Step value not found %s", "foo bar with 1 step")
	}
}

func (s *MySuite) TestGetAvailableAllStepsShouldFilterConcepts(c *C) {
	var steps []*gauge.Step
	_, err := createFileIn(s.specsDir, "concept1.cpt", concept4)
	if err != nil {
		c.Error(err)
	}

	_, err = createFileIn(s.specsDir, "spec1.spec", specWithConcept)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(3)
	specInfoGatherer.initConceptsCache()
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()

	steps = specInfoGatherer.AllSteps(true)
	c.Assert(len(steps), Equals, 2)
	if hasStep(steps, "foo bar with 1 step") {
		c.Fatalf("Step value found %s", "foo bar with 1 step")
	}
	steps = specInfoGatherer.AllSteps(false)
	c.Assert(len(steps), Equals, 3)
	if !hasStep(steps, "foo bar with 1 step") {
		c.Fatalf("Step value not found %s", "foo bar with 1 step")
	}
}

func hasStep(steps []*gauge.Step, stepText string) bool {
	for _, step := range steps {
		if step.Value == stepText {
			return true
		}
	}
	return false
}

func (s *MySuite) TestHasSpecForSpecDetail(c *C) {
	c.Assert((&SpecDetail{}).HasSpec(), Equals, false)
	c.Assert((&SpecDetail{Spec: &gauge.Specification{}}).HasSpec(), Equals, false)
	c.Assert((&SpecDetail{Spec: &gauge.Specification{Heading: &gauge.Heading{}}}).HasSpec(), Equals, true)
}

func (s *MySuite) TestGetAvailableSpecDetails(c *C) {
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	sig := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}, specsCache: specsCache{specDetails: make(map[string]*SpecDetail)}}
	specFiles := util.FindSpecFilesIn(s.specsDir)
	sig.specsCache.specDetails[specFiles[0]] = &SpecDetail{Spec: &gauge.Specification{Heading: &gauge.Heading{Value: "Specification Heading"}}}

	details := sig.GetAvailableSpecDetails(specFiles)

	c.Assert(len(details), Equals, 1)
	c.Assert(details[0].Spec.Heading.Value, Equals, "Specification Heading")
}

func (s *MySuite) TestGetAvailableSpecDetailsInDefaultDir(c *C) {
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	wd, _ := os.Getwd()
	err = os.Chdir(s.projectDir)
	if err != nil {
		c.Error(err)
	}
	defer func() {
		err := os.Chdir(wd)
		if err != nil {
			c.Error(err)
		}
	}()
	sig := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}, specsCache: specsCache{specDetails: make(map[string]*SpecDetail)}}
	specFiles := util.FindSpecFilesIn(specDir)
	sig.specsCache.specDetails[specFiles[0]] = &SpecDetail{Spec: &gauge.Specification{Heading: &gauge.Heading{Value: "Specification Heading"}}}

	details := sig.GetAvailableSpecDetails([]string{})

	c.Assert(len(details), Equals, 1)
	c.Assert(details[0].Spec.Heading.Value, Equals, "Specification Heading")
}

func (s *MySuite) TestGetAvailableSpecDetailsWithEmptyCache(c *C) {
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	sig := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}

	details := sig.GetAvailableSpecDetails([]string{})

	c.Assert(len(details), Equals, 0)
}

func (s *MySuite) TestParamsForStepFile(c *C) {
	file, _ := createFileIn(s.specsDir, "spec3.spec", spec3)
	file, _ = filepath.Abs(file)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(2)
	specInfoGatherer.initConceptsCache()
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()
	specInfoGatherer.initParamsCache()

	staticParams := specInfoGatherer.Params(file, gauge.Static)
	c.Assert(len(staticParams), Equals, 1)
	dynamicParams := specInfoGatherer.Params(file, gauge.Dynamic)
	c.Assert(len(dynamicParams), Equals, 3)
	hasParam := func(param string, list []gauge.StepArg) bool {
		for _, p := range list {
			if p.ArgValue() == param {
				return true
			}
		}
		return false
	}
	if !hasParam("hello", staticParams) {
		c.Errorf(`Param "hello" not found`)
	}
	if !hasParam("bye", dynamicParams) {
		c.Errorf(`Param "bye" not found`)
	}
	if !hasParam("Col1", dynamicParams) {
		c.Errorf(`Param "Col1" not found`)
	}
	if !hasParam("Col2", dynamicParams) {
		c.Errorf(`Param "Col1" not found`)
	}
}

func (s *MySuite) TestParamsForConceptFile(c *C) {
	file, _ := createFileIn(s.specsDir, "concept3.cpt", concept3)
	file, _ = filepath.Abs(file)
	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.waitGroup.Add(2)
	specInfoGatherer.initConceptsCache()
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()
	specInfoGatherer.initParamsCache()

	staticParams := specInfoGatherer.Params(file, gauge.Static)
	c.Assert(len(staticParams), Equals, 1)
	dynamicParams := specInfoGatherer.Params(file, gauge.Dynamic)
	c.Assert(len(dynamicParams), Equals, 2)
	hasParam := func(param string, list []gauge.StepArg) bool {
		for _, p := range list {
			if p.ArgValue() == param {
				return true
			}
		}
		return false
	}
	if !hasParam("foo", staticParams) {
		c.Errorf(`Param "foo" not found`)
	}
	if !hasParam("param", dynamicParams) {
		c.Errorf(`Param "param" not found`)
	}
	if !hasParam("final", dynamicParams) {
		c.Errorf(`Param "final" not found`)
	}
}

func (s *MySuite) TestAllStepsOnFileRename(c *C) {
	_, err := createFileIn(s.specsDir, "spec1.spec", spec1)
	if err != nil {
		c.Error(err)
	}

	specInfoGatherer := &SpecInfoGatherer{SpecDirs: []string{s.specsDir}}
	specInfoGatherer.initSpecsCache()
	specInfoGatherer.initStepsCache()

	c.Assert(len(specInfoGatherer.AllSteps(true)), Equals, 2)
	_, err = renameFileIn(s.specsDir, "spec1.spec", "spec42.spec")
	if err != nil {
		c.Error(err)
	}

	c.Assert(len(specInfoGatherer.AllSteps(true)), Equals, 2)
}

func createFileIn(dir string, fileName string, data []byte) (string, error) {
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return "", fmt.Errorf("unable to create %s: %s", dir, err.Error())
	}

	err = os.WriteFile(filepath.Join(dir, fileName), data, 0644)
	return filepath.Join(dir, fileName), err
}

func renameFileIn(dir string, oldFileName string, newFileName string) (string, error) {
	err := os.Rename(filepath.Join(dir, oldFileName), filepath.Join(dir, newFileName))
	return filepath.Join(dir, newFileName), err
}

func createDirIn(dir string, dirName string) (string, error) {
	tempDir, _ := os.MkdirTemp(dir, dirName)
	fullDirName := filepath.Join(dir, dirName)
	err := os.Rename(tempDir, fullDirName)
	return fullDirName, err
}
