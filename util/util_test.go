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
	os.Clearenv()
	os.Setenv(env.ConceptsDir, "path1, path2,    path3")
	c.Assert(GetConceptsPaths(), DeepEquals, []string{"path1", "path2", "path3"})
}
