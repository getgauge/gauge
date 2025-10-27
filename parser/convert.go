/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
)

func (parser *SpecParser) initializeConverters() []func(*Token, *int, *gauge.Specification) ParseResult {
	specConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.SpecKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if spec.Heading != nil {
			return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{spec.FileName, token.LineNo, token.SpanEnd, "Multiple spec headings found in same file", token.LineText()}}}
		}

		spec.AddHeading(&gauge.Heading{LineNo: token.LineNo, Value: token.Value, SpanEnd: token.SpanEnd})
		addStates(state, specScope)
		return ParseResult{Ok: true}
	})

	scenarioConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.ScenarioKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if spec.Heading == nil {
			return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{spec.FileName, token.LineNo, token.SpanEnd, "Scenario should be defined after the spec heading", token.LineText()}}}
		}
		for _, scenario := range spec.Scenarios {
			if strings.EqualFold(scenario.Heading.Value, token.Value) {
				return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{spec.FileName, token.LineNo, token.SpanEnd, "Duplicate scenario definition '" + scenario.Heading.Value + "' found in the same specification", token.LineText()}}}
			}
		}
		scenario := &gauge.Scenario{Span: &gauge.Span{Start: token.LineNo, End: token.LineNo}}
		if len(spec.Scenarios) > 0 {
			spec.LatestScenario().Span.End = token.LineNo - 1
		}
		scenario.AddHeading(&gauge.Heading{Value: token.Value, LineNo: token.LineNo, SpanEnd: token.SpanEnd})
		spec.AddScenario(scenario)

		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return ParseResult{Ok: true}
	})

	stepConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && isInState(*state, scenarioScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		latestScenario := spec.LatestScenario()
		stepToAdd, parseDetails := createStep(spec, latestScenario, token)
		if stepToAdd == nil {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		latestScenario.AddStep(stepToAdd)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	contextConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope) && !isInState(*state, tearDownScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		stepToAdd, parseDetails := createStep(spec, nil, token)
		if stepToAdd == nil {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.AddContext(stepToAdd)
		retainStates(state, specScope)
		addStates(state, contextScope)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			parseDetails.Ok = false
			return *parseDetails
		}
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	tearDownConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.TearDownKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		retainStates(state, specScope)
		addStates(state, tearDownScope)
		spec.AddItem(&gauge.TearDown{LineNo: token.LineNo, Value: token.Value})
		return ParseResult{Ok: true}
	})

	tearDownStepConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && isInState(*state, tearDownScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		stepToAdd, parseDetails := createStep(spec, nil, token)
		if stepToAdd == nil {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.TearDownSteps = append(spec.TearDownSteps, stepToAdd)
		spec.AddItem(stepToAdd)
		retainStates(state, specScope, tearDownScope)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			parseDetails.Ok = false
			return *parseDetails
		}
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	commentConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.CommentKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		comment := &gauge.Comment{Value: token.Value, LineNo: token.LineNo}
		if isInState(*state, scenarioScope) {
			spec.LatestScenario().AddComment(comment)
		} else {
			spec.AddComment(comment)
		}
		retainStates(state, specScope, scenarioScope, tearDownScope)
		addStates(state, commentScope)
		return ParseResult{Ok: true}
	})

	keywordConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.DataTableKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		resolvedArg, err := newSpecialTypeResolver().resolve(token.Value)
		if resolvedArg == nil || err != nil {
			errMessage := fmt.Sprintf("Could not resolve table. %s", err)
			gaugeDataDir := env.GaugeDataDir()
			if gaugeDataDir != "." {
				errMessage = fmt.Sprintf("%s GAUGE_DATA_DIR property is set to '%s', Gauge will look for data files in this location.", errMessage, gaugeDataDir)
			}
			e := ParseError{FileName: spec.FileName, LineNo: token.LineNo, LineText: token.LineText(), Message: errMessage}
			return ParseResult{ParseErrors: []ParseError{e}, Ok: false}
		}
		if isInAnyState(*state, scenarioScope) {
			scn := spec.LatestScenario()
			if !scn.DataTable.IsInitialized() {
				externalTable := &gauge.DataTable{}
				externalTable.Table = &resolvedArg.Table
				externalTable.LineNo = token.LineNo
				externalTable.Value = token.Value
				externalTable.IsExternal = true
				scn.AddExternalDataTable(externalTable)
			} else {
				value := "Multiple data table present, ignoring table"
				scn.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
				return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, token.SpanEnd, value}}}
			}
		} else if isInState(*state, specScope) && !spec.DataTable.IsInitialized() {
			externalTable := &gauge.DataTable{}
			externalTable.Table = &resolvedArg.Table
			externalTable.LineNo = token.LineNo
			externalTable.Value = token.Value
			externalTable.IsExternal = true
			spec.AddExternalDataTable(externalTable)
		} else if isInState(*state, specScope) && spec.DataTable.IsInitialized() {
			value := "Multiple data table present, ignoring table"
			spec.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, token.SpanEnd, value}}}
		} else {
			value := "Data table not associated with spec or scenario"
			spec.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, token.SpanEnd, value}}}
		}
		retainStates(state, specScope, scenarioScope)
		addStates(state, keywordScope)
		return ParseResult{Ok: true}
	})

	tableHeaderConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.TableHeader && isInAnyState(*state, specScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.LatestScenario()
			latestStep := latestScenario.LatestStep()
			addInlineTableHeader(latestStep, token)
		} else if isInState(*state, contextScope) {
			latestContext := spec.LatestContext()
			addInlineTableHeader(latestContext, token)
		} else if isInState(*state, tearDownScope) {
			if len(spec.TearDownSteps) > 0 {
				latestTeardown := spec.LatestTeardown()
				addInlineTableHeader(latestTeardown, token)
			} else {
				spec.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
			}
		} else if isInState(*state, scenarioScope) {
			scn := spec.LatestScenario()
			if !scn.DataTable.Table.IsInitialized() {
				dataTable := &gauge.Table{LineNo: token.LineNo}
				dataTable.AddHeaders(token.Args)
				scn.AddDataTable(dataTable)
			} else {
				scn.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
				return ParseResult{Ok: false, Warnings: []*Warning{
					&Warning{spec.FileName, token.LineNo, token.SpanEnd, "Multiple data table present, ignoring table"}}}
			}
		} else {
			if !spec.DataTable.Table.IsInitialized() {
				dataTable := &gauge.Table{LineNo: token.LineNo}
				dataTable.AddHeaders(token.Args)
				spec.AddDataTable(dataTable)
			} else {
				spec.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
				return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName,
					token.LineNo, token.SpanEnd, "Multiple data table present, ignoring table"}}}
			}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope)
		addStates(state, tableScope)
		return ParseResult{Ok: true}
	})

	tableRowConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.TableRow
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		var result ParseResult
		//When table is to be treated as a comment
		if !isInState(*state, tableScope) {
			if isInState(*state, scenarioScope) {
				spec.LatestScenario().AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
			} else {
				spec.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
			}
		} else if areUnderlined(token.Args) && !isInState(*state, tableSeparatorScope) {
			retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope, tableScope)
			addStates(state, tableSeparatorScope)
			// skip table separator
			result = ParseResult{Ok: true}
		} else if isInState(*state, stepScope) {
			latestScenario := spec.LatestScenario()
			tables := []*gauge.Table{spec.DataTable.Table}
			if latestScenario.DataTable.IsInitialized() {
				tables = append(tables, latestScenario.DataTable.Table)
			}
			latestStep := latestScenario.LatestStep()
			result = addInlineTableRow(latestStep, token, new(gauge.ArgLookup).FromDataTables(tables...), spec.FileName)
		} else if isInState(*state, contextScope) {
			latestContext := spec.LatestContext()
			result = addInlineTableRow(latestContext, token, new(gauge.ArgLookup).FromDataTables(spec.DataTable.Table), spec.FileName)
		} else if isInState(*state, tearDownScope) {
			if len(spec.TearDownSteps) > 0 {
				latestTeardown := spec.LatestTeardown()
				result = addInlineTableRow(latestTeardown, token, new(gauge.ArgLookup).FromDataTables(spec.DataTable.Table), spec.FileName)
			} else {
				spec.AddComment(&gauge.Comment{Value: token.LineText(), LineNo: token.LineNo})
			}
		} else {
			t := spec.DataTable
			if isInState(*state, scenarioScope) {
				t = spec.LatestScenario().DataTable
			}

			tableValues, warnings, err := validateTableRows(token, new(gauge.ArgLookup).FromDataTables(t.Table), spec.FileName)
			if len(err) > 0 {
				result = ParseResult{Ok: false, Warnings: warnings, ParseErrors: err}
			} else {
				t.Table.AddRowValues(tableValues)
				result = ParseResult{Ok: true, Warnings: warnings}
			}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope, tableScope, tableSeparatorScope)
		return result
	})

	tagConverter := converterFn(func(token *Token, state *int) bool {
		return (token.Kind == gauge.TagKind)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		tags := &gauge.Tags{RawValues: [][]string{token.Args}}
		if isInState(*state, scenarioScope) {
			if isInState(*state, tagsScope) {
				spec.LatestScenario().Tags.Add(tags.RawValues[0])
			} else {
				if spec.LatestScenario().NTags() != 0 {
					return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{FileName: spec.FileName, LineNo: token.LineNo, Message: "Tags can be defined only once per scenario", LineText: token.LineText()}}}
				}
				spec.LatestScenario().AddTags(tags)
			}
		} else {
			if isInState(*state, tagsScope) {
				spec.Tags.Add(tags.RawValues[0])
			} else {
				if spec.NTags() != 0 {
					return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{FileName: spec.FileName, LineNo: token.LineNo, Message: "Tags can be defined only once per specification", LineText: token.LineText()}}}
				}
				spec.AddTags(tags)
			}
		}
		addStates(state, tagsScope)
		return ParseResult{Ok: true}
	})

	converter := []func(*Token, *int, *gauge.Specification) ParseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter, keywordConverter, tearDownConverter, tearDownStepConverter,
	}

	return converter
}

func converterFn(predicate func(token *Token, state *int) bool, apply func(token *Token, spec *gauge.Specification, state *int) ParseResult) func(*Token, *int, *gauge.Specification) ParseResult {
	return func(token *Token, state *int, spec *gauge.Specification) ParseResult {
		if !predicate(token, state) {
			return ParseResult{Ok: true}
		}
		return apply(token, spec, state)
	}
}

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *gauge.Step, token *Token) {
	step.Value = fmt.Sprintf("%s %s", step.Value, gauge.ParameterPlaceholder)
	step.HasInlineTable = true
	step.AddInlineTableHeaders(token.Args)
}

func addInlineTableRow(step *gauge.Step, token *Token, argLookup *gauge.ArgLookup, fileName string) ParseResult {
	tableValues, warnings, err := validateTableRows(token, argLookup, fileName)
	if len(err) > 0 {
		return ParseResult{Ok: false, Warnings: warnings, ParseErrors: err}
	}
	step.AddInlineTableRow(tableValues)
	return ParseResult{Ok: true, Warnings: warnings}
}

func validateTableRows(token *Token, argLookup *gauge.ArgLookup, fileName string) ([]gauge.TableCell, []*Warning, []ParseError) {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	specialArgMatcher := regexp.MustCompile("^<(file:.*)>$")
	tableValues := make([]gauge.TableCell, 0)
	warnings := make([]*Warning, 0)
	error := make([]ParseError, 0)
	for _, tableValue := range token.Args {
		if specialArgMatcher.MatchString(tableValue) {
			match := specialArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			file := strings.TrimSpace(strings.TrimPrefix(param, "file:"))
			tableValues = append(tableValues, gauge.TableCell{Value: param, CellType: gauge.SpecialString})
			if _, err := util.GetFileContents(file); err != nil {
				error = append(error, ParseError{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Missing file: %s", param, file), LineText: token.LineText()})
			}
		} else if dynamicArgMatcher.MatchString(tableValue) {
			match := dynamicArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			if !argLookup.ContainsArg(param) {
				tableValues = append(tableValues, gauge.TableCell{Value: tableValue, CellType: gauge.Static})
				warnings = append(warnings, &Warning{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Treating it as static param", param)})
			} else {
				tableValues = append(tableValues, gauge.TableCell{Value: param, CellType: gauge.Dynamic})
			}
		} else {
			tableValues = append(tableValues, gauge.TableCell{Value: tableValue, CellType: gauge.Static})
		}
	}
	return tableValues, warnings, error
}
