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

func (s *MySuite) SetUpSuite(c *C){
	var err error
	dir, err = ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TearDownSuite(c *C) {
	err := os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestFindAllSpecFiles(c *C) {
	data := []byte(`Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
`)
	spec1, err := CreateFileIn(dir, "gaugeSpec1.spec", data)
	c.Assert(err, Equals, nil)

	dataRead, err := ioutil.ReadFile(spec1)
	c.Assert(err, Equals, nil)
	c.Assert(string(dataRead), Equals, string(data))
	c.Assert(len(FindSpecFilesIn(dir)), Equals, 1)

	spec2, err := CreateFileIn(dir, "gaugeSpec2.spec", data)
	c.Assert(err, Equals, nil)

	c.Assert(len(FindSpecFilesIn(dir)), Equals, 2)
	defer os.Remove(spec1)
	defer os.Remove(spec2)
}


func (s *MySuite) TestFindAllConceptFiles(c *C) {
	data := []byte(`#Concept Heading
* Say "hello" to gauge
* Hello gauge
`)
	cpt1, err := CreateFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	cpt2, err := CreateFileIn(dir, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)

	defer os.Remove(cpt1)
	defer os.Remove(cpt2)
}

func (s *MySuite) TestFindAllConceptFilesInNestedDir(c *C) {
	data := []byte(`#Concept Heading
* Say "hello" to gauge
* Hello gauge
`)
	cpt1, err := CreateFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	dir1, err := ioutil.TempDir(dir, "gaugeTest1")
	c.Assert(err, Equals, nil)

	cpt2, err := CreateFileIn(dir1, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)

	defer os.Remove(cpt1)
	defer os.Remove(cpt2)
	defer os.Remove(dir1)
}
