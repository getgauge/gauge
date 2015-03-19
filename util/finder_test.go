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
	spec1, err := createFileIn(dir, "gaugeSpec1.spec", data)
	c.Assert(err, Equals, nil)

	dataRead, err := ioutil.ReadFile(spec1)
	c.Assert(err, Equals, nil)
	c.Assert(string(dataRead), Equals, string(data))
	c.Assert(len(findSpecFilesIn(dir)), Equals, 1)

	_, err = createFileIn(dir, "gaugeSpec2.spec", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(findSpecFilesIn(dir)), Equals, 2)

	err = os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}

func createFileIn(dir string, fileName string, data []byte) (string, error) {
	tempFile, err := ioutil.TempFile(dir, "gauge1")
	fullFileName := dir + fmt.Sprintf("%c",filepath.Separator)+fileName
	err = os.Rename(tempFile.Name(), fullFileName)
	err = ioutil.WriteFile(fullFileName, data, 0644)
	return fullFileName, err
}

func (s *MySuite) TestFindAllConceptFiles(c *C) {
	dir, err := ioutil.TempDir("", "gaugeTest")
	c.Assert(err, Equals, nil)
	data := []byte(`#Concept Heading
* Say "hello" to gauge
* Hello gauge
`)
	_, err = createFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(findConceptFilesIn(dir)), Equals, 1)

	_, err = createFileIn(dir, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(findConceptFilesIn(dir)), Equals, 2)

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
	_, err = createFileIn(dir, "concept1.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(findConceptFilesIn(dir)), Equals, 1)

	dir1, err := ioutil.TempDir(dir, "gaugeTest1")
	c.Assert(err, Equals, nil)

	_, err = createFileIn(dir1, "concept2.cpt", data)
	c.Assert(err, Equals, nil)
	c.Assert(len(findConceptFilesIn(dir)), Equals, 2)

	err = os.RemoveAll(dir)
	c.Assert(err, Equals, nil)
}
