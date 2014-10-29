package main

import (
	. "launchpad.net/gocheck"
	"regexp"
	"strconv"
)

var acceptedExtensions1 = make(map[string]bool)

func init() {
	acceptedExtensions1[".spec"] = true
	acceptedExtensions1[".md"] = true
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
	specName, scenarioNum := GetIndexedSpecName("specs/hello_world.spec:67")
	c.Assert(specName, Equals, "specs/hello_world.spec")
	c.Assert(scenarioNum, Equals, 67)
}
func (s *MySuite) TestToObtainIndexedSpecName1(c *C) {
	specName, scenarioNum := GetIndexedSpecName("hello_world.spec:67342")
	c.Assert(specName, Equals, "hello_world.spec")
	c.Assert(scenarioNum, Equals, 67342)
}

func isIndexedSpec(specSource string) bool {
	return getIndex(specSource) != nil
}

func GetIndexedSpecName(IndexedSpec string) (string, int) {
	var specName, scenarioNum string
	index := getIndex(IndexedSpec)
	for i := 0; i < index[0]; i++ {
		specName += string(IndexedSpec[i])
	}
	typeOfSpec := getTypeOfSpecFile(IndexedSpec)
	for i := index[0] + len(typeOfSpec) + 1; i < index[1]; i++ {
		scenarioNum += string(IndexedSpec[i])
	}
	scenarioNumber, _ := strconv.Atoi(scenarioNum)
	return specName + typeOfSpec, scenarioNumber
}

func getIndex(specSource string) []int {
	re, _ := regexp.Compile(getTypeOfSpecFile(specSource) + ":[0-9]+$")
	index := re.FindStringSubmatchIndex(specSource)
	if index != nil {
		return index
	}
	return nil
}

func getTypeOfSpecFile(specSource string) string {
	for ext, accepted := range acceptedExtensions1 {
		if accepted {
			re, _ := regexp.Compile(ext)
			if re.FindStringSubmatchIndex(specSource) != nil {
				return ext
			}
		}
	}
	return ""
}
