package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestConstructPluginJsonInstallUrl(c *C) {
	constructedUrl := constructPluginInstallJsonUrl("java")
	c.Assert(constructedUrl, Equals, "http://raw.github.com/getgauge/gauge-repository/master/java-install.json")
}

func (s *MySuite) TestFindVersion(c *C) {
	installDescription := createInstallDescriptionWithVersions("0.0.4", "0.6.7", "0.7.4", "3.6.5")
	versionInstall, err := installDescription.getVersion("0.7.4")
	c.Assert(err, Equals, nil)
	c.Assert(versionInstall.Version, Equals, "0.7.4")
}

func (s *MySuite) TestFindVersionFailing(c *C) {
	installDescription := createInstallDescriptionWithVersions("0.0.4", "0.6.7", "0.7.4", "3.6.5")
	_, err := installDescription.getVersion("0.9.4")
	c.Assert(err, NotNil)
}

func (s *MySuite) TestCheckVersionCompatibilitySuccess(c *C) {
	versionSupported := &versionSupport{"0.6.5", "1.8.5"}
	gaugeVersion := &version{0, 6, 7}
	c.Assert(checkCompatiblity(gaugeVersion, versionSupported), Equals, nil)

	versionSupported = &versionSupport{"0.0.1", "0.0.1"}
	gaugeVersion = &version{0, 0, 1}
	c.Assert(checkCompatiblity(gaugeVersion, versionSupported), Equals, nil)

	versionSupported = &versionSupport{Minimum: "0.0.1"}
	gaugeVersion = &version{1, 5, 2}
	c.Assert(checkCompatiblity(gaugeVersion, versionSupported), Equals, nil)

	versionSupported = &versionSupport{Minimum: "0.5.1"}
	gaugeVersion = &version{0, 5, 1}
	c.Assert(checkCompatiblity(gaugeVersion, versionSupported), Equals, nil)

}

func (s *MySuite) TestCheckVersionCompatibilityFailure(c *C) {
	versionsSupported := &versionSupport{"0.6.5", "1.8.5"}
	gaugeVersion := &version{1, 9, 9}
	c.Assert(checkCompatiblity(gaugeVersion, versionsSupported), NotNil)

	versionsSupported = &versionSupport{"0.0.1", "0.0.1"}
	gaugeVersion = &version{0, 0, 2}
	c.Assert(checkCompatiblity(gaugeVersion, versionsSupported), NotNil)

	versionsSupported = &versionSupport{Minimum: "1.3.1"}
	gaugeVersion = &version{1, 3, 0}
	c.Assert(checkCompatiblity(gaugeVersion, versionsSupported), NotNil)

	versionsSupported = &versionSupport{Minimum: "0.5.1"}
	gaugeVersion = &version{0, 0, 9}
	c.Assert(checkCompatiblity(gaugeVersion, versionsSupported), NotNil)

}

func createInstallDescriptionWithVersions(versionNumbers ...string) *installDescription {
	versionInstallDescriptions := make([]versionInstallDescription, 0)
	for _, version := range versionNumbers {
		versionInstallDescriptions = append(versionInstallDescriptions, versionInstallDescription{Version: version})
	}
	return &installDescription{Name: "my-plugin", Versions: versionInstallDescriptions}
}
