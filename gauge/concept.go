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

type ConceptDictionary struct {
	ConceptsMap     map[string]*Concept
	constructionMap map[string][]*Step
}

type Concept struct {
	ConceptStep *Step
	FileName    string
}

func NewConceptDictionary() *ConceptDictionary {
	return &ConceptDictionary{ConceptsMap: make(map[string]*Concept, 0), constructionMap: make(map[string][]*Step, 0)}
}

func (dict *ConceptDictionary) Search(stepValue string) *Concept {
	if concept, ok := dict.ConceptsMap[stepValue]; ok {
		return concept
	}
	return nil
}

func (dict *ConceptDictionary) ReplaceNestedConceptSteps(conceptStep *Step) {
	dict.updateStep(conceptStep)
	for i, stepInsideConcept := range conceptStep.ConceptSteps {
		if nestedConcept := dict.Search(stepInsideConcept.Value); nestedConcept != nil {
			//replace step with actual concept
			conceptStep.ConceptSteps[i].ConceptSteps = nestedConcept.ConceptStep.ConceptSteps
			conceptStep.ConceptSteps[i].IsConcept = nestedConcept.ConceptStep.IsConcept
			conceptStep.ConceptSteps[i].Lookup = *nestedConcept.ConceptStep.Lookup.GetCopy()
		} else {
			dict.updateStep(stepInsideConcept)
		}
	}
}

//mutates the step with concept steps so that anyone who is referencing the step will now refer a concept
func (dict *ConceptDictionary) updateStep(step *Step) {
	dict.constructionMap[step.Value] = append(dict.constructionMap[step.Value], step)
	if !dict.constructionMap[step.Value][0].IsConcept {
		dict.constructionMap[step.Value] = append(dict.constructionMap[step.Value], step)
		for _, allSteps := range dict.constructionMap[step.Value] {
			allSteps.IsConcept = step.IsConcept
			allSteps.ConceptSteps = step.ConceptSteps
			allSteps.Lookup = *step.Lookup.GetCopy()
		}
	}
}

func (dict *ConceptDictionary) UpdateLookupForNestedConcepts() {
	for _, concept := range dict.ConceptsMap {
		for _, stepInsideConcept := range concept.ConceptStep.ConceptSteps {
			stepInsideConcept.Parent = concept.ConceptStep
			if nestedConcept := dict.Search(stepInsideConcept.Value); nestedConcept != nil {
				for i, arg := range nestedConcept.ConceptStep.Args {
					stepInsideConcept.Lookup.AddArgValue(arg.Value, &StepArg{ArgType: stepInsideConcept.Args[i].ArgType, Value: stepInsideConcept.Args[i].Value})
				}
			}
		}
	}
}

type ByLineNo []*Concept

func (s ByLineNo) Len() int {
	return len(s)
}

func (s ByLineNo) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLineNo) Less(i, j int) bool {
	return s[i].ConceptStep.LineNo < s[j].ConceptStep.LineNo
}
