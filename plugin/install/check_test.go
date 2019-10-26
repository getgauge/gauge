// Copyright 2015 ThoughtWorks, Inc.

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

package install

import (
	"github.com/getgauge/gauge/version"
	. "gopkg.in/check.v1"
)

var _ = Suite(&MySuite{})

func (s *MySuite) TestCheckGaugeUpdateWhenThereIsAnUpdate(c *C) {
	getLatestGaugeVersion = func(url string) (string, error) {
		return "0.1.0", nil
	}
	version.CurrentGaugeVersion = &version.Version{Major: 0, Minor: 0, Patch: 1}
	updateInfo := checkGaugeUpdate()[0]
	c.Assert(updateInfo.CompatibleVersion, Equals, "0.1.0")
	c.Assert(updateInfo.Name, Equals, "Gauge")
}

func (s *MySuite) TestCheckGaugeUpdateWhenThereIsNoUpdate(c *C) {
	getLatestGaugeVersion = func(url string) (string, error) {
		return "0.1.0", nil
	}
	version.CurrentGaugeVersion = &version.Version{Major: 0, Minor: 2, Patch: 0}
	updateInfos := checkGaugeUpdate()
	c.Assert(len(updateInfos), Equals, 0)
}

func (s *MySuite) TestCreatePluginUpdateDetailWhenThereIsAnUpdate(c *C) {
	version.CurrentGaugeVersion = &version.Version{Major: 0, Minor: 1, Patch: 1}
	ruby := "ruby"
	i := installDescription{Name: ruby, Versions: []versionInstallDescription{versionInstallDescription{Version: "0.1.1", GaugeVersionSupport: version.VersionSupport{Minimum: "0.1.0", Maximum: "0.1.2"}}}}
	updateDetails := createPluginUpdateDetail("0.1.0", i)
	c.Assert(len(updateDetails), Equals, 1)
	c.Assert(updateDetails[0].Name, Equals, ruby)
	c.Assert(updateDetails[0].CompatibleVersion, Equals, "0.1.1")
	c.Assert(updateDetails[0].Message, Equals, "Run 'gauge update ruby'")
}

func (s *MySuite) TestCreatePluginUpdateDetailWhenThereIsNoUpdate(c *C) {
	version.CurrentGaugeVersion = &version.Version{Major: 0, Minor: 1, Patch: 1}
	ruby := "ruby"
	i := installDescription{Name: ruby, Versions: []versionInstallDescription{versionInstallDescription{Version: "0.1.0", GaugeVersionSupport: version.VersionSupport{Minimum: "0.1.0", Maximum: "0.1.2"}}}}
	updateDetails := createPluginUpdateDetail("0.1.0", i)
	c.Assert(len(updateDetails), Equals, 0)
}

func (s *MySuite) TestGetGaugeVersionFromURL(c *C) {
	url := `https://github.com/getgauge/gauge/releases/tag/v1.0.4`

	v, err := getGaugeVersionFromURL(url)

	c.Assert(err, Equals, nil)
	c.Assert(v, Equals, "1.0.4")
}

func (s *MySuite) TestCreatePluginUpdateDetailForNightly(c *C) {
	version.CurrentGaugeVersion = &version.Version{Major: 0, Minor: 1, Patch: 1}
	ruby := "ruby"
	i := installDescription{Name: ruby, Versions: []versionInstallDescription{versionInstallDescription{Version: "0.1.1.nightly.2050-02-01", GaugeVersionSupport: version.VersionSupport{Minimum: "0.1.0", Maximum: "0.1.2"}}}}
	updateDetails := createPluginUpdateDetail("0.1.0", i)
	c.Assert(len(updateDetails), Equals, 0)
}
