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
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/version"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

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
		&version.VersionSupport{"0.0.2", "0.8.7"},
		&version.VersionSupport{"1.2.4", "1.2.6"},
		&version.VersionSupport{"0.9.8", "1.2.1"},
		&version.VersionSupport{Minimum: "0.7.7"})
	versionInstallDesc, err := installDescription.getLatestCompatibleVersionTo(&version.Version{1, 0, 0})
	c.Assert(err, Equals, nil)
	c.Assert(versionInstallDesc.Version, Equals, "4.8.9")
}

func (s *MySuite) TestFindingLatestCompatibleVersionFailing(c *C) {
	installDescription := createInstallDescriptionWithVersions("2.8.8", "0.7.8", "4.8.9", "1.7.6")
	addVersionSupportToInstallDescription(installDescription,
		&version.VersionSupport{"0.0.2", "0.8.7"},
		&version.VersionSupport{"1.2.4", "1.2.6"},
		&version.VersionSupport{"0.9.8", "1.0.0"},
		&version.VersionSupport{Minimum: "1.7.7"})
	_, err := installDescription.getLatestCompatibleVersionTo(&version.Version{1, 1, 0})
	c.Assert(err, NotNil)
}

func createInstallDescriptionWithVersions(versionNumbers ...string) *installDescription {
	var versionInstallDescriptions []versionInstallDescription
	for _, version := range versionNumbers {
		versionInstallDescriptions = append(versionInstallDescriptions, versionInstallDescription{Version: version})
	}
	return &installDescription{Name: "my-plugin", Versions: versionInstallDescriptions}
}

func addVersionSupportToInstallDescription(installDescription *installDescription, versionSupportList ...*version.VersionSupport) {
	for i := range installDescription.Versions {
		installDescription.Versions[i].GaugeVersionSupport = *versionSupportList[i]
	}
}

//func (s *MySuite) TestInstallGaugePluginFromNonExistingZipFile(c *C) {
//	result := InstallPluginFromZipFile(filepath.Join("test_resources", "notPresent.zip"), "ruby")
//	c.Assert(result.Error.Error(), Equals, fmt.Sprintf("ZipFile %s does not exist", filepath.Join("test_resources", "notPresent.zip")))
//}
//
func (s *MySuite) TestReturnsErrorInsteadOfPanicDuringZipFileInstall(c *C) {

	sourceFileName := filepath.Join("_testdata", "plugin.json")
	targetPathSingleDot := filepath.Join("_testdata", "zip_with_single_dot.zip")
	targetPathMultipleDots := filepath.Join("_testdata", "zip_with_multiple-dot.s.zip")
	err := createZipFromFile(sourceFileName, targetPathSingleDot)
	c.Assert(err, IsNil)
	defer os.Remove(targetPathSingleDot)

	err = createZipFromFile(sourceFileName, targetPathMultipleDots)
	c.Assert(err, IsNil)
	defer os.Remove(targetPathMultipleDots)

	result := InstallPluginFromZipFile(targetPathSingleDot, "ruby")
	c.Assert(result.Error.Error(), Equals, fmt.Sprintf("provided zip file is not a valid plugin of ruby"))

	result = InstallPluginFromZipFile(targetPathMultipleDots, "ruby")
	c.Assert(result.Error.Error(), Equals, fmt.Sprintf("provided plugin is not compatible with OS linux amd64"))
}

func (s *MySuite) TestGetVersionedPluginDirName(c *C) {
	name := getVersionedPluginDirName("abcd/foo/bar/html-report-2.0.1.nightly-2016-02-09-darwin.x86.zip")
	c.Assert(name, Equals, "2.0.1.nightly-2016-02-09")

	name = getVersionedPluginDirName("abcd/foo/bar/xml-report-2.0.1.nightly-2016-02-09-darwin.x86.zip")
	c.Assert(name, Equals, "2.0.1.nightly-2016-02-09")

	name = getVersionedPluginDirName("abcd/foo/bar/html-report-0.3.4-windows.x86_64.zip")
	c.Assert(name, Equals, "0.3.4")

	name = getVersionedPluginDirName("abcd/foo/bar/gauge-java-0.3.4.nightly-2016-02-09-linux.x86.zip")
	c.Assert(name, Equals, "0.3.4.nightly-2016-02-09")

	name = getVersionedPluginDirName("abcd/foo/bar/gauge-java-0.3.4-linux.x86_64.zip")
	c.Assert(name, Equals, "0.3.4")

	name = getVersionedPluginDirName("abcd/foo/gauge-ruby-0.1.2.nightly-2016-02-09-linux.x86.zip")
	c.Assert(name, Equals, "0.1.2.nightly-2016-02-09")

	if util.IsWindows() {
		name = getVersionedPluginDirName("C:\\Users\\apoorvam\\AppData\\Local\\Temp\\gauge_temp1456130044460213700\\gauge-java-0.3.4-windows.x86_64.zip")
		c.Assert(name, Equals, "0.3.4")
	}
}

func (s *MySuite) TestGetGaugePluginForJava(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "java"))
	p, err := parsePluginJSON(path, "java")
	c.Assert(err, Equals, nil)
	c.Assert(p.ID, Equals, "java")
	c.Assert(p.Version, Equals, "0.3.4")
	c.Assert(p.Description, Equals, "Java support for gauge")
	c.Assert(p.PreInstall.Darwin[0], Equals, "pre install command")
	c.Assert(p.PreUnInstall.Darwin[0], Equals, "pre uninstall command")
	c.Assert(p.GaugeVersionSupport.Minimum, Equals, "0.3.0")
	c.Assert(p.GaugeVersionSupport.Maximum, Equals, "")
}

func (s *MySuite) TestGetGaugePluginForReportPlugin(c *C) {
	path, _ := filepath.Abs("_testdata")
	p, err := parsePluginJSON(path, "html-report")
	c.Assert(err, Equals, nil)
	c.Assert(p.ID, Equals, "html-report")
	c.Assert(p.Version, Equals, "2.0.1")
	c.Assert(p.Description, Equals, "Html reporting plugin")
	c.Assert(p.PreInstall.Darwin[0], Equals, "pre install command")
	c.Assert(p.PreUnInstall.Darwin[0], Equals, "pre uninstall command")
	c.Assert(p.GaugeVersionSupport.Minimum, Equals, "0.3.0")
	c.Assert(p.GaugeVersionSupport.Maximum, Equals, "")
}

func (s *MySuite) TestMatchesUninstallVersionIfUninstallPluginVersionIsNotGiven(c *C) {
	dirPath := "somepath"
	uninstallVersion := ""

	c.Assert(matchesUninstallVersion(dirPath, uninstallVersion), Equals, true)
}

func (s *MySuite) TestMatchesUninstallVersionIfUninstallPluginVersionMatches(c *C) {
	dirPath := "0.1.1-nightly-2016-05-05"
	uninstallVersion := "0.1.1-nightly-2016-05-05"

	c.Assert(matchesUninstallVersion(dirPath, uninstallVersion), Equals, true)
}

func (s *MySuite) TestMatchesUninstallVersionIfUninstallPluginVersionDoesntMatches(c *C) {
	dirPath := "0.1.1"
	uninstallVersion := "0.1.1-nightly-2016-05-05"

	c.Assert(matchesUninstallVersion(dirPath, uninstallVersion), Equals, false)
}

func (s *MySuite) TestIsPlatformIndependentZipFile(c *C) {
	javaReleased := "java-3.1.0.nightly-2017-02-08-darwin.x86_64.zip"
	csharpReleased := "gauge-csharp-0.10.1.zip"
	javaNightly := "gauge-java-3.1.0.nightly-2017-02-08-darwin.x86_64.zip"
	csharpNightly := "gauge-csharp-0.10.1.nightly-2017-02-17.zip"

	c.Assert(isPlatformIndependent(javaReleased), Equals, false)
	c.Assert(isPlatformIndependent(csharpReleased), Equals, true)
	c.Assert(isPlatformIndependent(javaNightly), Equals, false)
	c.Assert(isPlatformIndependent(csharpNightly), Equals, true)
}

func createZipFromFile(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("file expected, directory found")
	}

	file, err := os.OpenFile(source, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	header, err := zip.FileInfoHeader(info)
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
