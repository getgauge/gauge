package main

import (
	. "launchpad.net/gocheck"
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
	agent.refactor(&specs)

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
	agent.refactor(&specs)

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
	agent.refactor(&specs)

	c.Assert(err, Equals, nil)
	c.Assert(len(specs[0].scenarios[0].steps), Equals, 1)
	c.Assert(specs[0].scenarios[0].steps[0].value, Equals, newStep)

	c.Assert(len(specs[1].scenarios[0].steps), Equals, 1)
	c.Assert(specs[1].scenarios[0].steps[0].value, Equals, newStep)

}
