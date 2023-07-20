package util

import (
	"os"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestSpecDirEnvVariableAllowsCommaSeparatedList(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := env.LoadEnv("multipleSpecs", nil)
	c.Assert(e, Equals, nil)
	c.Assert(GetSpecDirs(), DeepEquals, []string{"spec1", "spec2", "spec3"})
}

func (s *MySuite) TestConceptsDirEnvVariableAllowsCommaSeparatedList(c *C) {
	c.Skip("This is causing the net/httptest panic in util/httpUtils_test.go fail on windows.")
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := env.LoadEnv("multipleSpecs", nil)
	c.Assert(e, Equals, nil)
	c.Assert(GetConceptsPaths(), DeepEquals, []string{"dir1", "dir2", "dir3"})
}
