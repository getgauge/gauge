package main

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"os"
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

	stepsMap, _ := specInfoGatherer.getAllStepsFromSpecs()
	c.Assert(len(stepsMap), Equals, 1)
	steps, ok := stepsMap[specFile]
	c.Assert(ok, Equals, true)
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].lineText, Equals, "say hello")
	c.Assert(steps[1].lineText, Equals, "say \"hello\" to me")
}

func (s *MySuite) TestGetAllTagsFromSpecs(c *C) {
	data := []byte(`Specification Heading
=====================
tags : hello world, first spec
Scenario 1
----------
tags: first scenario
* say hello
* say "hello" to me
`)
	_, err := util.CreateFileIn(s.specsDir, "Spec1.spec", data)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(specInfoGatherer)

	tagsMap := specInfoGatherer.getAllTags()
	c.Assert(len(tagsMap), Equals, 3)
}
