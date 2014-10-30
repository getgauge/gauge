package main

import (
	. "launchpad.net/gocheck"
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
