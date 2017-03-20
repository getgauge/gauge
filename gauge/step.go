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

package gauge

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getgauge/gauge/gauge_messages"
)

type StepValue struct {
	Args                   []string
	StepValue              string
	ParameterizedStepValue string
}

type Step struct {
	LineNo         int
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
}

func (step *Step) GetArg(name string) *StepArg {
	arg := step.Lookup.GetArg(name)
	// Return static values
	if arg != nil && arg.ArgType != Dynamic {
		return arg
	}
	if step.Parent == nil {
		return step.Lookup.GetArg(name)
	}
	return step.Parent.GetArg(step.Lookup.GetArg(name).Value)
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

func (step *Step) Rename(oldStep Step, newStep Step, isRefactored bool, orderMap map[int]int, isConcept *bool) bool {
	if strings.TrimSpace(step.Value) != strings.TrimSpace(oldStep.Value) {
		return isRefactored
	}
	if step.IsConcept {
		*isConcept = true
	}
	step.Value = newStep.Value

	step.Args = step.getArgsInOrder(newStep, orderMap)
	return true
}

func (step *Step) UsesDynamicArgs(args ...string) bool{
	return step.containsAtLeastOneDynamicArgsInStep(args...)

}

func (step *Step)containsAtLeastOneDynamicArgsInStep(args ...string)bool{
	for _, arg := range args {
		for _, stepArg := range step.Args {
			if stepArg.Value == arg && stepArg.ArgType == Dynamic {
				return true
			}
		}
	}
	return false
}

func (step *Step) getArgsInOrder(newStep Step, orderMap map[int]int) []*StepArg {
	args := make([]*StepArg, len(newStep.Args))
	for key, value := range orderMap {
		arg := &StepArg{Value: newStep.Args[key].Value, ArgType: Static}
		if step.IsConcept {
			arg = &StepArg{Value: newStep.Args[key].Value, ArgType: Dynamic}
		}
		if value != -1 {
			arg = step.Args[value]
		}
		args[key] = arg
	}
	return args
}

func (step *Step) deepCopyStepArgs() []*StepArg {
	copiedStepArgs := make([]*StepArg, 0)
	for _, conceptStepArg := range step.Args {
		temp := new(StepArg)
		*temp = *conceptStepArg
		copiedStepArgs = append(copiedStepArgs, temp)
	}
	return copiedStepArgs
}

func (step *Step) ReplaceArgsWithDynamic(args []*StepArg) {
	for i, arg := range step.Args {
		for _, conceptArg := range args {
			if arg.String() == conceptArg.String() {
				if conceptArg.ArgType == SpecialString || conceptArg.ArgType == SpecialTable {
					reg := regexp.MustCompile(".*:")
					step.Args[i] = &StepArg{Value: reg.ReplaceAllString(conceptArg.Name, ""), ArgType: Dynamic}
					continue
				}
				step.Args[i] = &StepArg{Value: replaceParamChar(conceptArg.Value), ArgType: Dynamic}
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
func (self *Step) GetCopy() *Step {
	if !self.IsConcept {
		return self
	}
	nestedStepsCopy := make([]*Step, 0)
	for _, nestedStep := range self.ConceptSteps {
		nestedStepsCopy = append(nestedStepsCopy, nestedStep.GetCopy())
	}

	copiedConceptStep := new(Step)
	*copiedConceptStep = *self
	copiedConceptStep.ConceptSteps = nestedStepsCopy
	copiedConceptStep.Lookup = *self.Lookup.GetCopy()
	return copiedConceptStep
}

func (self *Step) CopyFrom(another *Step) {
	self.IsConcept = another.IsConcept

	if another.Args == nil {
		self.Args = nil
	} else {
		self.Args = make([]*StepArg, len(another.Args))
		copy(self.Args, another.Args)
	}

	if another.ConceptSteps == nil {
		self.ConceptSteps = nil
	} else {
		self.ConceptSteps = make([]*Step, len(another.ConceptSteps))
		copy(self.ConceptSteps, another.ConceptSteps)
	}

	if another.Fragments == nil {
		self.Fragments = nil
	} else {
		self.Fragments = make([]*gauge_messages.Fragment, len(another.Fragments))
		copy(self.Fragments, another.Fragments)
	}

	self.LineNo = another.LineNo
	self.LineText = another.LineText
	self.HasInlineTable = another.HasInlineTable
	self.Value = another.Value
	self.Lookup = another.Lookup
	self.Parent = another.Parent
}

func (step Step) Kind() TokenKind {
	return StepKind
}

func replaceParamChar(text string) string {
	return strings.Replace(strings.Replace(text, "<", "{", -1), ">", "}", -1)
}
