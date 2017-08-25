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

package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
)

type eventType string
type status string

const (
	suiteStart    eventType = "suiteStart"
	specStart     eventType = "specStart"
	scenarioStart eventType = "scenarioStart"
	scenarioEnd   eventType = "scenarioEnd"
	specEnd       eventType = "specEnd"
	suiteEnd      eventType = "suiteEnd"
	errorResult   eventType = "error"
	pass          status    = "pass"
	fail          status    = "fail"
	skip          status    = "skip"
)

type jsonConsole struct {
	*sync.Mutex
	writer     io.Writer
	isParallel bool
	stream     int
}

type executionEvent struct {
	EventType eventType        `json:"type"`
	ID        string           `json:"id,omitempty"`
	ParentID  string           `json:"parentId,omitempty"`
	Name      string           `json:"name,omitempty"`
	Filename  string           `json:"filename,omitempty"`
	Line      int              `json:"line,omitempty"`
	Stream    int              `json:"stream,omitempty"`
	Res       *executionResult `json:"result,omitempty"`
}

type executionResult struct {
	Status            status           `json:"status,omitempty"`
	Time              int64            `json:"time"`
	Stdout            string           `json:"out,omitempty"`
	Errors            []executionError `json:"errors,omitempty"`
	BeforeHookFailure *executionError  `json:"beforeHookFailure,omitempty"`
	AfterHookFailure  *executionError  `json:"afterHookFailure,omitempty"`
	Table             *tableInfo       `json:"table,omitempty"`
}

type tableInfo struct {
	Text string `json:"text"`
	Row  int    `json:"rowIndex"`
}

type executionError struct {
	Text       string `json:"text"`
	Filename   string `json:"filename"`
	Message    string `json:"message"`
	StackTrace string `json:"stackTrace"`
}

func newJSONConsole(out io.Writer, isParallel bool, stream int) *jsonConsole {
	return &jsonConsole{Mutex: &sync.Mutex{}, writer: out, isParallel: isParallel, stream: stream}
}

func (c *jsonConsole) SuiteStart() {
	c.Lock()
	defer c.Unlock()
	c.write(executionEvent{EventType: suiteStart, Stream: c.stream})
}

func (c *jsonConsole) SuiteEnd(res result.Result) {
	c.Lock()
	defer c.Unlock()
	sRes := res.(*result.SuiteResult)
	c.write(executionEvent{
		EventType: suiteEnd,
		Stream:    c.stream,
		Res: &executionResult{
			Status:            getStatus(sRes.IsFailed, false),
			BeforeHookFailure: getHookFailure(res.GetPreHook(), "Before Suite"),
			AfterHookFailure:  getHookFailure(res.GetPostHook(), "After Suite"),
		},
	})
}

func (c *jsonConsole) SpecStart(spec *gauge.Specification, res result.Result) {
	c.Lock()
	defer c.Unlock()
	addRow := c.isParallel && spec.DataTable.IsInitialized()
	id := getIDWithRow(spec.FileName, spec.Scenarios[0].DataTableRowIndex, addRow)
	c.write(executionEvent{
		EventType: specStart,
		ID:        id,
		Name:      spec.Heading.Value,
		Filename:  spec.FileName,
		Line:      spec.Heading.LineNo,
		Stream:    c.stream,
	})
}

func (c *jsonConsole) SpecEnd(spec *gauge.Specification, res result.Result) {
	c.Lock()
	defer c.Unlock()
	protoSpec := (res.Item().(*gm.ProtoSpec))
	sRes := res.(*result.SpecResult)
	addRow := c.isParallel && spec.DataTable.IsInitialized()
	e := executionEvent{
		EventType: specEnd,
		ID:        getIDWithRow(spec.FileName, spec.Scenarios[0].DataTableRowIndex, addRow),
		Name:      protoSpec.GetSpecHeading(),
		Filename:  spec.FileName,
		Line:      spec.Heading.LineNo,
		Stream:    c.stream,
		Res: &executionResult{
			Status:            getStatus(sRes.GetFailed(), sRes.Skipped),
			BeforeHookFailure: getHookFailure(res.GetPreHook(), "Before Specification"),
			AfterHookFailure:  getHookFailure(res.GetPostHook(), "After Specification"),
		},
	}
	c.write(e)
}

func (c *jsonConsole) ScenarioStart(scenario *gauge.Scenario, i gm.ExecutionInfo, res result.Result) {
	c.Lock()
	defer c.Unlock()
	addRow := c.isParallel && scenario.DataTableRow.IsInitialized()
	parentID := getIDWithRow(i.CurrentSpec.FileName, scenario.DataTableRowIndex, addRow)
	e := executionEvent{
		EventType: scenarioStart,
		ID:        parentID + ":" + strconv.Itoa(scenario.Span.Start),
		ParentID:  parentID,
		Filename:  i.CurrentSpec.FileName,
		Line:      scenario.Heading.LineNo,
		Name:      scenario.Heading.Value,
		Stream:    c.stream,
		Res:       &executionResult{Table: getTable(scenario)},
	}
	c.write(e)
}

func (c *jsonConsole) ScenarioEnd(scenario *gauge.Scenario, res result.Result, i gm.ExecutionInfo) {
	c.Lock()
	defer c.Unlock()
	addRow := c.isParallel && scenario.DataTableRow.IsInitialized()
	parentID := getIDWithRow(i.CurrentSpec.FileName, scenario.DataTableRowIndex, addRow)
	e := executionEvent{
		EventType: scenarioEnd,
		ID:        parentID + ":" + strconv.Itoa(scenario.Span.Start),
		ParentID:  parentID,
		Filename:  i.CurrentSpec.FileName,
		Line:      scenario.Heading.LineNo,
		Name:      scenario.Heading.Value,
		Stream:    c.stream,
		Res: &executionResult{
			Status:            getScenarioStatus(res.(*result.ScenarioResult)),
			Time:              res.ExecTime(),
			Errors:            getErrors(getAllStepsFromScenario(res.(*result.ScenarioResult).ProtoScenario), i.CurrentSpec.FileName),
			BeforeHookFailure: getHookFailure(res.GetPreHook(), "Before Scenario"),
			AfterHookFailure:  getHookFailure(res.GetPostHook(), "After Scenario"),
			Table:             getTable(scenario),
		},
	}
	c.write(e)
}

func getAllStepsFromScenario(scenario *gm.ProtoScenario) []*gm.ProtoItem {
	return append(scenario.GetContexts(), append(scenario.GetScenarioItems(), scenario.GetTearDownSteps()...)...)
}

func (c *jsonConsole) StepStart(stepText string) {
}

func (c *jsonConsole) StepEnd(step gauge.Step, res result.Result, execInfo gm.ExecutionInfo) {
}

func (c *jsonConsole) ConceptStart(conceptHeading string) {
}

func (c *jsonConsole) ConceptEnd(res result.Result) {
}

func (c *jsonConsole) DataTable(table string) {
}

func (c *jsonConsole) Errorf(err string, args ...interface{}) {
	c.Lock()
	defer c.Unlock()
}

func (c *jsonConsole) Write(b []byte) (int, error) {
	c.Lock()
	defer c.Unlock()
	fmt.Fprint(c.writer, string(b))
	return len(b), nil
}

func (c *jsonConsole) write(e executionEvent) {
	b, _ := json.Marshal(e)
	fmt.Fprint(c.writer, string(b)+newline)
}

func getIDWithRow(name string, row int, isDataTable bool) string {
	if !isDataTable {
		return name
	}
	return name + ":" + strconv.Itoa(row)
}

func getScenarioStatus(result *result.ScenarioResult) status {
	return getStatus(result.ProtoScenario.GetExecutionStatus() == gm.ExecutionStatus_FAILED,
		result.ProtoScenario.GetExecutionStatus() == gm.ExecutionStatus_SKIPPED)
}

func getStatus(failed, skipped bool) status {
	if failed {
		return fail
	}
	if skipped {
		return skip
	}
	return pass
}

func getErrors(items []*gm.ProtoItem, id string) (errors []executionError) {
	for _, item := range items {
		executionResult := item.GetStep().GetStepExecutionResult()
		res := executionResult.GetExecutionResult()
		switch item.GetItemType() {
		case gm.ProtoItem_Step:
			filename := util.RelPathToProjectRoot(id)
			if err := executionResult.GetPreHookFailure(); err != nil {
				errors = append(errors, *getHookFailure([]*gm.ProtoHookFailure{err}, "BeforeStep hook for step: "+item.Step.ActualText))
			} else {
				if executionResult.GetSkipped() {
					errors = append(errors, executionError{
						Text:     item.Step.ActualText,
						Filename: filename,
						Message:  executionResult.SkippedReason,
					})
				} else if res.GetFailed() {
					errors = append(errors, executionError{
						Text:       item.Step.ActualText,
						Filename:   filename,
						StackTrace: res.StackTrace,
						Message:    res.ErrorMessage,
					})
				}
			}
			if err := executionResult.GetPostHookFailure(); err != nil {
				errors = append(errors, *getHookFailure([]*gm.ProtoHookFailure{err}, "AfterStep hook for step: "+item.Step.ActualText))
			}
		case gm.ProtoItem_Concept:
			errors = append(errors, getErrors(item.GetConcept().GetSteps(), id)...)
		}
	}
	return
}

func getTable(scenario *gauge.Scenario) *tableInfo {
	if scenario.DataTableRow.IsInitialized() {
		return &tableInfo{
			Text: formatter.FormatTable(&scenario.DataTableRow),
			Row:  scenario.DataTableRowIndex,
		}
	}
	return nil
}

func getHookFailure(hookFailure []*gm.ProtoHookFailure, text string) *executionError {
	if len(hookFailure) > 0 {
		return &executionError{
			Text:       text,
			Message:    hookFailure[0].ErrorMessage,
			StackTrace: hookFailure[0].StackTrace,
		}
	}
	return nil
}
