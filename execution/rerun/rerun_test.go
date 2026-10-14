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
	c.Assert(failedMeta.failedItemsMap[spec1Abs][failureKey{filePath: spec1Rel, line: 2}], Equals, true)
}

func (s *MySuite) TestScenarioPassingOnRetryRemovesFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	sce := &gauge.Scenario{Span: &gauge.Span{Start: 2}}
	execInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: spec1Abs}}
	failedResult := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}}
	passedResult := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_PASSED}}

	prepareScenarioFailedMetadata(failedResult, sce, execInfo)
	c.Assert(failedMeta.failedItemsMap[spec1Abs][failureKey{filePath: spec1Rel, line: 2}], Equals, true)

	prepareScenarioFailedMetadata(passedResult, sce, execInfo)
	_, exists := failedMeta.failedItemsMap[spec1Abs]
	c.Assert(exists, Equals, false)
}

func (s *MySuite) TestPassingTableRowDoesNotRemoveFailedTableRowMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	specTableRow := *gauge.NewTable([]string{"Word"}, [][]gauge.TableCell{{{Value: "Snap", CellType: gauge.Static}}}, 0)
	execInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: spec1Abs}}
	failedScenario := &gauge.Scenario{
		Span:                  &gauge.Span{Start: 13},
		HasSpecDataTable:      true,
		SpecDataTableRow:      specTableRow,
		SpecDataTableRowIndex: 2,
	}
	passedScenario := &gauge.Scenario{
		Span:                  &gauge.Span{Start: 13},
		HasSpecDataTable:      true,
		SpecDataTableRow:      specTableRow,
		SpecDataTableRowIndex: 3,
	}
	failedResult := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}}
	passedResult := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_PASSED}}

	prepareScenarioFailedMetadata(failedResult, failedScenario, execInfo)
	c.Assert(failedMeta.failedItemsMap[spec1Abs][failureKey{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 2}], Equals, true)

	prepareScenarioFailedMetadata(passedResult, passedScenario, execInfo)
	c.Assert(failedMeta.failedItemsMap[spec1Abs][failureKey{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 2}], Equals, true)
}

func (s *MySuite) TestScenarioFailureRefWithScenarioDataTableRow(c *C) {
	scenarioTableRow := *gauge.NewTable([]string{"Color"}, [][]gauge.TableCell{{{Value: "Red", CellType: gauge.Static}}}, 0)
	sce := &gauge.Scenario{
		Span:                      &gauge.Span{Start: 5},
		ScenarioDataTableRow:      scenarioTableRow,
		ScenarioDataTableRowIndex: 3,
	}

	key := newScenarioFailureKey("specs/example.spec", sce)

	c.Assert(key.filePath, Equals, "specs/example.spec")
	c.Assert(key.line, Equals, 5)
	c.Assert(key.hasSpecDataTableRow, Equals, false)
	c.Assert(key.hasScenarioDataTableRow, Equals, true)
	c.Assert(key.scenarioDataTableRow, Equals, 3)
}

func (s *MySuite) TestScenarioFailureRefWithBothDataTableRows(c *C) {
	specTableRow := *gauge.NewTable([]string{"Word"}, [][]gauge.TableCell{{{Value: "Snap", CellType: gauge.Static}}}, 0)
	scenarioTableRow := *gauge.NewTable([]string{"Color"}, [][]gauge.TableCell{{{Value: "Red", CellType: gauge.Static}}}, 0)
	sce := &gauge.Scenario{
		Span:                      &gauge.Span{Start: 5},
		HasSpecDataTable:          true,
		SpecDataTableRow:          specTableRow,
		SpecDataTableRowIndex:     2,
		ScenarioDataTableRow:      scenarioTableRow,
		ScenarioDataTableRowIndex: 1,
	}

	key := newScenarioFailureKey("specs/example.spec", sce)

	c.Assert(key.filePath, Equals, "specs/example.spec")
	c.Assert(key.line, Equals, 5)
	c.Assert(key.hasSpecDataTableRow, Equals, true)
	c.Assert(key.specDataTableRow, Equals, 2)
	c.Assert(key.hasScenarioDataTableRow, Equals, true)
	c.Assert(key.scenarioDataTableRow, Equals, 1)
}

func (s *MySuite) TestSameTableRowPassingOnRetryRemovesFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	specTableRow := *gauge.NewTable([]string{"Word"}, [][]gauge.TableCell{{{Value: "Snap", CellType: gauge.Static}}}, 0)
	execInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: spec1Abs}}
	sce := &gauge.Scenario{
		Span:                  &gauge.Span{Start: 13},
		HasSpecDataTable:      true,
		SpecDataTableRow:      specTableRow,
		SpecDataTableRowIndex: 2,
	}
	failedResult := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_FAILED}}
	passedResult := &result.ScenarioResult{ProtoScenario: &gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_PASSED}}

	prepareScenarioFailedMetadata(failedResult, sce, execInfo)
	c.Assert(failedMeta.failedItemsMap[spec1Abs][failureKey{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 2}], Equals, true)

	prepareScenarioFailedMetadata(passedResult, sce, execInfo)
	_, exists := failedMeta.failedItemsMap[spec1Abs]
	c.Assert(exists, Equals, false)
}

func (s *MySuite) TestGetFailedItemsUsesFileLineForTableDrivenScenarios(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	metaData := newFailedMetaData()
	metaData.failedItemsMap[spec1Rel] = map[failureKey]bool{
		{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 2}:                                                                    true,
		{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 4, hasScenarioDataTableRow: true, scenarioDataTableRow: 1}: true,
	}

	failedItems := metaData.getFailedItems()
	sort.Strings(failedItems)

	c.Assert(failedItems, DeepEquals, []string{spec1Rel + ":13"})
}

func (s *MySuite) TestScenarioFailureRefDistinguishesSpecRowsForNestedTable(c *C) {
	scenarioTableRow := *gauge.NewTable([]string{"Color"}, [][]gauge.TableCell{{{Value: "Red", CellType: gauge.Static}}}, 0)
	sce0 := &gauge.Scenario{
		Span:                      &gauge.Span{Start: 13},
		HasSpecDataTable:          true,
		SpecDataTableRowIndex:     0,
		ScenarioDataTableRow:      scenarioTableRow,
		ScenarioDataTableRowIndex: 1,
	}
	sce1 := &gauge.Scenario{
		Span:                      &gauge.Span{Start: 13},
		HasSpecDataTable:          true,
		SpecDataTableRowIndex:     1,
		ScenarioDataTableRow:      scenarioTableRow,
		ScenarioDataTableRowIndex: 1,
	}

	key0 := newScenarioFailureKey("specs/example.spec", sce0)
	key1 := newScenarioFailureKey("specs/example.spec", sce1)

	c.Assert(key0, Not(Equals), key1)
	c.Assert(key0.outputRef(), Equals, "specs/example.spec:13")
	c.Assert(key1.outputRef(), Equals, "specs/example.spec:13")
}

func (s *MySuite) TestGetFailedItemsDeduplicatesNestedTableScenarios(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	metaData := newFailedMetaData()
	metaData.failedItemsMap[spec1Rel] = map[failureKey]bool{
		{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 0, hasScenarioDataTableRow: true, scenarioDataTableRow: 1}: true,
		{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 1, hasScenarioDataTableRow: true, scenarioDataTableRow: 1}: true,
	}

	failedItems := metaData.getFailedItems()

	c.Assert(failedItems, DeepEquals, []string{spec1Rel + ":13"})
}

func (s *MySuite) TestAddSpecPreHookFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}

	addFailedMetadata(spec1, []string{}, addSpecFailedMetadata)

	c.Assert(len(failedMeta.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Rel][failureKey{filePath: spec1Rel}], Equals, true)
}

func (s *MySuite) TestAddSpecPostHookFailedMetadata(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PostHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}

	addFailedMetadata(spec1, []string{}, addSpecFailedMetadata)

	c.Assert(len(failedMeta.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Rel][failureKey{filePath: spec1Rel}], Equals, true)
}

func (s *MySuite) TestAddSpecFailedMetadataOverwritesPreviouslyAddedValues(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)
	spec1 := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailures: []*gauge_messages.ProtoHookFailure{{ErrorMessage: "error"}}, FileName: spec1Abs}}
	failedMeta.failedItemsMap[spec1Rel] = make(map[failureKey]bool)
	failedMeta.failedItemsMap[spec1Rel][failureKey{filePath: spec1Rel, line: 1}] = true
	failedMeta.failedItemsMap[spec1Rel][failureKey{filePath: spec1Rel, line: 2}] = true

	addSpecFailedMetadata(spec1, []string{})

	c.Assert(len(failedMeta.failedItemsMap[spec1Rel]), Equals, 1)
	c.Assert(failedMeta.failedItemsMap[spec1Rel][failureKey{filePath: spec1Rel}], Equals, true)
}

func (s *MySuite) TestGetRelativePath(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(config.ProjectRoot, spec1Rel)

	path := util.RelPathToProjectRoot(spec1Abs)

	c.Assert(path, Equals, spec1Rel)
}

func (s *MySuite) TestGetFailedItemsWithCustomSpecExtension(c *C) {
	spec1Rel := filepath.Join("specs", "example1.foo")
	metaData := newFailedMetaData()
	metaData.failedItemsMap[spec1Rel] = map[failureKey]bool{
		{filePath: spec1Rel, line: 13, hasSpecDataTableRow: true, specDataTableRow: 2}: true,
	}

	failedItems := metaData.getFailedItems()

	c.Assert(failedItems, DeepEquals, []string{spec1Rel + ":13"})
}

func (s *MySuite) TestGetAllFailedItems(c *C) {
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec2Rel := filepath.Join("specs", "example2.spec")
	metaData := newFailedMetaData()
	metaData.failedItemsMap[spec1Rel] = make(map[failureKey]bool)
	metaData.failedItemsMap[spec2Rel] = make(map[failureKey]bool)
	metaData.failedItemsMap[spec1Rel][failureKey{filePath: "scn1"}] = true
	metaData.failedItemsMap[spec1Rel][failureKey{filePath: "scn2"}] = true
	metaData.failedItemsMap[spec2Rel][failureKey{filePath: "scn3"}] = true

	failedItems := metaData.getFailedItems()
	sort.Strings(failedItems)

	c.Assert(failedItems, DeepEquals, []string{"scn1", "scn2", "scn3"})
}
