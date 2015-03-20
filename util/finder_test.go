package util

import (
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
	"os"
	"fmt"
	"path/filepath"
)
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestFindAllSpecFiles(c *C) {
	dir, err := ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
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

	_, err = CreateFileIn(dir, "gaugeSpec2.spec", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindSpecFilesIn(dir)), Equals, 2)

	err = os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}


func (s *MySuite) TestFindAllConceptFiles(c *C) {
	dir, err := ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
	data := []byte(`#Concept Heading
* Say "hello" to gauge
* Hello gauge
`)
	_, err = CreateFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	_, err = CreateFileIn(dir, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)

	err = os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestFindAllConceptFilesInNestedDir(c *C) {
	dir, err := ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)

	data := []byte(`#Concept Heading
* Say "hello" to gauge
* Hello gauge
`)
	_, err = CreateFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 1)

	dir1, err := ioutil.TempDir(dir, "gaugeTest1")
	c.Assert(err, Equals, nil)

	_, err = CreateFileIn(dir1, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(FindConceptFilesIn(dir)), Equals, 2)

	err = os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}
