package main

import . "launchpad.net/gocheck"

func (s *MySuite) TestConstructRunnerJsonInstallUrl(c *C) {
	constructedUrl := constructLanguageInstallJsonUrl("java", "")
	c.Assert(constructedUrl, Equals, "http://raw.github.com/getgauge/gauge-repository/master/runners/java/current/java-install.json")
}

func (s *MySuite) TestConstructRunnerJsonInstallUrlWithVersion(c *C) {
	constructedUrl := constructLanguageInstallJsonUrl("java", "0.0.1")
	c.Assert(constructedUrl, Equals, "http://raw.github.com/getgauge/gauge-repository/master/runners/java/0.0.1/java-install.json")
}

func (s *MySuite) TestVersionCompatibilitySuccess(c *C) {
	installDescription := createInstallDescriptionWithMinimumMaximumSupportVersions("0.6.5", "1.8.5")
	gaugeVersion := &version{0, 6, 7}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), Equals, nil)

	installDescription = createInstallDescriptionWithMinimumMaximumSupportVersions("0.0.1", "0.0.1")
	gaugeVersion = &version{0, 0, 1}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), Equals, nil)

	installDescription = createInstallDescriptionWithMinimumMaximumSupportVersions("0.0.1")
	gaugeVersion = &version{1, 5, 2}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), Equals, nil)

	installDescription = createInstallDescriptionWithMinimumMaximumSupportVersions("0.5.1")
	gaugeVersion = &version{0, 5, 1}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), Equals, nil)

}

func (s *MySuite) TestVersionCompatibilityFailure(c *C) {
	installDescription := createInstallDescriptionWithMinimumMaximumSupportVersions("0.6.5", "1.8.5")
	gaugeVersion := &version{1, 9, 9}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), NotNil)

	installDescription = createInstallDescriptionWithMinimumMaximumSupportVersions("0.0.1", "0.0.1")
	gaugeVersion = &version{0, 0, 2}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), NotNil)

	installDescription = createInstallDescriptionWithMinimumMaximumSupportVersions("1.3.1")
	gaugeVersion = &version{1, 3, 0}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), NotNil)

	installDescription = createInstallDescriptionWithMinimumMaximumSupportVersions("0.5.1")
	gaugeVersion = &version{0, 0, 9}
	c.Assert(checkVersionCompatibilityWithGauge(installDescription, gaugeVersion), NotNil)

}

func createInstallDescriptionWithMinimumMaximumSupportVersions(supportVersions ...string) *installDescription {
	if len(supportVersions) == 2 {
		return &installDescription{Name: "Test", Version: "1.2.3", GaugeVersionSupport: versionSupport{supportVersions[0], supportVersions[1]}}
	} else if len(supportVersions) == 1 {
		return &installDescription{Name: "Test", Version: "1.2.3", GaugeVersionSupport: versionSupport{Minimum: supportVersions[0]}}
	}
	return nil
}
