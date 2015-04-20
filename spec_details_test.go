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
	c.Assert(2, Equals, len(allSteps))
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
	c.Assert(len(specInfoGatherer.specStepMapCache), Equals, 2)
	c.Assert(len(specInfoGatherer.availableSpecs), Equals, 2)

	allSteps := specInfoGatherer.getAvailableSteps()
	c.Assert(5, Equals, len(allSteps))

	specInfoGatherer.removeSpec(filepath.Join(s.specsDir, "Spec1.spec"))

	allSteps = specInfoGatherer.getAvailableSteps()
	c.Assert(3, Equals, len(allSteps))

	stepValues := make([]string, 0)
	for _, stepValue := range allSteps {
		stepValues = append(stepValues, stepValue.stepValue)
	}

	c.Assert(true, Equals, stringInSlice("say hello 1", stepValues))
	c.Assert(true, Equals, stringInSlice("say {} to me 1", stepValues))
	c.Assert(true, Equals, stringInSlice("a common step", stepValues))
	c.Assert(false, Equals, stringInSlice("say hello", stepValues))
	c.Assert(false, Equals, stringInSlice("say {} to me", stepValues))

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
