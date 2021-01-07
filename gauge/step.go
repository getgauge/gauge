/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
)

type StepValue struct {
	Args                   []string
	StepValue              string
	ParameterizedStepValue string
}

type Step struct {
	LineNo         int
	FileName       string
	Value          string
	LineText       string
	Args           []*StepArg
	IsConcept      bool
	Lookup         ArgLookup
	ConceptSteps   []*Step
	Fragments      []*gauge_messages.Fragment
	Parent         *Step
	HasInlineTable bool
	Items          []Item
	PreComments    []*Comment
	Suffix         string
	LineSpanEnd    int
}

type StepDiff struct {
	OldStep   Step
	NewStep   *Step
	IsConcept bool
}

func (step *Step) GetArg(name string) (*StepArg, error) {
	arg, err := step.Lookup.GetArg(name)
	if err != nil {
		return nil, err
	}
	// Return static values
	if arg != nil && arg.ArgType != Dynamic {
		return arg, nil
	}
	if step.Parent == nil {
		return arg, nil
	}
	return step.Parent.GetArg(arg.Value)
}

func (step *Step) GetFragments() []*gauge_messages.Fragment {
	return step.Fragments
}

func (step *Step) GetLineText() string {
	if step.HasInlineTable {
		return fmt.Sprintf("%s <%s>", step.LineText, TableArg)
	}
	return step.LineText
}

func (step *Step) Rename(oldStep *Step, newStep *Step, isRefactored bool, orderMap map[int]int, isConcept *bool) (*StepDiff, bool) {
	diff := &StepDiff{OldStep: *step}
	if strings.TrimSpace(step.Value) != strings.TrimSpace(oldStep.Value) {
		return nil, isRefactored
	}
	if step.IsConcept {
		*isConcept = true
	}
	step.Value = newStep.Value
	diff.IsConcept = *isConcept
	step.Args = step.getArgsInOrder(newStep, orderMap)
	diff.NewStep = step
	return diff, true
}

func (step *Step) UsesDynamicArgs(args ...string) bool {
	for _, arg := range args {
		for _, stepArg := range step.Args {
			if (stepArg.Value == arg && stepArg.ArgType == Dynamic) || (stepArg.ArgType == TableArg && tableUsesDynamicArgs(stepArg, arg)) {
				return true
			}
		}
	}
	return false
}

func tableUsesDynamicArgs(tableArg *StepArg, arg string) bool {
	for _, cells := range tableArg.Table.Columns {
		for _, cell := range cells {
			if cell.CellType == Dynamic && cell.Value == arg {
				return true
			}
		}
	}
	return false
}

func (step *Step) getArgsInOrder(newStep *Step, orderMap map[int]int) []*StepArg {
	args := make([]*StepArg, len(newStep.Args))
	for key, value := range orderMap {
		arg := &StepArg{Value: newStep.Args[key].Value, ArgType: Static}
		if newStep.Args[key].ArgType == SpecialString || newStep.Args[key].ArgType == SpecialTable {
			arg = &StepArg{Name: newStep.Args[key].Name, Value: newStep.Args[key].Value, ArgType: newStep.Args[key].ArgType}
		}
		if step.IsConcept {
			name := fmt.Sprintf("arg%d", key)
			if newStep.Args[key].Value != "" && newStep.Args[key].ArgType != SpecialString {
				name = newStep.Args[key].Value
			}
			arg = &StepArg{Name: name, Value: newStep.Args[key].Value, ArgType: Dynamic}
		}
		if value != -1 {
			arg = step.Args[value]
		}
		args[key] = arg
	}
	return args
}

func (step *Step) ReplaceArgsWithDynamic(args []*StepArg) {
	for i, arg := range step.Args {
		for _, conceptArg := range args {
			if arg.String() == conceptArg.String() {
				if conceptArg.ArgType == SpecialString || conceptArg.ArgType == SpecialTable {
					reg := regexp.MustCompile(".*:")
					step.Args[i] = &StepArg{Name: reg.ReplaceAllString(conceptArg.Name, ""), ArgType: Dynamic}
					continue
				}
				if conceptArg.ArgType == Dynamic {
					step.Args[i] = &StepArg{Name: replaceParamChar(conceptArg.Name), ArgType: Dynamic}
					continue
				}
				step.Args[i] = &StepArg{Name: replaceParamChar(conceptArg.Value), ArgType: Dynamic}
			}
		}
	}
}

func (step *Step) AddArgs(args ...*StepArg) {
	step.Args = append(step.Args, args...)
	step.PopulateFragments()
}

func (step *Step) AddInlineTableHeaders(headers []string) {
	tableArg := &StepArg{ArgType: TableArg}
	tableArg.Table.AddHeaders(headers)
	step.AddArgs(tableArg)
}

func (step *Step) AddInlineTableRow(row []TableCell) {
	lastArg := step.Args[len(step.Args)-1]
	lastArg.Table.addRows(row)
	step.PopulateFragments()
}

func (step *Step) GetLastArg() *StepArg {
	return step.Args[len(step.Args)-1]
}

func (step *Step) PopulateFragments() {
	r := regexp.MustCompile(ParameterPlaceholder)
	/*
		enter {} and {} bar
		returns
		[[6 8] [13 15]]
	*/
	argSplitIndices := r.FindAllStringSubmatchIndex(step.Value, -1)
	step.Fragments = make([]*gauge_messages.Fragment, 0)
	if len(step.Args) == 0 {
		step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text, Text: step.Value})
		return
	}

	textStartIndex := 0
	for argIndex, argIndices := range argSplitIndices {
		if textStartIndex < argIndices[0] {
			step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text, Text: step.Value[textStartIndex:argIndices[0]]})
		}
		parameter := convertToProtoParameter(step.Args[argIndex])
		step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: parameter})
		textStartIndex = argIndices[1]
	}
	if textStartIndex < len(step.Value) {
		step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text, Text: step.Value[textStartIndex:len(step.Value)]})
	}

}

// InConcept returns true if the step belongs to a concept
func (step *Step) InConcept() bool {
	return step.Parent != nil
}

// Not copying parent as it enters an infinite loop in case of nested concepts. This is because the steps under the concept
// are copied and their parent copying again comes back to copy the same concept.
func (step *Step) GetCopy() (*Step, error) {
	if !step.IsConcept {
		return step, nil
	}
	nestedStepsCopy := make([]*Step, 0)
	for _, nestedStep := range step.ConceptSteps {
		nestedStepCopy, err := nestedStep.GetCopy()
		if err != nil {
			return nil, err
		}
		nestedStepsCopy = append(nestedStepsCopy, nestedStepCopy)
	}

	copiedConceptStep := new(Step)
	*copiedConceptStep = *step
	copiedConceptStep.ConceptSteps = nestedStepsCopy
	lookupCopy, err := step.Lookup.GetCopy()
	if err != nil {
		return nil, err
	}
	copiedConceptStep.Lookup = *lookupCopy
	return copiedConceptStep, nil
}

func (step *Step) CopyFrom(another *Step) {
	step.IsConcept = another.IsConcept

	if another.Args == nil {
		step.Args = nil
	} else {
		step.Args = make([]*StepArg, len(another.Args))
		copy(step.Args, another.Args)
	}

	if another.ConceptSteps == nil {
		step.ConceptSteps = nil
	} else {
		step.ConceptSteps = make([]*Step, len(another.ConceptSteps))
		copy(step.ConceptSteps, another.ConceptSteps)
	}

	if another.Fragments == nil {
		step.Fragments = nil
	} else {
		step.Fragments = make([]*gauge_messages.Fragment, len(another.Fragments))
		copy(step.Fragments, another.Fragments)
	}

	step.LineText = another.LineText
	step.HasInlineTable = another.HasInlineTable
	step.Value = another.Value
	step.Lookup = another.Lookup
	step.Parent = another.Parent
}

// skipcq CRT-P0003
func (step Step) Kind() TokenKind {
	return StepKind
}

func replaceParamChar(text string) string {
	return strings.Replace(strings.Replace(text, "<", "{", -1), ">", "}", -1)
}

func UsesArgs(steps []*Step, args ...string) bool {
	for _, s := range steps {
		if s.UsesDynamicArgs(args...) {
			return true
		}
	}
	return false
}
