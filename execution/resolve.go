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

package execution

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
)

type setSkipInfoFn func(protoStep *gauge_messages.ProtoStep, step *gauge.Step)
type argLookupFn func() (*gauge.ArgLookup, error)

func resolveItems(items []gauge.Item, lookupFn argLookupFn, skipFn setSkipInfoFn) ([]*gauge_messages.ProtoItem, error) {
	var protoItems []*gauge_messages.ProtoItem
	for _, item := range items {
		if item.Kind() != gauge.TearDownKind {
			protoItem, err := resolveToProtoItem(item, lookupFn, skipFn)
			if err != nil {
				return nil, err
			}
			protoItems = append(protoItems, protoItem)
		}
	}
	return protoItems, nil
}

func resolveToProtoItem(item gauge.Item, lookupFn argLookupFn, skipFn setSkipInfoFn) (*gauge_messages.ProtoItem, error) {
	var protoItem *gauge_messages.ProtoItem
	var err error
	switch item.Kind() {
	case gauge.StepKind:
		if (item.(*gauge.Step)).IsConcept {
			concept := item.(*gauge.Step)
			// lookup, err := e.dataTableLookup()
			protoItem, err = resolveToProtoConceptItem(*concept, lookupFn, skipFn)
		} else {
			protoItem, err = resolveToProtoStepItem(item.(*gauge.Step), lookupFn, skipFn)
		}
		break

	default:
		protoItem = gauge.ConvertToProtoItem(item)
	}
	return protoItem, err
}

// Not passing pointer as we cannot modify the original concept step's lookup. This has to be populated for each iteration over data table.
func resolveToProtoConceptItem(concept gauge.Step, lookupFn argLookupFn, skipFn setSkipInfoFn) (*gauge_messages.ProtoItem, error) {
	lookup, err := lookupFn()
	if err != nil {
		return nil, err
	}
	if err := parser.PopulateConceptDynamicParams(&concept, lookup); err != nil {
		return nil, err
	}
	protoConceptItem := gauge.ConvertToProtoItem(&concept)
	protoConceptItem.Concept.ConceptStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}
	for stepIndex, step := range concept.ConceptSteps {
		// Need to reset parent as the step.parent is pointing to a concept whose lookup is not populated yet
		if step.IsConcept {
			step.Parent = &concept
			protoItem, err := resolveToProtoConceptItem(*step, func() (*gauge.ArgLookup, error) { return &concept.Lookup, nil }, skipFn)
			if err != nil {
				return nil, err
			}
			protoConceptItem.GetConcept().GetSteps()[stepIndex] = protoItem
		} else {
			conceptStep := protoConceptItem.Concept.Steps[stepIndex].Step
			err := parser.Resolve(step, &concept, &concept.Lookup, conceptStep)
			if err != nil {
				return nil, err
			}
			// e.setSkipInfo(conceptStep, step)
			skipFn(conceptStep, step)
		}
	}
	protoConceptItem.Concept.ConceptStep.StepExecutionResult.Skipped = false
	return protoConceptItem, nil
}

func resolveToProtoStepItem(step *gauge.Step, lookupFn argLookupFn, skipFn setSkipInfoFn) (*gauge_messages.ProtoItem, error) {
	protoStepItem := gauge.ConvertToProtoItem(step)
	lookup, err := lookupFn()
	if err != nil {
		return nil, err
	}
	err = parser.Resolve(step, nil, lookup, protoStepItem.Step)
	if err != nil {
		return nil, err
	}
	// e.setSkipInfo(protoStepItem.Step, step)
	skipFn(protoStepItem.Step, step)
	return protoStepItem, err
}
