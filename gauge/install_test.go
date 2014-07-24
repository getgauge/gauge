package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestConstructRunnerJsonInstallUrl(c *C) {
	constructedUrl := constructLanguageInstallJsonUrl("java", "")
	c.Assert(constructedUrl, Equals, "http://raw.github.com/getgauge/gauge-repository/master/runners/java/current/java-install.json")
}

func (s *MySuite) TestConstructRunnerJsonInstallUrlWithVersion(c *C) {
	constructedUrl := constructLanguageInstallJsonUrl("java", "0.0.1")
	c.Assert(constructedUrl, Equals, "http://raw.github.com/getgauge/gauge-repository/master/runners/java/0.0.1/java-install.json")
}
