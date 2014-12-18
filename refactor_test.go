package main

import (
	. "gopkg.in/check.v1"
	"reflect"
)

func (s *MySuite) TestGetRefactoringAgentGivesRenameRefactorerWhenThereIsNoParametersInStep(c *C) {
	oldStep := "first step"
	newStep := "second step"
	agent, err := getRefactorAgent(oldStep, newStep)

	c.Assert(err, Equals, nil)
	c.Assert(reflect.TypeOf(agent).Elem().Name(), Equals, "renameRefactorer")
}

func (s *MySuite) TestGetRefactoringAgentGivesNilWhenThereIsNoRefactorerPresentToHandleAParticularRefactoring(c *C) {
	oldStep := "first step \" a \" "
	newStep := "second step"
	agent, err := getRefactorAgent(oldStep, newStep)

	c.Assert(err.Error(), Equals, ERROR_MESSAGE)
	c.Assert(agent, Equals, nil)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgs(c *C) {
	oldStep := "first step"
	newStep := "second step"
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: oldStep, lineNo: 3},
	}
	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	agent, err := getRefactorAgent(oldStep, newStep)
	specs := append(make([]*specification, 0), spec)
	agent.refactor(&specs, new(conceptDictionary))

	c.Assert(err, Equals, nil)
	c.Assert(len(specs[0].scenarios[0].steps), Equals, 1)
	c.Assert(specs[0].scenarios[0].steps[0].value, Equals, newStep)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgsAndWithMoreThanOneScenario(c *C) {
	oldStep := "first step"
	newStep := "second step"
	unchanged := "unchanged"
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: oldStep, lineNo: 3},
		&token{kind: stepKind, value: oldStep, lineNo: 5},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 20},
		&token{kind: stepKind, value: unchanged, lineNo: 30},
		&token{kind: stepKind, value: oldStep, lineNo: 50},
	}
	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	agent, err := getRefactorAgent(oldStep, newStep)
	specs := append(make([]*specification, 0), spec)
	agent.refactor(&specs, new(conceptDictionary))

	c.Assert(err, Equals, nil)
	c.Assert(len(specs[0].scenarios), Equals, 2)
	c.Assert(len(specs[0].scenarios[0].steps), Equals, 2)
	c.Assert(specs[0].scenarios[0].steps[0].value, Equals, newStep)
	c.Assert(specs[0].scenarios[0].steps[1].value, Equals, newStep)

	c.Assert(len(specs[0].scenarios[1].steps), Equals, 2)
	c.Assert(specs[0].scenarios[1].steps[0].value, Equals, unchanged)
	c.Assert(specs[0].scenarios[1].steps[1].value, Equals, newStep)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgsAndWithMoreThanOneSpec(c *C) {
	oldStep := " first step"
	newStep := "second step"
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: oldStep, lineNo: 3},
	}
	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	tokens = []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 10},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 20},
		&token{kind: stepKind, value: oldStep, lineNo: 30},
	}
	spec1, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	specs := append(make([]*specification, 0), spec)
	specs = append(specs, spec1)
	agent, err := getRefactorAgent(oldStep, newStep)
	specRefactored, _ := agent.refactor(&specs, new(conceptDictionary))

	for _, isRefactored := range specRefactored {
		c.Assert(true, Equals, isRefactored)
	}
	c.Assert(err, Equals, nil)
	c.Assert(len(specs[0].scenarios[0].steps), Equals, 1)
	c.Assert(specs[0].scenarios[0].steps[0].value, Equals, newStep)

	c.Assert(len(specs[1].scenarios[0].steps), Equals, 1)
	c.Assert(specs[1].scenarios[0].steps[0].value, Equals, newStep)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgsInConceptFiles(c *C) {
	oldStep := "first step"
	newStep := "second step"
	unchanged := "unchanged"
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 20},
	}
	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	agent, _ := getRefactorAgent(oldStep, newStep)
	specs := append(make([]*specification, 0), spec)
	dictionary := new(conceptDictionary)
	step1 := &step{value: oldStep + "sdsf", isConcept: true}
	step2 := &step{value: unchanged, isConcept: true, items: []item{&step{value: oldStep, isConcept: false}, &step{value: oldStep + "T", isConcept: false}}}
	dictionary.add([]*step{step1, step2}, "file.cpt")

	agent.refactor(&specs, dictionary)

	c.Assert(dictionary.conceptsMap[unchanged].conceptStep.items[0].(*step).value, Equals, newStep)
	c.Assert(dictionary.conceptsMap[unchanged].conceptStep.items[1].(*step).value, Equals, oldStep+"T")
}

func (s *MySuite) TestRefactoringGivesOnlySpecsThatAreRefactored(c *C) {
	oldStep := " first step"
	newStep := "second step"
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: oldStep, lineNo: 3},
	}
	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	tokens = []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 10},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 20},
		&token{kind: stepKind, value: newStep, lineNo: 30},
	}
	spec1, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	specs := append(make([]*specification, 0), spec)
	specs = append(specs, spec1)
	agent, _ := getRefactorAgent(oldStep, newStep)
	specRefactored, _ := agent.refactor(&specs, new(conceptDictionary))

	c.Assert(true, Equals, specRefactored[specs[0]])
	c.Assert(false, Equals, specRefactored[specs[1]])
}

func (s *MySuite) TestRefactoringGivesOnlyThoseConceptFilesWhichAreRefactored(c *C) {
	oldStep := "first step"
	newStep := "second step"
	unchanged := "unchanged"
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 20},
	}
	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	agent, _ := getRefactorAgent(oldStep, newStep)
	specs := append(make([]*specification, 0), spec)
	dictionary := new(conceptDictionary)
	step1 := &step{value: oldStep + "sdsf", isConcept: true}
	step2 := &step{value: unchanged, isConcept: true, items: []item{&step{value: newStep, isConcept: false}, &step{value: oldStep + "T", isConcept: false}}}
	step3 := &step{value: "Concept value", isConcept: true, items: []item{&step{value: oldStep, isConcept: false}, &step{value: oldStep + "T", isConcept: false}}}
	fileName := "file.cpt"
	dictionary.add([]*step{step1, step2}, fileName)
	dictionary.add([]*step{step3}, "e"+fileName)

	_, filesRefactored := agent.refactor(&specs, dictionary)

	c.Assert(filesRefactored[fileName], Equals, false)
	c.Assert(filesRefactored["e"+fileName], Equals, true)
}
