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

package rerun

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"

	"sort"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {
	p, _ := filepath.Abs("_testdata")
	config.ProjectRoot = p
	runInfo = newLastRunInfo()
}

func (s *MySuite) TestIfFailedFileIsCreated(c *C) {
	msg := "hello world"
	writeLastRunInfo(msg)

	file := filepath.Join(config.ProjectRoot, dotGauge, infoFileName)
	c.Assert(common.FileExists(file), Equals, true)
	expected := msg

	content, _ := ioutil.ReadFile(file)

	c.Assert(string(content), Equals, expected)
	os.RemoveAll(filepath.Join(config.ProjectRoot, dotGauge))
}

func (s *MySuite) TestGetScenarioFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	sce := &gauge.Scenario{Span: &gauge.Span{Start: 2}}
	sr1 := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}}

	prepareScenarioFailedMetadata(sr1, sce, gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: spec1Abs}})

	c.Assert(len(runInfo.failedItemsMap[spec1Abs]), Equals, 1)
	c.Assert(runInfo.failedItemsMap[spec1Abs][spec1Rel+":2"], Equals, true)
}

func (s *MySuite) TestAddSpecPreHookFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}

	addFailedMetadata(spec1, []string{}, addSpecFailedMetadata)

	c.Assert(len(runInfo.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(runInfo.failedItemsMap[spec1Rel][spec1Rel], Equals, true)
}

func (s *MySuite) TestAddSpecPostHookFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PostHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}

	addFailedMetadata(spec1, []string{}, addSpecFailedMetadata)

	c.Assert(len(runInfo.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(runInfo.failedItemsMap[spec1Rel][spec1Rel], Equals, true)
}

func (s *MySuite) TestAddSpecFailedMetadataOverwritesPreviouslyAddedValues(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}
	runInfo.failedItemsMap[spec1Rel] = make(map[string]bool)
	runInfo.failedItemsMap[spec1Rel]["scn1"] = true
	runInfo.failedItemsMap[spec1Rel]["scn2"] = true

	addSpecFailedMetadata(spec1, []string{})

	c.Assert(len(runInfo.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(runInfo.failedItemsMap[spec1Rel][spec1Rel], Equals, true)
}

func (s *MySuite) TestGetRelativePath(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)

	path := util.RelPathToProjectRoot(spec1Abs)

	c.Assert(path, Equals, spec1Rel)
}

func (s *MySuite) TestGetAllFailedItems(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec2Rel := filepath.Join("specs", "example2.spec")
	metaData := newLastRunInfo()
	metaData.failedItemsMap[spec1Rel] = make(map[string]bool)
	metaData.failedItemsMap[spec2Rel] = make(map[string]bool)
	metaData.failedItemsMap[spec1Rel]["scn1"] = true
	metaData.failedItemsMap[spec1Rel]["scn2"] = true
	metaData.failedItemsMap[spec2Rel]["scn3"] = true

	failedItems := metaData.getFailedItems()
	sort.Strings(failedItems)

	c.Assert(failedItems, DeepEquals, []string{"scn1", "scn2", "scn3"})
}

func (s *MySuite) TestCaptureLastRunInfo(c *C) {
	suiteRes := result.NewSuiteResult("", time.Now())
	ei := gauge_messages.ExecutionInfo{}
	suiteEnd := event.NewExecutionEvent(event.SuiteEnd, nil, suiteRes, 1, ei)

	wg := &sync.WaitGroup{}
	event.InitRegistry()
	ListenFailedScenarios(wg, []string{"foo.spec"})
	event.Notify(suiteEnd)
	wg.Wait()

	c.Assert(len(runInfo.Items), Equals, 1)
	c.Assert(runInfo.Items[0], Equals, "foo.spec")

	os.RemoveAll(filepath.Join(config.ProjectRoot, dotGauge))
}
