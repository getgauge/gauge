/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"fmt"
	"net"
	"testing"

	"github.com/getgauge/gauge/runner"

	"sync"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/validation"
	. "gopkg.in/check.v1"
)

type specBuilder struct {
	lines []string
}

func newSpecBuilder() *specBuilder {
	return &specBuilder{lines: make([]string, 0)}
}

func (specBuilder *specBuilder) addPrefix(prefix string, line string) string {
	return fmt.Sprintf("%s%s\n", prefix, line)
}

func (specBuilder *specBuilder) String() string {
	var specResult string
	for _, line := range specBuilder.lines {
		specResult = fmt.Sprintf("%s%s", specResult, line)
	}
	return specResult
}

func (specBuilder *specBuilder) specHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("#", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) scenarioHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("##", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) step(stepText string) *specBuilder {
	line := specBuilder.addPrefix("* ", stepText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) tableHeader(cells ...string) *specBuilder {
	return specBuilder.tableRow(cells...)
}
func (specBuilder *specBuilder) tableRow(cells ...string) *specBuilder {
	rowInMarkdown := "|"
	for _, cell := range cells {
		rowInMarkdown = fmt.Sprintf("%s%s|", rowInMarkdown, cell)
	}
	specBuilder.lines = append(specBuilder.lines, fmt.Sprintf("%s\n", rowInMarkdown))
	return specBuilder
}

type tableRow struct {
	name   string
	input  string // input by user for data table rows
	output []int  // data table indexes to be executed
}

var tableRowTests = []*tableRow{
	{"Valid single row number", "2", []int{1}},
	{"Valid row numbers list", "2,3,4", []int{1, 2, 3}},
	{"Valid table rows range", "2-5", []int{1, 2, 3, 4}},
	{"Empty table rows range", "", []int(nil)},
	{"Table rows list with spaces", "2, 4 ", []int{1, 3}},
}

func (s *MySuite) TestToGetDataTableRowsRangeFromInputFlag(c *C) {
	for _, test := range tableRowTests {
		got := getDataTableRows(test.input)
		want := test.output
		c.Assert(got, DeepEquals, want, Commentf(test.name))
	}
}

func (s *MySuite) TestCreateSkippedSpecResult(c *C) {
	spec := &gauge.Specification{Heading: &gauge.Heading{LineNo: 0, Value: "SPEC_HEADING"}, FileName: "FILE"}
	r := &mockRunner{}
	se := newSpecExecutor(spec, r, nil, nil, 0)
	se.errMap = getValidationErrorMap()
	se.specResult = &result.SpecResult{}
	se.skipSpecForError(fmt.Errorf("ERROR"))

	c.Assert(se.specResult.IsFailed, Equals, false)
	c.Assert(se.specResult.Skipped, Equals, true)
	c.Assert(len(se.errMap.SpecErrs[spec]), Equals, 1)
}

func (s *MySuite) TestCreateSkippedSpecResultWithScenarios(c *C) {
	r := &mockRunner{}
	se := newSpecExecutor(anySpec(), r, nil, nil, 0)
	se.errMap = getValidationErrorMap()
	se.specResult = &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{}}
	se.skipSpecForError(fmt.Errorf("ERROR"))

	c.Assert(len(se.errMap.ScenarioErrs[se.specification.Scenarios[0]]), Equals, 1)
	c.Assert(len(se.errMap.SpecErrs[se.specification]), Equals, 1)
}

func anySpec() *gauge.Specification {

	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	spec, _, _ := new(parser.SpecParser).Parse(specText, gauge.NewConceptDictionary(), "")
	spec.FileName = "FILE"
	return spec
}

func (s *MySuite) TestSpecIsSkippedIfDataRangeIsInvalid(c *C) {
	errMap := &gauge.BuildErrors{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}
	r := &mockRunner{}
	spec := anySpec()
	errMap.SpecErrs[spec] = []error{validation.NewSpecValidationError("Table row number out of range", spec.FileName)}
	se := newSpecExecutor(spec, r, nil, errMap, 0)

	specResult := se.execute(true, false, false)
	c.Assert(specResult.Skipped, Equals, true)
}

func (s *MySuite) TestDataTableRowsAreSkippedForUnimplemetedStep(c *C) {
	MaxRetriesCount = 1
	stepText := "Unimplememted step"

	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableRow("123", "foo", "8800").
		tableRow("666", "bar", "9900").
		scenarioHeading("First scenario").
		step(stepText).
		step("create user <id> <name> and <phone>").
		String()

	spec, _, _ := new(parser.SpecParser).Parse(specText, gauge.NewConceptDictionary(), "")

	errMap := &gauge.BuildErrors{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}
	r := &mockRunner{}
	errMap.SpecErrs[spec] = []error{validation.NewSpecValidationError("Step implementation not found", spec.FileName)}
	errMap.ScenarioErrs[spec.Scenarios[0]] = []error{validation.NewSpecValidationError("Step implementation not found", spec.FileName)}
	se := newSpecExecutor(spec, r, nil, errMap, 0)

	specResult := se.execute(true, true, true)
	c.Assert(specResult.ProtoSpec.GetIsTableDriven(), Equals, true)
	c.Assert(specResult.Skipped, Equals, true)
}

func (s *MySuite) TestConvertParseErrorToGaugeMessagesError(c *C) {
	spec := &gauge.Specification{Heading: &gauge.Heading{LineNo: 0, Value: "SPEC_HEADING"}, FileName: "FILE"}
	e := parser.ParseError{Message: "Message", LineNo: 5, FileName: "filename"}
	r := &mockRunner{}
	se := newSpecExecutor(spec, r, nil, nil, 0)

	errs := se.convertErrors([]error{e})

	expected := &gauge_messages.Error{
		Type:       gauge_messages.Error_PARSE_ERROR,
		Message:    "filename:5 Message => ''",
		LineNumber: 5,
		Filename:   "filename",
	}

	c.Assert(len(errs), DeepEquals, 1)
	c.Assert(errs[0], DeepEquals, expected)
}

func (s *MySuite) TestConvertSpecValidationErrorToGaugeMessagesError(c *C) {
	spec := &gauge.Specification{Heading: &gauge.Heading{LineNo: 0, Value: "SPEC_HEADING"}, FileName: "FILE"}
	e := validation.NewSpecValidationError("Message", "filename")
	r := &mockRunner{}
	se := newSpecExecutor(spec, r, nil, nil, 0)

	errs := se.convertErrors([]error{e})

	expected := &gauge_messages.Error{
		Type:    gauge_messages.Error_VALIDATION_ERROR,
		Message: "filename Message",
	}

	c.Assert(len(errs), DeepEquals, 1)
	c.Assert(errs[0], DeepEquals, expected)
}

func (s *MySuite) TestConvertStepValidationErrorToGaugeMessagesError(c *C) {
	spec := &gauge.Specification{Heading: &gauge.Heading{LineNo: 0, Value: "SPEC_HEADING"}, FileName: "FILE"}
	e := validation.NewStepValidationError(&gauge.Step{LineText: "step", LineNo: 3}, "Step Message", "filename", nil, "")
	r := &mockRunner{}
	se := newSpecExecutor(spec, r, nil, nil, 0)

	errs := se.convertErrors([]error{e})

	expected := &gauge_messages.Error{
		Type:    gauge_messages.Error_VALIDATION_ERROR,
		Message: "filename:3 Step Message => 'step'",
	}

	c.Assert(len(errs), DeepEquals, 1)
	c.Assert(errs[0], DeepEquals, expected)
}

type mockRunner struct {
	ExecuteAndGetStatusFunc func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult
}

func (r *mockRunner) ExecuteMessageWithTimeout(m *gauge_messages.Message) (*gauge_messages.Message, error) {
	return nil, nil
}
func (r *mockRunner) ExecuteAndGetStatus(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	return r.ExecuteAndGetStatusFunc(m)
}

func (r *mockRunner) Alive() bool {
	return false
}

func (r *mockRunner) Kill() error {
	return nil
}

func (r *mockRunner) Connection() net.Conn {
	return nil
}

func (r *mockRunner) IsMultithreaded() bool {
	return false
}

func (r *mockRunner) Info() *runner.RunnerInfo {
	return &runner.RunnerInfo{Killed: false}
}

func (r *mockRunner) Pid() int {
	return -1
}

type mockPluginHandler struct {
	NotifyPluginsfunc         func(*gauge_messages.Message)
	GracefullyKillPluginsfunc func()
}

func (h *mockPluginHandler) NotifyPlugins(m *gauge_messages.Message) {
	h.NotifyPluginsfunc(m)
}

func (h *mockPluginHandler) GracefullyKillPlugins() {
	h.GracefullyKillPluginsfunc()
}

func (h *mockPluginHandler) ExtendTimeout(id string) {

}

var exampleSpec = &gauge.Specification{Heading: &gauge.Heading{Value: "Example Spec"}, FileName: "example.spec", Tags: &gauge.Tags{}}

var exampleSpecWithScenarios = &gauge.Specification{
	Heading:  &gauge.Heading{Value: "Example Spec"},
	FileName: "example.spec",
	Tags:     &gauge.Tags{},
	Scenarios: []*gauge.Scenario{
		{Heading: &gauge.Heading{Value: "Example Scenario 1"}, Items: make([]gauge.Item, 0), Tags: &gauge.Tags{}, Span: &gauge.Span{}},
		{Heading: &gauge.Heading{Value: "Example Scenario 2"}, Items: make([]gauge.Item, 0), Tags: &gauge.Tags{}, Span: &gauge.Span{}},
	},
}

func TestExecuteFailsWhenSpecHasParseErrors(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	errs.SpecErrs[exampleSpec] = append(errs.SpecErrs[exampleSpec], parser.ParseError{Message: "some error"})
	se := newSpecExecutor(exampleSpec, r, nil, errs, 0)

	res := se.execute(false, true, false)

	if !res.GetFailed() {
		t.Errorf("Expected result.Failed=true, got %t", res.GetFailed())
	}

	c := len(res.Errors)
	if c != 1 {
		t.Errorf("Expected result to contain 1 error, got %d", c)
	}
}

func TestExecuteSkipsWhenSpecHasErrors(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	errs.SpecErrs[exampleSpec] = append(errs.SpecErrs[exampleSpec], fmt.Errorf("some error"))
	se := newSpecExecutor(exampleSpec, r, nil, errs, 0)

	res := se.execute(false, true, false)

	if !res.Skipped {
		t.Errorf("Expected result.Skipped=true, got %t", res.Skipped)
	}
}

func TestExecuteInitSpecDatastore(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	dataStoreInitCalled := false
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecDataStoreInit {
			dataStoreInitCalled = true
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	se := newSpecExecutor(exampleSpecWithScenarios, r, h, errs, 0)
	se.execute(true, false, false)

	if !dataStoreInitCalled {
		t.Error("Expected runner to be called with SpecDataStoreInit")
	}
}

func TestExecuteShouldNotInitSpecDatastoreWhenBeforeIsFalse(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}

	dataStoreInitCalled := false
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecDataStoreInit {
			dataStoreInitCalled = true
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	se := newSpecExecutor(exampleSpec, r, nil, errs, 0)
	se.execute(false, false, false)

	if dataStoreInitCalled {
		t.Error("Expected SpecDataStoreInit to not be called")
	}
}

func TestExecuteSkipsWhenSpecDatastoreInitFails(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}

	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "datastore init error"}
	}
	se := newSpecExecutor(exampleSpecWithScenarios, r, nil, errs, 0)
	res := se.execute(true, false, false)

	if !res.Skipped {
		t.Errorf("Expected result.Skipped=true, got %t", res.Skipped)
	}

	e := res.Errors[0]
	expected := "example.spec:0 Failed to initialize spec datastore. Error: datastore init error => 'Example Spec'"
	if e.Message != expected {
		t.Errorf("Expected error = '%s', got '%s'", expected, e.Message)
	}
}

func TestExecuteBeforeSpecHook(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}

	beforeSpecHookCalled := false
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecExecutionStarting {
			beforeSpecHookCalled = true
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	se := newSpecExecutor(exampleSpecWithScenarios, r, h, errs, 0)
	se.execute(true, false, false)

	if !beforeSpecHookCalled {
		t.Error("Expected runner to be called with SpecExecutionStarting")
	}
}

func TestExecuteShouldNotifyBeforeSpecEvent(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}

	eventRaised := false
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		return &gauge_messages.ProtoExecutionResult{}
	}

	ch := make(chan event.ExecutionEvent)
	event.InitRegistry()
	event.Register(ch, event.SpecStart)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			e := <-ch
			t.Log(e.Topic)
			if e.Topic == event.SpecStart {
				eventRaised = true
				wg.Done()
			}
		}
	}()
	se := newSpecExecutor(exampleSpecWithScenarios, r, h, errs, 0)
	se.execute(true, false, false)

	wg.Wait()
	if !eventRaised {
		t.Error("Expected SpecStart event to be raised")
	}
	event.InitRegistry()
}
func TestExecuteAfterSpecHook(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}

	afterSpecHookCalled := false
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		switch m.MessageType {
		case gauge_messages.Message_SpecExecutionEnding:
			afterSpecHookCalled = true
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	se := newSpecExecutor(exampleSpecWithScenarios, r, h, errs, 0)
	se.execute(false, false, true)

	if !afterSpecHookCalled {
		t.Error("Expected runner to be called with SpecExecutionAfter")
	}
}

func TestExecuteAddsSpecHookExecutionMessages(t *testing.T) {
	errs := gauge.NewBuildErrors()
	mockRunner := &mockRunner{}
	mockHandler := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}

	mockRunner.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		switch m.MessageType {
		case gauge_messages.Message_SpecExecutionEnding:
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"After Spec Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		case gauge_messages.Message_SpecExecutionStarting:
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Before Spec Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	se := newSpecExecutor(exampleSpec, mockRunner, mockHandler, errs, 0)
	se.execute(true, false, true)

	gotPreHookMessages := se.specResult.ProtoSpec.PreHookMessages
	gotPostHookMessages := se.specResult.ProtoSpec.PostHookMessages

	if len(gotPreHookMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotPreHookMessages))
	}
	if gotPreHookMessages[0] != "Before Spec Called" {
		t.Errorf("Expected `Before Spec Called` message, got : %s", gotPreHookMessages[0])
	}
	if len(gotPostHookMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotPostHookMessages))
	}
	if gotPostHookMessages[0] != "After Spec Called" {
		t.Errorf("Expected `After Spec Called` message, got : %s", gotPostHookMessages[0])
	}
}

func TestExecuteAddsSpecHookExecutionScreenshots(t *testing.T) {
	errs := gauge.NewBuildErrors()
	mockRunner := &mockRunner{}
	mockHandler := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}

	mockRunner.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		switch m.MessageType {
		case gauge_messages.Message_SpecExecutionEnding:
			return &gauge_messages.ProtoExecutionResult{
				ScreenshotFiles: []string{"screenshot1.png", "screenshot2.png"},
				Failed:          false,
				ExecutionTime:   10,
			}
		case gauge_messages.Message_SpecExecutionStarting:
			return &gauge_messages.ProtoExecutionResult{
				ScreenshotFiles: []string{"screenshot3.png", "screenshot4.png"},
				Failed:          false,
				ExecutionTime:   10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	se := newSpecExecutor(exampleSpec, mockRunner, mockHandler, errs, 0)
	se.execute(true, false, true)

	beforeSpecScreenshots := se.specResult.ProtoSpec.PreHookScreenshotFiles
	afterSpecScreenshots := se.specResult.ProtoSpec.PostHookScreenshotFiles
	expectedAfterSpecScreenshots := []string{"screenshot1.png", "screenshot2.png"}
	expectedBeforeSpecScreenshots := []string{"screenshot3.png", "screenshot4.png"}

	if len(beforeSpecScreenshots) != len(expectedBeforeSpecScreenshots) {
		t.Errorf("Expected 2 screenshots, got : %d", len(beforeSpecScreenshots))
	}
	for i, e := range expectedBeforeSpecScreenshots {
		if beforeSpecScreenshots[i] != e {
			t.Errorf("Expected `%s` screenshot, got : %s", e, beforeSpecScreenshots[i])
		}
	}
	if len(afterSpecScreenshots) != len(expectedAfterSpecScreenshots) {
		t.Errorf("Expected 2 screenshots, got : %d", len(afterSpecScreenshots))
	}
	for i, e := range expectedAfterSpecScreenshots {
		if afterSpecScreenshots[i] != e {
			t.Errorf("Expected `%s` screenshot, got : %s", e, afterSpecScreenshots[i])
		}
	}
}

func TestExecuteShouldNotifyAfterSpecEvent(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}

	eventRaised := false
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		return &gauge_messages.ProtoExecutionResult{}
	}

	ch := make(chan event.ExecutionEvent)
	event.InitRegistry()
	event.Register(ch, event.SpecEnd)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			e := <-ch
			t.Log(e.Topic)
			if e.Topic == event.SpecEnd {
				eventRaised = true
				wg.Done()
			}
		}
	}()
	se := newSpecExecutor(exampleSpecWithScenarios, r, h, errs, 0)
	se.execute(false, false, true)

	wg.Wait()
	if !eventRaised {
		t.Error("Expected SpecEnd event to be raised")
	}
	event.InitRegistry()
}

type mockExecutor struct {
	executeFunc func(i gauge.Item, r result.Result)
}

func (e *mockExecutor) execute(i gauge.Item, r result.Result) {
	e.executeFunc(i, r)
}

func TestExecuteScenario(t *testing.T) {
	MaxRetriesCount = 1
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	se := newSpecExecutor(exampleSpecWithScenarios, r, nil, errs, 0)
	executedScenarios := make([]string, 0)
	se.scenarioExecutor = &mockExecutor{
		executeFunc: func(i gauge.Item, r result.Result) {
			executedScenarios = append(executedScenarios, i.(*gauge.Scenario).Heading.Value)
		},
	}
	se.execute(false, true, false)
	got := len(executedScenarios)
	if got != 2 {
		t.Errorf("Expected 2 scenarios to be executed, got %d", got)
	}

	expected := []string{"Example Scenario 1", "Example Scenario 2"}
	for i, s := range executedScenarios {
		if s != expected[i] {
			t.Errorf("Expected '%s' scenario to be executed. Got %s", s, executedScenarios)
		}
	}
}

func TestExecuteScenarioWithRetries(t *testing.T) {
	MaxRetriesCount = 3
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	se := newSpecExecutor(exampleSpecWithScenarios, r, nil, errs, 0)

	count := 1
	se.scenarioExecutor = &mockExecutor{
		executeFunc: func(i gauge.Item, r result.Result) {
			if count < MaxRetriesCount {
				r.SetFailure()
			} else {
				r.(*result.ScenarioResult).ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_PASSED
			}

			count++
		},
	}

	sceResult, _ := se.executeScenario(exampleSpecWithScenarios.Scenarios[0])

	if sceResult.GetFailed() {
		t.Errorf("Expect sceResult.GetFailed() = false, got true")
	}
}

var exampleSpecWithTags = &gauge.Specification{
	Heading:  &gauge.Heading{Value: "Example Spec"},
	FileName: "example.spec",
	Tags:     &gauge.Tags{RawValues: [][]string{{"tagSpec"}}},
	Scenarios: []*gauge.Scenario{
		{Heading: &gauge.Heading{Value: "Example Scenario 1"}, Items: make([]gauge.Item, 0), Tags: &gauge.Tags{RawValues: [][]string{{"tagSce"}}}, Span: &gauge.Span{}},
	},
}

func TestExecuteScenarioShouldNotRetryIfNotMatchTags(t *testing.T) {
	MaxRetriesCount = 2
	RetryOnlyTags = "tagN"

	se := newSpecExecutorForTestsWithRetry()
	sceResult, _ := se.executeScenario(exampleSpecWithTags.Scenarios[0])

	if !sceResult.GetFailed() {
		t.Errorf("Expect sceResult.GetFailed() = true, got false")
	}
}

func TestExecuteScenarioShouldRetryIfSpecificationMatchTags(t *testing.T) {
	MaxRetriesCount = 2
	RetryOnlyTags = "tagSpec"

	se := newSpecExecutorForTestsWithRetry()

	sceResult, _ := se.executeScenario(exampleSpecWithTags.Scenarios[0])

	if sceResult.GetFailed() {
		t.Errorf("Expect sceResult.GetFailed() = false, got true")
	}
}

func TestExecuteScenarioShouldRetryIfScenarioMatchTags(t *testing.T) {
	MaxRetriesCount = 2
	RetryOnlyTags = "tagSce"

	se := newSpecExecutorForTestsWithRetry()

	sceResult, _ := se.executeScenario(exampleSpecWithTags.Scenarios[0])

	if sceResult.GetFailed() {
		t.Errorf("Expect sceResult.GetFailed() = false, got true")
	}
}

func newSpecExecutorForTestsWithRetry() *specExecutor {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	se := newSpecExecutor(exampleSpecWithTags, r, nil, errs, 0)

	count := 1
	se.scenarioExecutor = &mockExecutor{
		executeFunc: func(i gauge.Item, r result.Result) {
			if count < MaxRetriesCount {
				r.SetFailure()
			} else {
				r.(*result.ScenarioResult).ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_PASSED
			}

			count++
		},
	}

	return se
}

func TestExecuteShouldMarkSpecAsSkippedWhenAllScenariosSkipped(t *testing.T) {
	errs := gauge.NewBuildErrors()
	r := &mockRunner{}
	se := newSpecExecutor(exampleSpecWithScenarios, r, nil, errs, 0)
	se.scenarioExecutor = &mockExecutor{
		executeFunc: func(i gauge.Item, r result.Result) {
			r.(*result.ScenarioResult).ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_SKIPPED
		},
	}
	res := se.execute(false, true, false)
	if !res.Skipped {
		t.Error("Expect SpecResult.Skipped = true, got false")
	}
}

func TestExecuteScenarioShoulHaveRetriesInfo(t *testing.T) {
	MaxRetriesCount = 3
	RetryOnlyTags = "tagSce"

	se := newSpecExecutorForTestsWithRetry()
	sceResult, _ := se.executeScenario(exampleSpecWithTags.Scenarios[0])

	if sceResult.GetFailed() {
		t.Errorf("Expect sceResult.GetFailed() = false, got true")
	}
	if se.currentExecutionInfo.CurrentScenario.Retries.MaxRetries != int32(MaxRetriesCount-1) {
		t.Errorf("Expected MaxRetries %d, got %d",
			int32(MaxRetriesCount-1),
			se.currentExecutionInfo.CurrentScenario.Retries.MaxRetries)
	}
	if se.currentExecutionInfo.CurrentScenario.Retries.CurrentRetry != int32(MaxRetriesCount-1) {
		t.Errorf("Expected CurrentRetry %d, got %d",
			int32(MaxRetriesCount-1),
			se.currentExecutionInfo.CurrentScenario.Retries.CurrentRetry)
	}
}
