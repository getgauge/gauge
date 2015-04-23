package main

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"path/filepath"
)

type MySuite struct {
	specsDir   string
	projectDir string
}

func (s *MySuite) SetUpTest(c *C) {
	s.projectDir, _ = ioutil.TempDir("", "gaugeTest")
	s.specsDir, _ = util.CreateDirIn(s.projectDir, "specs")
	config.ProjectRoot = s.projectDir
}

func (s *MySuite) TearDownTest(c *C) {
	os.RemoveAll(s.projectDir)
}

func (s *MySuite) TestGetAllStepsFromSpecs(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
`)

	specFile, err := util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromSpecs()

	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 1)
	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)

	steps, ok := specInfoGatherer.specStepMapCache[specFile]
	c.Assert(ok, Equals, true)
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].lineText, Equals, "say hello")
	c.Assert(steps[1].lineText, Equals, "say \"hello\" to me")
}

func (s *MySuite) TestRemoveSpec(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
`)
	data1 := []byte(`Specification Heading2
=====================
Scenario 1
----------
* say hello 1
* say "hello" to me 1
`)

	specFile1, err := util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	specFile2, err := util.CreateFileIn(s.specsDir, "Spec2.spec", data1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromSpecs()
	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 2)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 2)

	steps, ok := specInfoGatherer.specStepMapCache[specFile1]
	c.Assert(ok, Equals, true)
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].lineText, Equals, "say hello")
	c.Assert(steps[1].lineText, Equals, "say \"hello\" to me")

	steps, ok = specInfoGatherer.specStepMapCache[specFile2]
	c.Assert(ok, Equals, true)
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].lineText, Equals, "say hello 1")
	c.Assert(steps[1].lineText, Equals, "say \"hello\" to me 1")

	specInfoGatherer.removeSpec(filepath.Join(s.specsDir, "Spec1.spec"))

	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 2)

	steps, ok = specInfoGatherer.specStepMapCache[specFile2]
	c.Assert(ok, Equals, true)

	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 2)
}

func (s *MySuite) TestRemoveSpecWithCommonSteps(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
* a common step
`)
	data1 := []byte(`Specification Heading2
=====================
Scenario 1
----------
* say hello 1
* say "hello" to me 1
* a common step
`)

	util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	util.CreateFileIn(s.specsDir, "Spec2.spec", data1)

	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromSpecs()
	specInfoGatherer.updateAllStepsList()
	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 2)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 2)

	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 5)

	specInfoGatherer.removeSpec(filepath.Join(s.specsDir, "Spec1.spec"))

	allSteps = specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 3)

	stepValues := make([]string, 0)
	for _, stepValue := range allSteps {
		stepValues = append(stepValues, stepValue.stepValue)
	}

	c.Assert(stringInSlice("say hello 1", stepValues), Equals, true)
	c.Assert(stringInSlice("say {} to me 1", stepValues), Equals, true)
	c.Assert(stringInSlice("a common step", stepValues), Equals, true)
	c.Assert(stringInSlice("say hello", stepValues), Equals, false)
	c.Assert(stringInSlice("say {} to me", stepValues), Equals, false)

}

func (s *MySuite) TestAddSpec(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* first step with "foo"
* say "hello" to me
* a "final" step
`)
	data1 := []byte(`Specification Heading2
=====================
Scenario 1
----------
* say hello to gauge
* testing steps with "params"
* last step
`)

	util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromSpecs()
	specInfoGatherer.updateAllStepsList()

	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 1)
	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 3)

	util.CreateFileIn(s.specsDir, "Spec2.spec", data1)
	specInfoGatherer.addSpec(filepath.Join(s.specsDir, "Spec2.spec"))
	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 2)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 2)

	allSteps = specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 6)

	stepValues := make([]string, 0)
	for _, stepValue := range allSteps {
		stepValues = append(stepValues, stepValue.stepValue)
	}

	c.Assert(stringInSlice("first step with {}", stepValues), Equals, true)
	c.Assert(stringInSlice("say {} to me", stepValues), Equals, true)
	c.Assert(stringInSlice("a {} step", stepValues), Equals, true)
	c.Assert(stringInSlice("say hello to gauge", stepValues), Equals, true)
	c.Assert(stringInSlice("testing steps with {}", stepValues), Equals, true)
	c.Assert(stringInSlice("last step", stepValues), Equals, true)
}

func (s *MySuite) TestSameSpecAddedTwice(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* first step with "foo"
* say "hello" to me
* a "final" step
`)

	util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromSpecs()
	specInfoGatherer.updateAllStepsList()

	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 1)
	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 3)

	specInfoGatherer.addSpec(filepath.Join(s.specsDir, "Spec1.spec"))
	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 1)

	allSteps = specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 3)

	stepValues := make([]string, 0)
	for _, stepValue := range allSteps {
		stepValues = append(stepValues, stepValue.stepValue)
	}

	c.Assert(stringInSlice("first step with {}", stepValues), Equals, true)
	c.Assert(stringInSlice("say {} to me", stepValues), Equals, true)
	c.Assert(stringInSlice("a {} step", stepValues), Equals, true)
}

func (s *MySuite) TestAddingSpecWithParseFailures(c *C) {
	data := []byte(`NO heading parse failure
* first step with "foo"
* say "hello" to me
* a "final" step
`)

	util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromSpecs()
	specInfoGatherer.updateAllStepsList()

	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 0)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 0)
	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 0)
}

func (s *MySuite) TestFindingStepsAndConceptInfosFromConcepts(c *C) {
	data := []byte(`# foo bar
* first step with "foo"
* say "hello" to me
* a "final" step
`)

	util.CreateFileIn(s.specsDir, "concept.cpt", data)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.updateAllStepsList()

	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 0)
	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 3)

	conceptInfos := specInfoGatherer.getConceptInfos()
	c.Assert(len(conceptInfos), Equals, 1)
	c.Assert(conceptInfos[0].GetFilepath(), Equals, filepath.Join(s.specsDir, "concept.cpt"))
}

func (s *MySuite) TestAddingConcepts(c *C) {
	data := []byte(`# A concept
* first step with "foo"
* second say "hello" to me
* third "foo" step

# second concept
* fourth
* A concept
`)

	data1 := []byte(`# another cpt file
* second say "hello" to me
* fifth step

# concept with <a>
* sixth step with <a>
* seventh param step
`)

	util.CreateFileIn(s.specsDir, "concept.cpt", data)
	specInfoGatherer := new(specInfoGatherer)

	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.updateAllStepsList()

	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 1)
	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 4)

	c.Assert(len(specInfoGatherer.getConceptInfos()), Equals, 2)

	util.CreateFileIn(s.specsDir, "concept1.cpt", data1)

	specInfoGatherer.addConcept(filepath.Join(s.specsDir, "concept1.cpt"))
	allSteps = specInfoGatherer.getAvailableSteps()
	c.Assert(len(allSteps), Equals, 7)

	c.Assert(len(specInfoGatherer.getConceptInfos()), Equals, 4)

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
