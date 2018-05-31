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

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})
var dir string

func (s *MySuite) SetUpTest(c *C) {
	var err error
	dir, err = ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TearDownTest(c *C) {
	err := os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestFindAllSpecFiles(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* say hello
`)
	spec1, err := createFileIn(dir, "gaugeSpec1.spec", data)
	c.Assert(err, Equals, nil)

	dataRead, err := ioutil.ReadFile(spec1)
	c.Assert(err, Equals, nil)
	c.Assert(string(dataRead), Equals, string(data))
	c.Assert(len(FindSpecFilesIn(dir)), Equals, 1)

	_, err = createFileIn(dir, "gaugeSpec2.spec", data)
	c.Assert(err, Equals, nil)

	c.Assert(len(FindSpecFilesIn(dir)), Equals, 2)
}

func (s *MySuite) TestFindAllConceptFiles(c *C) {
	data := []byte(`#Concept Heading`)
	_, err := createFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	_, err = createFileIn(dir, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)
}

func (s *MySuite) TestIsValidSpecExension(c *C) {
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.spec"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.sPeC"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.SPEC"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.md"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.MD"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myconcept.cpt"), Equals, false)
}

func (s *MySuite) TestIsValidConcpetExension(c *C) {
	c.Assert(IsValidConceptExtension("/home/user/foo/myconcept.cpt"), Equals, true)
	c.Assert(IsValidConceptExtension("/home/user/foo/myconcept.CPT"), Equals, true)
	c.Assert(IsValidConceptExtension("/home/user/foo/myconcept.cPt"), Equals, true)
	c.Assert(IsValidConceptExtension("/home/user/foo/myspec.spC"), Equals, false)
}

func (s *MySuite) TestFindAllConceptFilesShouldFilterDirectoriesThatAreSkipped(c *C) {
	config.ProjectRoot = dir
	data := []byte(`#Concept Heading`)
	git, _ := createDirIn(dir, ".git")
	bin, _ := createDirIn(dir, "gauge_bin")
	reports, _ := createDirIn(dir, "reports")
	env, _ := createDirIn(dir, "env")

	createFileIn(git, "concept1.cpt", data)
	createFileIn(bin, "concept2.cpt", data)
	createFileIn(reports, "concept3.cpt", data)
	createFileIn(env, "concept4.cpt", data)

	c.Assert(len(FindConceptFilesIn(dir)), Equals, 0)

	_, err := createFileIn(dir, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)
}

func (s *MySuite) TestFindAllConceptFilesInNestedDir(c *C) {
	data := []byte(`#Concept Heading
* Say "hello" to gauge
`)
	_, err := createFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	dir1, err := ioutil.TempDir(dir, "gaugeTest1")
	c.Assert(err, Equals, nil)

	_, err = createFileIn(dir1, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)
}

func (s *MySuite) TestGetConceptFiles(c *C) {
	config.ProjectRoot = "_testdata"
	specsDir, _ := filepath.Abs(filepath.Join("_testdata", "specs"))

	config.ProjectRoot = specsDir
	c.Assert(len(GetConceptFiles()), Equals, 2)

	config.ProjectRoot = filepath.Join(specsDir, "concept1.cpt")
	c.Assert(len(GetConceptFiles()), Equals, 1)

	config.ProjectRoot = filepath.Join(specsDir, "subdir")
	c.Assert(len(GetConceptFiles()), Equals, 1)

	config.ProjectRoot = filepath.Join(specsDir, "subdir", "concept2.cpt")
	c.Assert(len(GetConceptFiles()), Equals, 1)
}

func (s *MySuite) TestFindAllNestedDirs(c *C) {
	nested1 := filepath.Join(dir, "nested")
	nested2 := filepath.Join(dir, "nested2")
	nested3 := filepath.Join(dir, "nested2", "deep")
	nested4 := filepath.Join(dir, "nested2", "deep", "deeper")
	os.Mkdir(nested1, 0755)
	os.Mkdir(nested2, 0755)
	os.Mkdir(nested3, 0755)
	os.Mkdir(nested4, 0755)

	nestedDirs := FindAllNestedDirs(dir)
	c.Assert(len(nestedDirs), Equals, 4)
	c.Assert(stringInSlice(nested1, nestedDirs), Equals, true)
	c.Assert(stringInSlice(nested2, nestedDirs), Equals, true)
	c.Assert(stringInSlice(nested3, nestedDirs), Equals, true)
	c.Assert(stringInSlice(nested4, nestedDirs), Equals, true)
}

func (s *MySuite) TestFindAllNestedDirsWhenDirDoesNotExist(c *C) {
	nestedDirs := FindAllNestedDirs("unknown-dir")
	c.Assert(len(nestedDirs), Equals, 0)
}

func (s *MySuite) TestIsDir(c *C) {
	c.Assert(IsDir(dir), Equals, true)
	c.Assert(IsDir(filepath.Join(dir, "foo.txt")), Equals, false)
	c.Assert(IsDir("unknown path"), Equals, false)
	c.Assert(IsDir("foo/goo.txt"), Equals, false)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (s *MySuite) TestGetPathToFile(c *C) {
	var path string
	config.ProjectRoot = "PROJECT_ROOT"
	absPath, _ := filepath.Abs("resources")
	path = GetPathToFile(absPath)
	c.Assert(path, Equals, absPath)

	path = GetPathToFile("resources")
	c.Assert(path, Equals, filepath.Join(config.ProjectRoot, "resources"))
}

func createFileIn(dir string, fileName string, data []byte) (string, error) {
	os.MkdirAll(dir, 0755)
	err := ioutil.WriteFile(filepath.Join(dir, fileName), data, 0644)
	return filepath.Join(dir, fileName), err
}

func createDirIn(dir string, dirName string) (string, error) {
	tempDir, err := ioutil.TempDir(dir, dirName)
	fullDirName := filepath.Join(dir, dirName)
	err = os.Rename(tempDir, fullDirName)
	return fullDirName, err
}
