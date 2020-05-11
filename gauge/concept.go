/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	return &ConceptDictionary{ConceptsMap: make(map[string]*Concept), constructionMap: make(map[string][]*Step)}
}

func (dict *ConceptDictionary) Search(stepValue string) *Concept {
	if concept, ok := dict.ConceptsMap[stepValue]; ok {
		return concept
	}
	return nil
}

func (dict *ConceptDictionary) ReplaceNestedConceptSteps(conceptStep *Step) error {
	if err := dict.updateStep(conceptStep); err != nil {
		return err
	}
	for i, stepInsideConcept := range conceptStep.ConceptSteps {
		if nestedConcept := dict.Search(stepInsideConcept.Value); nestedConcept != nil {
			//replace step with actual concept
			conceptStep.ConceptSteps[i].ConceptSteps = nestedConcept.ConceptStep.ConceptSteps
			conceptStep.ConceptSteps[i].IsConcept = nestedConcept.ConceptStep.IsConcept
			lookupCopy, err := nestedConcept.ConceptStep.Lookup.GetCopy()
			if err != nil {
				return err
			}
			conceptStep.ConceptSteps[i].Lookup = *lookupCopy
		} else {
			if err := dict.updateStep(stepInsideConcept); err != nil {
				return err
			}

		}
	}
	return nil
}

//mutates the step with concept steps so that anyone who is referencing the step will now refer a concept
func (dict *ConceptDictionary) updateStep(step *Step) error {
	dict.constructionMap[step.Value] = append(dict.constructionMap[step.Value], step)
	if !dict.constructionMap[step.Value][0].IsConcept {
		dict.constructionMap[step.Value] = append(dict.constructionMap[step.Value], step)
		for _, allSteps := range dict.constructionMap[step.Value] {
			allSteps.IsConcept = step.IsConcept
			allSteps.ConceptSteps = step.ConceptSteps
			lookupCopy, err := step.Lookup.GetCopy()
			if err != nil {
				return err
			}
			allSteps.Lookup = *lookupCopy
		}
	}
	return nil
}

func (dict *ConceptDictionary) UpdateLookupForNestedConcepts() error {
	for _, concept := range dict.ConceptsMap {
		for _, stepInsideConcept := range concept.ConceptStep.ConceptSteps {
			stepInsideConcept.Parent = concept.ConceptStep
			if nestedConcept := dict.Search(stepInsideConcept.Value); nestedConcept != nil {
				for i, arg := range nestedConcept.ConceptStep.Args {
					stepArg := StepArg{ArgType: stepInsideConcept.Args[i].ArgType, Value: stepInsideConcept.Args[i].Value, Table: stepInsideConcept.Args[i].Table}
					if err := stepInsideConcept.Lookup.AddArgValue(arg.Value, &stepArg); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (dict *ConceptDictionary) Remove(stepValue string) {
	delete(dict.ConceptsMap, stepValue)
	delete(dict.constructionMap, stepValue)
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
