/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package rerun

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
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
	failedMeta = newFailedMetaData()
}

func (s *MySuite) TestIfFailedFileIsCreated(c *C) {
	msg := "hello world"

	writeFailedMeta(msg)

	file := filepath.Join(config.ProjectRoot, common.DotGauge, failedFile)
	c.Assert(common.FileExists(file), Equals, true)
	expected := msg

	content, _ := os.ReadFile(file)

	c.Assert(string(content), Equals, expected)
	_ = os.RemoveAll(filepath.Join(config.ProjectRoot, common.DotGauge))
}

func (s *MySuite) TestGetScenarioFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	sce := &gauge.Scenario{Span: &gauge.Span{Start: 2}}
	sr1 := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}}

	prepareScenarioFailedMetadata(sr1, sce, &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: spec1Abs}})

	c.Assert(len(failedMeta.failedItemsMap[spec1Abs]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Abs][spec1Rel+":2"], Equals, true)
}

func (s *MySuite) TestAddSpecPreHookFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}

	addFailedMetadata(spec1, []string{}, addSpecFailedMetadata)

	c.Assert(len(failedMeta.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Rel][spec1Rel], Equals, true)
}

func (s *MySuite) TestAddSpecPostHookFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PostHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}

	addFailedMetadata(spec1, []string{}, addSpecFailedMetadata)

	c.Assert(len(failedMeta.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Rel][spec1Rel], Equals, true)
}

func (s *MySuite) TestAddSpecFailedMetadataOverwritesPreviouslyAddedValues(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}
	failedMeta.failedItemsMap[spec1Rel] = make(map[string]bool)
	failedMeta.failedItemsMap[spec1Rel]["scn1"] = true
	failedMeta.failedItemsMap[spec1Rel]["scn2"] = true

	addSpecFailedMetadata(spec1, []string{})

	c.Assert(len(failedMeta.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Rel][spec1Rel], Equals, true)
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
	metaData := newFailedMetaData()
	metaData.failedItemsMap[spec1Rel] = make(map[string]bool)
	metaData.failedItemsMap[spec2Rel] = make(map[string]bool)
	metaData.failedItemsMap[spec1Rel]["scn1"] = true
	metaData.failedItemsMap[spec1Rel]["scn2"] = true
	metaData.failedItemsMap[spec2Rel]["scn3"] = true

	failedItems := metaData.getFailedItems()
	sort.Strings(failedItems)

	c.Assert(failedItems, DeepEquals, []string{"scn1", "scn2", "scn3"})
}
