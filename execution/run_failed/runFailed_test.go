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

package run_failed

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge_messages"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestIfFailedFileIsCreated(c *C) {
	p, _ := filepath.Abs("_testdata")
	config.ProjectRoot = p
	failedInfo = "hello world"

	addFailedInfo()

	file := filepath.Join(config.ProjectRoot, dotGauge, failedFile)
	c.Assert(common.FileExists(file), Equals, true)
	expected := "gauge\n" + failedInfo

	content, _ := ioutil.ReadFile(file)

	c.Assert(string(content), Equals, expected)
	os.RemoveAll(filepath.Join(config.ProjectRoot, dotGauge))
}

func (s *MySuite) TestListenToSpecFailure(c *C) {
	p, _ := filepath.Abs("_testdata")
	failedInfo = ""
	config.ProjectRoot = p
	event.InitRegistry()

	ListenFailedScenarios()
	fileName := filepath.Join(p, "specs", "example.spec")
	sr := &result.SpecResult{IsFailed: true, ProtoSpec: &gauge_messages.ProtoSpec{FileName: &fileName}}
	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, sr, 0))

	c.Assert(failedInfo, Equals, filepath.Join("specs", "example.spec")+"\n")
}

func (s *MySuite) TestListenToSpecPass(c *C) {
	p, _ := filepath.Abs("_testdata")
	failedInfo = ""
	config.ProjectRoot = p
	event.InitRegistry()

	ListenFailedScenarios()
	fileName := filepath.Join(p, "specs", "example.spec")
	sr := &result.SpecResult{IsFailed: false, ProtoSpec: &gauge_messages.ProtoSpec{FileName: &fileName}}
	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, sr, 0))

	c.Assert(failedInfo, Equals, "")
}

func (s *MySuite) TestPrepCommandShouldNotAddUnsetFlags(c *C) {
	obtained := prepareCmd()

	c.Assert(obtained, Equals, "gauge\n")
}

func (s *MySuite) TestPrepareCommand(c *C) {
	Environment = "chrome"
	Tags = "tag1&tag2"
	Verbose = true
	SimpleConsole = false
	TableRows = "1-2"

	obtained := prepareCmd()

	expected := "gauge --env=chrome --tags=tag1&tag2 --tableRows=1-2 --verbose\n"
	c.Assert(obtained, Equals, expected)
}
