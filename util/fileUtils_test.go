/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})
var dir string

func (s *MySuite) SetUpTest(c *C) {
	var err error
	dir, err = os.MkdirTemp("", "gaugeTest")
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TearDownTest(c *C) {
	err := os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
	c.Assert(os.Unsetenv(env.SpecsDir), Equals, nil)
	c.Assert(os.Unsetenv(env.ConceptsDir), Equals, nil)
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

	dataRead, err := os.ReadFile(spec1)
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

func (s *MySuite) TestIsValidSpecExensionDefault(c *C) {
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.spec"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.sPeC"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.SPEC"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.md"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myspec.MD"), Equals, true)
	c.Assert(IsValidSpecExtension("/home/user/foo/myconcept.cpt"), Equals, false)
}

func TestIsValidSpecExensionWhenSet(t *testing.T) {
	var tests = map[string]bool{
		"/home/user/foo/myspec.spec":   true,
		"/home/user/foo/myspec.sPeC":   true,
		"/home/user/foo/myspec.SPEC":   true,
		"/home/user/foo/myspec.md":     false,
		"/home/user/foo/myspec.MD":     false,
		"/home/user/foo/myspec.foo":    true,
		"/home/user/foo/myspec.Foo":    true,
		"/home/user/foo/myconcept.cpt": false,
	}
	for k, v := range tests {
		t.Run(filepath.Ext(k), func(t *testing.T) {
			old := env.GaugeSpecFileExtensions
			env.GaugeSpecFileExtensions = func() []string { return []string{".spec", ".foo"} }
			if IsValidSpecExtension(k) != v {
				t.Errorf("Expected IsValidSpecExtension(%s) to be %t", k, v)
			}
			env.GaugeSpecFileExtensions = old
		})
	}
}

func (s *MySuite) TestIsValidConceptExension(c *C) {
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

	_, err := createFileIn(git, "concept1.cpt", data)
	c.Assert(err, IsNil)
	_, err = createFileIn(bin, "concept2.cpt", data)
	c.Assert(err, IsNil)
	_, err = createFileIn(reports, "concept3.cpt", data)
	c.Assert(err, IsNil)
	_, err = createFileIn(env, "concept4.cpt", data)
	c.Assert(err, IsNil)

	c.Assert(len(FindConceptFilesIn(dir)), Equals, 0)

	_, err = createFileIn(dir, "concept2.cpt", data)
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

	dir1, err := os.MkdirTemp(dir, "gaugeTest1")
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

func (s *MySuite) TestGetConceptFilesWhenSpecDirIsOutsideProjectRoot(c *C) {
	config.ProjectRoot = "_testdata"
	_ = os.Setenv(env.SpecsDir, "../_testSpecDir")
	c.Assert(len(GetConceptFiles()), Equals, 3)
}

func (s *MySuite) TestGetConceptFilesWhenSpecDirIsWithInProject(c *C) {
	config.ProjectRoot = "_testdata"
	_ = os.Setenv(env.SpecsDir, "_testdata/specs")
	c.Assert(len(GetConceptFiles()), Equals, 2)
}

func (s *MySuite) TestGetConceptFilesWhenConceptsDirSet(c *C) {
	config.ProjectRoot = "_testdata"
	c.Assert(len(GetConceptFiles()), Equals, 2)

	_ = os.Setenv(env.ConceptsDir, filepath.Join("specs", "concept1.cpt"))
	c.Assert(len(GetConceptFiles()), Equals, 1)

	_ = os.Setenv(env.ConceptsDir, filepath.Join("specs", "subdir"))
	c.Assert(len(GetConceptFiles()), Equals, 1)

	conceptPath, _ := filepath.Abs(filepath.Join("_testdata", "specs", "subdir", "concept2.cpt"))
	_ = os.Setenv(env.ConceptsDir, conceptPath)
	c.Assert(len(GetConceptFiles()), Equals, 1)
}

func (s *MySuite) TestGetConceptFilesWhenConceptDirIsOutsideProjectRoot(c *C) {
	config.ProjectRoot = "_testdata"
	_ = os.Setenv(env.ConceptsDir, "../_testSpecDir,../_testdata/specs")
	c.Assert(len(GetConceptFiles()), Equals, 3)
}

func (s *MySuite) TestFindAllNestedDirs(c *C) {
	nested1 := filepath.Join(dir, "nested")
	nested2 := filepath.Join(dir, "nested2")
	nested3 := filepath.Join(dir, "nested2", "deep")
	nested4 := filepath.Join(dir, "nested2", "deep", "deeper")
	err := os.Mkdir(nested1, 0755)
	c.Assert(err, IsNil)
	err = os.Mkdir(nested2, 0755)
	c.Assert(err, IsNil)
	err = os.Mkdir(nested3, 0755)
	c.Assert(err, IsNil)
	err = os.Mkdir(nested4, 0755)
	c.Assert(err, IsNil)

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

func (s *MySuite) TestGetPathToFileInGaugeDataDir(c *C) {
	var path string
	config.ProjectRoot = "PROJECT_ROOT"
	oldGaugeDataDirFn := env.GaugeDataDir
	defer func(fn func() string) { env.GaugeDataDir = fn }(oldGaugeDataDirFn)
	env.GaugeDataDir = func() string { return "foo" }

	absPath, _ := filepath.Abs("foo.csv")
	path = GetPathToFile(absPath)
	c.Assert(path, Equals, absPath)
}

func (s *MySuite) TestGetPathToAbsFileWithGaugeDataDir(c *C) {
	var path string
	config.ProjectRoot = "PROJECT_ROOT"
	oldGaugeDataDirFn := env.GaugeDataDir
	defer func(fn func() string) { env.GaugeDataDir = fn }(oldGaugeDataDirFn)
	env.GaugeDataDir = func() string { return "foo" }

	path = GetPathToFile("foo.csv")
	c.Assert(path, Equals, filepath.Join(config.ProjectRoot, "foo", "foo.csv"))
}

func (s *MySuite) TestGetSpecFilesWhenSpecsDirDoesNotExists(c *C) {
	var expectedErrorMessage string
	exitWithMessage = func(message string) {
		expectedErrorMessage = message
	}
	GetSpecFiles([]string{"dir1"})
	c.Assert(expectedErrorMessage, Equals, "Specs directory dir1 does not exist.")
}

func (s *MySuite) TestGetConceptFilesWhenConceptsDirDoesNotExists(c *C) {
	var expectedErrorMessage string
	exitWithMessage = func(message string) {
		expectedErrorMessage = message
	}
	config.ProjectRoot = "_testdata"

	_ = os.Setenv(env.SpecsDir, "specs2")
	GetConceptFiles()
	directory, _ := filepath.Abs(filepath.Join(config.ProjectRoot, "specs2"))
	c.Assert(expectedErrorMessage, Equals, fmt.Sprintf("No such file or directory: %s", directory))

	_ = os.Setenv(env.SpecsDir, "_testSpecsDir,non-exisitng")
	GetConceptFiles()
	directory, _ = filepath.Abs(filepath.Join(config.ProjectRoot, "non-exisitng"))
	c.Assert(expectedErrorMessage, Equals, fmt.Sprintf("No such file or directory: %s", directory))
}

func (s *MySuite) TestGetSpecFilesWhenSpecsDirIsEmpty(c *C) {
	var expectedErrorMessage string
	exitWithMessage = func(message string) {
		expectedErrorMessage = message
	}
	GetSpecFiles([]string{dir})
	c.Assert(expectedErrorMessage, Equals, fmt.Sprintf("No specifications found in %s.", dir))
}

func (s *MySuite) TestGetSpecFiles(c *C) {
	expectedSpecFiles := []string{"spec-file-1.spec", "spec-file2.spec"}
	old := FindSpecFilesIn
	FindSpecFilesIn = func(dir string) []string {
		return []string{"spec-file-1.spec", "spec-file2.spec"}
	}
	actualSpecFiles := GetSpecFiles([]string{dir})
	c.Assert(actualSpecFiles, DeepEquals, expectedSpecFiles)
	FindSpecFilesIn = old
}

func createFileIn(dir string, fileName string, data []byte) (string, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(filepath.Join(dir, fileName), data, 0644)
	return filepath.Join(dir, fileName), err
}

func createDirIn(dir string, dirName string) (string, error) {
	tempDir, err := os.MkdirTemp(dir, dirName)
	if err != nil {
		return "", err
	}

	fullDirName := filepath.Join(dir, dirName)
	err = os.Rename(tempDir, fullDirName)
	return fullDirName, err
}
