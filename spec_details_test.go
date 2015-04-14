package main

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
	"io/ioutil"
)

func (s *MySuite) TestGetAllStepsFromSpecs(c *C) {
	dir, err := ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
`)
	specsDir, err := util.CreateDirIn(dir, "specs")
	c.Assert(err, Equals, nil)

	specFile, err := util.CreateFileIn(specsDir, "Spec1.spec", data)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(specInfoGatherer)
	config.ProjectRoot = dir

	stepsMap, _ := specInfoGatherer.getAllStepsFromSpecs()
	c.Assert(len(stepsMap), Equals, 1)
	steps, ok := stepsMap[specFile]
	c.Assert(ok, Equals, true)
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].lineText, Equals, "say hello")
	c.Assert(steps[1].lineText, Equals, "say \"hello\" to me")
}

func (s *MySuite) TestGetAllTagsFromSpecs(c *C) {
	dir, err := ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
	data := []byte(`Specification Heading
=====================
tags : hello world, first spec
Scenario 1
----------
tags: first scenario
* say hello
* say "hello" to me
`)
	specsDir, err := util.CreateDirIn(dir, "specs")
	c.Assert(err, Equals, nil)

	_, err = util.CreateFileIn(specsDir, "Spec1.spec", data)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(specInfoGatherer)
	config.ProjectRoot = dir

	tagsMap := specInfoGatherer.getAllTags()
	c.Assert(len(tagsMap), Equals, 3)
}
