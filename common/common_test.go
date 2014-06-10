package common

import (
	"fmt"
	. "launchpad.net/gocheck"
	"os"
	"path/filepath"
	"testing"
)

const dummyProject = "dummy_proj"

func Test(t *testing.T) { TestingT(t) }

type MySuite struct {
	testDir string
}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpSuite(c *C) {
	cwd, _ := os.Getwd()
	s.testDir, _ = filepath.Abs(cwd)
	createDummyProject(dummyProject)
}

func (s *MySuite) SetUpTest(c *C) {
	os.Chdir(s.testDir)
}

func (s *MySuite) TearDownTest(c *C) {
	os.Chdir(s.testDir)
}

func (s *MySuite) TearDownSuite(c *C) {
	os.RemoveAll(dummyProject)
}

func createDummyProject(project string) {
	dirsToCreate := []string{project,
		filepath.Join(project, "specs"),
		filepath.Join(project, "concepts"),
		filepath.Join(project, "specs", "nested"),
		filepath.Join(project, "specs", "nested", "deep_nested"),
		filepath.Join(project, "concepts", "nested"),
		filepath.Join(project, "concepts", "nested", "deep_nested")}

	filesToCreate := []string{filepath.Join(project, ManifestFile),
		filepath.Join(project, "specs", "first.spec"),
		filepath.Join(project, "specs", "second.spec"),
		filepath.Join(project, "specs", "nested", "nested.spec"),
		filepath.Join(project, "specs", "nested", "deep_nested", "deep_nested.spec"),
		filepath.Join(project, "concepts", "first.cpt"),
		filepath.Join(project, "concepts", "nested", "nested.cpt"),
		filepath.Join(project, "concepts", "nested", "deep_nested", "deep_nested.cpt")}

	for _, dirPath := range dirsToCreate {
		os.Mkdir(dirPath, (os.FileMode)(0777))
	}

	for _, filePath := range filesToCreate {
		_, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
	}
}

func (s *MySuite) TestGetProjectRoot(c *C) {
	expectedRoot, _ := filepath.Abs(filepath.Join(dummyProject))
	os.Chdir(dummyProject)

	root, err := GetProjectRoot()

	c.Assert(err, IsNil)
	c.Assert(root, Equals, expectedRoot)
}

func (s *MySuite) TestGetProjectRootFromNestedDir(c *C) {
	expectedRoot, _ := filepath.Abs(filepath.Join(dummyProject))
	os.Chdir(filepath.Join(dummyProject, "specs", "nested", "deep_nested"))

	root, err := GetProjectRoot()

	c.Assert(err, IsNil)
	c.Assert(root, Equals, expectedRoot)
}

func (s *MySuite) TestGetProjectFailing(c *C) {

	_, err := GetProjectRoot()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Failed to find project directory, run the command inside the project")
}

func (s *MySuite) TestGetDirInProject(c *C) {
	os.Chdir(dummyProject)

	concepts, err := GetDirInProject("concepts")

	c.Assert(err, IsNil)
	c.Assert(concepts, Equals, filepath.Join(s.testDir, dummyProject, "concepts"))
}

func (s *MySuite) TestGetDirInProjectFromNestedDir(c *C) {
	os.Chdir(filepath.Join(dummyProject, "specs", "nested", "deep_nested"))

	concepts, err := GetDirInProject("concepts")

	c.Assert(err, IsNil)
	c.Assert(concepts, Equals, filepath.Join(s.testDir, dummyProject, "concepts"))
}

func (s *MySuite) TestGetNotExistingDirInProject(c *C) {
	os.Chdir(filepath.Join(dummyProject, "specs", "nested", "deep_nested"))

	_, err := GetDirInProject("invalid")

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, fmt.Sprintf("Could not find invalid directory. %s does not exist", filepath.Join(s.testDir, dummyProject, "invalid")))
}

func (s *MySuite) TestFindFilesInDir(c *C) {
	foundSpecFiles := FindFilesInDir(filepath.Join(dummyProject, "specs"), func(filePath string) bool {
		return filepath.Ext(filePath) == ".spec"
	})

	c.Assert(len(foundSpecFiles), Equals, 4)

	foundConceptFiles := FindFilesInDir(filepath.Join(dummyProject, "concepts"), func(filePath string) bool {
		return filepath.Ext(filePath) == ".cpt"
	})

	c.Assert(len(foundConceptFiles), Equals, 3)
}

func (s *MySuite) TestFileExists(c *C) {
	c.Assert(FileExists(filepath.Join(dummyProject, ManifestFile)), Equals, true)
	c.Assert(FileExists("invalid"), Equals, false)
}
