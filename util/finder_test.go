package util

import (
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
	"os"
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
	spec1, err := CreateFileIn(dir, "gaugeSpec1.spec", data)
	c.Assert(err, Equals, nil)

	dataRead, err := ioutil.ReadFile(spec1)
	c.Assert(err, Equals, nil)
	c.Assert(string(dataRead), Equals, string(data))
	c.Assert(len(FindSpecFilesIn(dir)), Equals, 1)

	_, err = CreateFileIn(dir, "gaugeSpec2.spec", data)
	c.Assert(err, Equals, nil)

	c.Assert(len(FindSpecFilesIn(dir)), Equals, 2)
}


func (s *MySuite) TestFindAllConceptFiles(c *C) {
	data := []byte(`#Concept Heading`)
	_, err := CreateFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	_, err = CreateFileIn(dir, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)
}

func (s *MySuite) TestFindAllConceptFilesInNestedDir(c *C) {
	data := []byte(`#Concept Heading
* Say "hello" to gauge
`)
	_, err := CreateFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	dir1, err := ioutil.TempDir(dir, "gaugeTest1")
	c.Assert(err, Equals, nil)

	_, err = CreateFileIn(dir1, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)
}
