// Copyright 2014 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package main

import . "gopkg.in/check.v1"

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

func (s *MySuite) TestSortingVersionInstallDescriptionsInDecreasingVersionOrder(c *C) {
	installDescription := createInstallDescriptionWithVersions("5.8.8", "1.7.8", "4.8.9", "0.7.6", "3.5.6")
	installDescription.sortVersionInstallDescriptions()
	c.Assert(installDescription.Versions[0].Version, Equals, "5.8.8")
	c.Assert(installDescription.Versions[1].Version, Equals, "4.8.9")
	c.Assert(installDescription.Versions[2].Version, Equals, "3.5.6")
	c.Assert(installDescription.Versions[3].Version, Equals, "1.7.8")
	c.Assert(installDescription.Versions[4].Version, Equals, "0.7.6")
}

func (s *MySuite) TestFindingLatestCompatibleVersionSuccess(c *C) {
	installDescription := createInstallDescriptionWithVersions("5.8.8", "1.7.8", "4.8.9", "0.7.6")
	addVersionSupportToInstallDescription(installDescription,
		&versionSupport{"0.0.2", "0.8.7"},
		&versionSupport{"1.2.4", "1.2.6"},
		&versionSupport{"0.9.8", "1.2.1"},
		&versionSupport{Minimum: "0.7.7"})
	versionInstallDesc, err := installDescription.getLatestCompatibleVersionTo(&version{1, 0, 0})
	c.Assert(err, Equals, nil)
	c.Assert(versionInstallDesc.Version, Equals, "4.8.9")
}

func (s *MySuite) TestFindingLatestCompatibleVersionFailing(c *C) {
	installDescription := createInstallDescriptionWithVersions("2.8.8", "0.7.8", "4.8.9", "1.7.6")
	addVersionSupportToInstallDescription(installDescription,
		&versionSupport{"0.0.2", "0.8.7"},
		&versionSupport{"1.2.4", "1.2.6"},
		&versionSupport{"0.9.8", "1.0.0"},
		&versionSupport{Minimum: "1.7.7"})
	_, err := installDescription.getLatestCompatibleVersionTo(&version{1, 1, 0})
	c.Assert(err, NotNil)
}

func createInstallDescriptionWithVersions(versionNumbers ...string) *installDescription {
	versionInstallDescriptions := make([]versionInstallDescription, 0)
	for _, version := range versionNumbers {
		versionInstallDescriptions = append(versionInstallDescriptions, versionInstallDescription{Version: version})
	}
	return &installDescription{Name: "my-plugin", Versions: versionInstallDescriptions}
}

func addVersionSupportToInstallDescription(installDescription *installDescription, versionSupportList ...*versionSupport) {
	for i, _ := range installDescription.Versions {
		installDescription.Versions[i].GaugeVersionSupport = *versionSupportList[i]
	}
}
