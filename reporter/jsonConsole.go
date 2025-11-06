/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
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
	pass          status    = "pass"
	fail          status    = "fail"
	skip          status    = "skip"
)

type jsonConsole struct {
	*sync.Mutex
	writer     io.Writer
	isParallel bool
	stream     int
	stepCache  map[*gm.ScenarioInfo][]*stepInfo
}

type stepInfo struct {
	step      *gauge.Step
	protoStep *gm.ProtoStep
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
	LineNo     string `json:"lineNo"`
	StackTrace string `json:"stackTrace"`
}

func newJSONConsole(out io.Writer, isParallel bool, stream int) *jsonConsole {
	return &jsonConsole{Mutex: &sync.Mutex{}, writer: out, isParallel: isParallel, stream: stream, stepCache: make(map[*gm.ScenarioInfo][]*stepInfo)}
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
	c.write(executionEvent{
		EventType: specStart,
		ID:        getIDWithRow(spec.FileName, spec.Scenarios, addRow),
		Name:      spec.Heading.Value,
		Filename:  spec.FileName,
		Line:      spec.Heading.LineNo,
		Stream:    c.stream,
	})
}

func (c *jsonConsole) SpecEnd(spec *gauge.Specification, res result.Result) {
	c.Lock()
	defer c.Unlock()
	protoSpec := res.Item().(*gm.ProtoSpec)
	sRes := res.(*result.SpecResult)
	addRow := c.isParallel && spec.DataTable.IsInitialized()
	e := executionEvent{
		EventType: specEnd,
		ID:        getIDWithRow(spec.FileName, spec.Scenarios, addRow),
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

func (c *jsonConsole) ScenarioStart(scenario *gauge.Scenario, i *gm.ExecutionInfo, res result.Result) {
	c.Lock()
	defer c.Unlock()
	addRow := c.isParallel && scenario.SpecDataTableRow.IsInitialized()
	parentID := getIDWithRow(i.CurrentSpec.FileName, []*gauge.Scenario{scenario}, addRow)
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

func (c *jsonConsole) ScenarioEnd(scenario *gauge.Scenario, res result.Result, i *gm.ExecutionInfo) {
	c.Lock()
	defer c.Unlock()
	addRow := c.isParallel && scenario.SpecDataTableRow.IsInitialized()
	parentID := getIDWithRow(i.CurrentSpec.FileName, []*gauge.Scenario{scenario}, addRow)
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
			Errors:            getErrors(c.stepCache, getAllStepsFromScenario(res.(*result.ScenarioResult).ProtoScenario), i.CurrentSpec.FileName, i),
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

func (c *jsonConsole) StepEnd(step gauge.Step, res result.Result, execInfo *gm.ExecutionInfo) {
	si := &stepInfo{step: &step, protoStep: res.(*result.StepResult).Item().(*gm.ProtoStep)}
	c.stepCache[execInfo.CurrentScenario] = append(c.stepCache[execInfo.CurrentScenario], si)
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
	s := strings.Split(string(b), "\n")
	for _, m := range s {
		outMessage := &logger.OutMessage{MessageType: "out", Message: strings.Trim(m, "\n ")}
		t, err := outMessage.ToJSON()
		if err != nil {
			return 0, err
		}
		_, _ = fmt.Fprintf(c.writer, "%s\n", t)
	}
	return len(b), nil
}

func (c *jsonConsole) write(e executionEvent) {
	b, _ := json.Marshal(e)
	_, _ = fmt.Fprint(c.writer, string(b)+newline)
}

func getIDWithRow(name string, scenarios []*gauge.Scenario, isDataTable bool) string {
	if !isDataTable || len(scenarios) < 1 {
		return name
	}
	return name + ":" + strconv.Itoa(scenarios[0].SpecDataTableRowIndex)
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

func getErrors(stepCache map[*gm.ScenarioInfo][]*stepInfo, items []*gm.ProtoItem, id string, execInfo *gm.ExecutionInfo) (errors []executionError) {
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
						Filename: getFileName(filename, stepCache, item.GetStep(), execInfo),
						Message:  executionResult.SkippedReason,
					})
				} else if res.GetFailed() {
					errors = append(errors, executionError{
						Text:       item.Step.ActualText,
						Filename:   getFileName(filename, stepCache, item.GetStep(), execInfo),
						LineNo:     getLineNo(stepCache, item.GetStep(), execInfo),
						StackTrace: res.StackTrace,
						Message:    res.ErrorMessage,
					})
				}
			}
			if err := executionResult.GetPostHookFailure(); err != nil {
				errors = append(errors, *getHookFailure([]*gm.ProtoHookFailure{err}, "AfterStep hook for step: "+item.Step.ActualText))
			}
		case gm.ProtoItem_Concept:
			errors = append(errors, getErrors(stepCache, item.GetConcept().GetSteps(), id, execInfo)...)
		}
	}
	return
}

func getFileName(file string, stepCache map[*gm.ScenarioInfo][]*stepInfo, step *gm.ProtoStep, info *gm.ExecutionInfo) string {
	for _, si := range stepCache[info.CurrentScenario] {
		if si.protoStep == step {
			return si.step.FileName
		}
	}
	return file
}

func getLineNo(stepCache map[*gm.ScenarioInfo][]*stepInfo, step *gm.ProtoStep, info *gm.ExecutionInfo) string {
	for _, si := range stepCache[info.CurrentScenario] {
		if si.protoStep == step {
			return strconv.Itoa(si.step.LineNo)
		}
	}
	return ""
}

func getTable(scenario *gauge.Scenario) *tableInfo {
	if scenario.ScenarioDataTableRow.IsInitialized() {
		return &tableInfo{
			Text: formatter.FormatTable(&scenario.ScenarioDataTableRow),
			Row:  scenario.ScenarioDataTableRowIndex,
		}
	}
	if scenario.SpecDataTableRow.IsInitialized() {
		return &tableInfo{
			Text: formatter.FormatTable(&scenario.SpecDataTableRow),
			Row:  scenario.SpecDataTableRowIndex,
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
