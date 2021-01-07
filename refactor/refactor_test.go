/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package refactor

import (
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestRefactoringOfStepsWithNoArgs(c *C) {
	oldStep := "first step"
	newStep := "second step"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, errs := getRefactorAgent(oldStep, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	agent.rephraseInSpecsAndConcepts(&specs, gauge.NewConceptDictionary())

	c.Assert(len(errs), Equals, 0)
	c.Assert(len(specs[0].Scenarios[0].Steps), Equals, 1)
	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, newStep)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgsAndWithMoreThanOneScenario(c *C) {
	oldStep := "first step"
	newStep := "second step"
	unchanged := "unchanged"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 5},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 20},
		&parser.Token{Kind: gauge.StepKind, Value: unchanged, LineNo: 30},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 50},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, errs := getRefactorAgent(oldStep, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	agent.rephraseInSpecsAndConcepts(&specs, gauge.NewConceptDictionary())

	c.Assert(len(errs), Equals, 0)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(len(specs[0].Scenarios[0].Steps), Equals, 2)
	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, newStep)
	c.Assert(specs[0].Scenarios[0].Steps[1].Value, Equals, newStep)

	c.Assert(len(specs[0].Scenarios[1].Steps), Equals, 2)
	c.Assert(specs[0].Scenarios[1].Steps[0].Value, Equals, unchanged)
	c.Assert(specs[0].Scenarios[1].Steps[1].Value, Equals, newStep)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgsAndWithMoreThanOneSpec(c *C) {
	oldStep := " first step"
	newStep := "second step"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	tokens = []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 10},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 20},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 30},
	}
	spec1, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	specs := append(make([]*gauge.Specification, 0), spec)
	specs = append(specs, spec1)
	agent, errs := getRefactorAgent(oldStep, newStep, nil)
	specRefactored, _ := agent.rephraseInSpecsAndConcepts(&specs, gauge.NewConceptDictionary())

	for _, diffs := range specRefactored {
		c.Assert(1, Equals, len(diffs))
	}
	c.Assert(len(errs), Equals, 0)
	c.Assert(len(specs[0].Scenarios[0].Steps), Equals, 1)
	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, newStep)

	c.Assert(len(specs[1].Scenarios[0].Steps), Equals, 1)
	c.Assert(specs[1].Scenarios[0].Steps[0].Value, Equals, newStep)
}

func (s *MySuite) TestRefactoringOfStepsWithNoArgsInConceptFiles(c *C) {
	oldStep := "first step"
	newStep := "second step"
	unchanged := "unchanged"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 20},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	step1 := &gauge.Step{Value: oldStep + "sdsf", IsConcept: true}
	step2 := &gauge.Step{Value: unchanged, IsConcept: true, Items: []gauge.Item{&gauge.Step{Value: oldStep, IsConcept: false}, &gauge.Step{Value: oldStep + "T", IsConcept: false}}}

	dictionary.ConceptsMap[step1.Value] = &gauge.Concept{ConceptStep: step1, FileName: "file.cpt"}
	dictionary.ConceptsMap[step2.Value] = &gauge.Concept{ConceptStep: step2, FileName: "file.cpt"}

	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(dictionary.ConceptsMap[unchanged].ConceptStep.Items[0].(*gauge.Step).Value, Equals, newStep)
	c.Assert(dictionary.ConceptsMap[unchanged].ConceptStep.Items[1].(*gauge.Step).Value, Equals, oldStep+"T")
}

func (s *MySuite) TestRefactoringGivesOnlySpecsThatAreRefactored(c *C) {
	oldStep := " first step"
	newStep := "second step"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	tokens = []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 10},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 20},
		&parser.Token{Kind: gauge.StepKind, Value: newStep, LineNo: 30},
	}
	spec1, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	specs := append(make([]*gauge.Specification, 0), spec)
	specs = append(specs, spec1)
	agent, _ := getRefactorAgent(oldStep, newStep, nil)
	specRefactored, _ := agent.rephraseInSpecsAndConcepts(&specs, gauge.NewConceptDictionary())

	c.Assert(1, Equals, len(specRefactored[specs[0]]))
	c.Assert(0, Equals, len(specRefactored[specs[1]]))
	c.Assert(specRefactored[specs[0]][0].OldStep.Value, Equals, " first step")
	c.Assert(specRefactored[specs[0]][0].NewStep.Value, Equals, "second step")
}

func (s *MySuite) TestRefactoringGivesOnlyThoseConceptFilesWhichAreRefactored(c *C) {
	oldStep := "first step"
	newStep := "second step"
	unchanged := "unchanged"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 20},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	step1 := &gauge.Step{Value: oldStep + "sdsf", IsConcept: true}
	step2 := &gauge.Step{Value: unchanged, IsConcept: true, Items: []gauge.Item{&gauge.Step{Value: newStep, IsConcept: false}, &gauge.Step{Value: oldStep + "T", IsConcept: false}}}
	step3 := &gauge.Step{Value: "Concept value", IsConcept: true, Items: []gauge.Item{&gauge.Step{Value: oldStep, IsConcept: false}, &gauge.Step{Value: oldStep + "T", IsConcept: false}}}
	fileName := "file.cpt"
	dictionary.ConceptsMap[step1.Value] = &gauge.Concept{ConceptStep: step1, FileName: fileName}
	dictionary.ConceptsMap[step2.Value] = &gauge.Concept{ConceptStep: step2, FileName: fileName}
	dictionary.ConceptsMap[step3.Value] = &gauge.Concept{ConceptStep: step3, FileName: "e" + fileName}

	_, filesRefactored := agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(len(filesRefactored[fileName]), Equals, 0)
	c.Assert(len(filesRefactored["e"+fileName]), Equals, 1)
}

func (s *MySuite) TestRenamingWhenNumberOfArgumentsAreSame(c *C) {
	oldStep := "first step {static} and {static}"
	oldStep1 := "first step <a> and <b>"
	newStep := "second step <a> and <b>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)
	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "name")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "address")
}

func (s *MySuite) TestRenamingWhenArgumentsOrderIsChanged(c *C) {
	oldStep := "first step {static} and {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c> and <d>"
	newStep := "second step <d> and <b> and <c> and <a>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number", "id"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)
	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "id")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "number")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[3].Value, Equals, "name")
}

func (s *MySuite) TestCreateOrderGivesMapOfOldArgsAndNewArgs(c *C) {
	step1 := &gauge.Step{Args: []*gauge.StepArg{&gauge.StepArg{Name: "a"}, &gauge.StepArg{Name: "b"}, &gauge.StepArg{Name: "c"}, &gauge.StepArg{Name: "d"}}}
	step2 := &gauge.Step{Args: []*gauge.StepArg{&gauge.StepArg{Name: "d"}, &gauge.StepArg{Name: "b"}, &gauge.StepArg{Name: "c"}, &gauge.StepArg{Name: "a"}}}

	agent := &rephraseRefactorer{step1, step2, false, nil}
	orderMap := agent.createOrderOfArgs()

	c.Assert(orderMap[0], Equals, 3)
	c.Assert(orderMap[1], Equals, 1)
	c.Assert(orderMap[2], Equals, 2)
}

func (s *MySuite) TestCreateOrderGivesMapOfOldArgsAndNewWhenArgsAreAdded(c *C) {
	step1 := &gauge.Step{Args: []*gauge.StepArg{&gauge.StepArg{Name: "a"}, &gauge.StepArg{Name: "b"}, &gauge.StepArg{Name: "c"}, &gauge.StepArg{Name: "d"}}}
	step2 := &gauge.Step{Args: []*gauge.StepArg{&gauge.StepArg{Name: "d"}, &gauge.StepArg{Name: "e"}, &gauge.StepArg{Name: "b"}, &gauge.StepArg{Name: "c"}, &gauge.StepArg{Name: "a"}}}

	agent := &rephraseRefactorer{step1, step2, false, nil}
	orderMap := agent.createOrderOfArgs()

	c.Assert(orderMap[0], Equals, 3)
	c.Assert(orderMap[1], Equals, -1)
	c.Assert(orderMap[2], Equals, 1)
	c.Assert(orderMap[3], Equals, 2)
	c.Assert(orderMap[4], Equals, 0)
}

func (s *MySuite) TestCreateOrderGivesMapOfOldArgsAndNewWhenArgsAreRemoved(c *C) {
	step1 := &gauge.Step{Args: []*gauge.StepArg{&gauge.StepArg{Name: "a"}, &gauge.StepArg{Name: "b"}, &gauge.StepArg{Name: "c"}, &gauge.StepArg{Name: "d"}}}
	step2 := &gauge.Step{Args: []*gauge.StepArg{&gauge.StepArg{Name: "d"}, &gauge.StepArg{Name: "b"}, &gauge.StepArg{Name: "c"}}}

	agent := &rephraseRefactorer{step1, step2, false, nil}
	orderMap := agent.createOrderOfArgs()

	c.Assert(orderMap[0], Equals, 3)
	c.Assert(orderMap[1], Equals, 1)
	c.Assert(orderMap[2], Equals, 2)
}

func (s *MySuite) TestCreationOfOrderMapForStep(c *C) {
	agent, _ := getRefactorAgent("Say <greeting> to <name>", "Say <greeting> to <name> \"DD\"", nil)

	orderMap := agent.createOrderOfArgs()

	c.Assert(orderMap[0], Equals, 0)
	c.Assert(orderMap[1], Equals, 1)
	c.Assert(orderMap[2], Equals, -1)
}

func (s *MySuite) TestRenamingWhenArgumentsIsAddedAtLast(c *C) {
	oldStep := "first step {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c>"
	newStep := "second step <a> and <b> and <c> and <d>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "name")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "number")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[3].Value, Equals, "d")
}

func (s *MySuite) TestRenamingWhenArgumentsIsAddedAtFirst(c *C) {
	oldStep := "first step {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c>"
	newStep := "second step <d> and <a> and <b> and <c>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "d")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "name")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[3].Value, Equals, "number")
}

func (s *MySuite) TestRenamingWhenArgumentsIsAddedInMiddle(c *C) {
	oldStep := "first step {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c>"
	newStep := "second step <a> and <d> and <b> and <c>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "name")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "d")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[3].Value, Equals, "number")
}

func (s *MySuite) TestRenamingWhenArgumentsIsRemovedFromLast(c *C) {
	oldStep := "first step {static} and {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c> and <d>"
	newStep := "second step <a> and <b> and <c>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number", "id"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "name")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "number")
}

func (s *MySuite) TestRenamingWhenArgumentsIsRemovedFromBegining(c *C) {
	oldStep := "first step {static} and {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c> and <d>"
	newStep := "second step <b> and <c> and <d>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number", "id"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "number")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "id")
}

func (s *MySuite) TestRenamingWhenArgumentsIsRemovedFromMiddle(c *C) {
	oldStep := "first step {static} and {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c> and <d>"
	newStep := "second step <a> and <b> and <d>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number", "id"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)

	c.Assert(specs[0].Scenarios[0].Steps[0].Value, Equals, "second step {} and {} and {}")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[0].Value, Equals, "name")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[1].Value, Equals, "address")
	c.Assert(specs[0].Scenarios[0].Steps[0].Args[2].Value, Equals, "id")
}

func (s *MySuite) TestGenerateNewStepNameGivesLineTextWithActualParamNames(c *C) {
	args := []string{"name", "address", "id"}
	newStep := "second step <a> and <b> and <d>"
	orderMap := make(map[int]int)
	orderMap[0] = 1
	orderMap[1] = 2
	orderMap[2] = 0
	agent, _ := getRefactorAgent(newStep, newStep, nil)
	linetext := agent.generateNewStepName(args, orderMap)

	c.Assert(linetext, Equals, "second step <address> and <id> and <name>")
}

func (s *MySuite) TestGenerateNewStepNameWhenParametersAreAdded(c *C) {
	args := []string{"name", "address"}
	newStep := "changed step <a> and <b> and \"id\""
	orderMap := make(map[int]int)
	orderMap[0] = 1
	orderMap[1] = 0
	orderMap[2] = -1
	agent, _ := getRefactorAgent(newStep, newStep, nil)
	linetext := agent.generateNewStepName(args, orderMap)

	c.Assert(linetext, Equals, "changed step <address> and <name> and \"id\"")
}

func (s *MySuite) TestGenerateNewStepNameWhenParametersAreRemoved(c *C) {
	args := []string{"name", "address", "desc"}
	newStep := "changed step <b> and \"id\""
	orderMap := make(map[int]int)
	orderMap[0] = 1
	orderMap[1] = -1
	orderMap[2] = -1
	agent, _ := getRefactorAgent(newStep, newStep, nil)
	linetext := agent.generateNewStepName(args, orderMap)

	c.Assert(linetext, Equals, "changed step <address> and \"id\"")
}

func (s *MySuite) TestGenerateNewStepNameWhenParametersAreUnchanged(c *C) {
	args := []string{"a"}
	newStep := "make comment <a>"
	agent, _ := getRefactorAgent("Comment <a>", newStep, nil)
	linetext := agent.generateNewStepName(args, agent.createOrderOfArgs())

	c.Assert(linetext, Equals, "make comment <a>")
}

func (s *MySuite) TestRefactoringInContextStep(c *C) {
	oldStep := "first step {static} and {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c> and <d>"
	newStep := "second step <d> and <b> and <c> and <a>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number", "id"}},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep + " sdf", LineNo: 3, Args: []string{"name", "address", "number", "id"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)
	c.Assert(specs[0].Contexts[0].Value, Equals, "second step {} and {} and {} and {}")
	c.Assert(specs[0].Contexts[0].Args[0].Value, Equals, "id")
	c.Assert(specs[0].Contexts[0].Args[1].Value, Equals, "address")
	c.Assert(specs[0].Contexts[0].Args[2].Value, Equals, "number")
	c.Assert(specs[0].Contexts[0].Args[3].Value, Equals, "name")
}

func (s *MySuite) TestRefactoringInTearDownStep(c *C) {
	oldStep := "first step {static} and {static} and {static} and {static}"
	oldStep1 := "first step <a> and <b> and <c> and <d>"
	newStep := "second step <d> and <b> and <c> and <a>"
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep + "sdf", LineNo: 3, Args: []string{"name", "address", "number", "id"}},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep + " sdf", LineNo: 3, Args: []string{"name", "address", "number", "id"}},
		&parser.Token{Kind: gauge.TearDownKind, Value: "____", LineNo: 3},
		&parser.Token{Kind: gauge.StepKind, Value: oldStep, LineNo: 3, Args: []string{"name", "address", "number", "id"}},
	}
	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	agent, _ := getRefactorAgent(oldStep1, newStep, nil)
	specs := append(make([]*gauge.Specification, 0), spec)
	dictionary := gauge.NewConceptDictionary()
	agent.rephraseInSpecsAndConcepts(&specs, dictionary)
	c.Assert(specs[0].TearDownSteps[0].Value, Equals, "second step {} and {} and {} and {}")
	c.Assert(specs[0].TearDownSteps[0].Args[0].Value, Equals, "id")
	c.Assert(specs[0].TearDownSteps[0].Args[1].Value, Equals, "address")
	c.Assert(specs[0].TearDownSteps[0].Args[2].Value, Equals, "number")
	c.Assert(specs[0].TearDownSteps[0].Args[3].Value, Equals, "name")
}

func (s *MySuite) TestRefactoringExternalSteps(c *C) {
	oldStep := "first step"
	newStep := "second step"
	agent, _ := getRefactorAgent(oldStep, newStep, nil)
	r := &mockRunner{
		response: &gauge_messages.Message{
			StepNameResponse: &gauge_messages.StepNameResponse{
				IsExternal:    true,
				IsStepPresent: true,
				HasAlias:      false,
				StepName:      []string{oldStep},
			},
		},
	}

	stepName, err, _ := agent.getStepNameFromRunner(r)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "external step: Cannot refactor 'first step' is in external project or library")
	c.Assert(stepName, Equals, "")
}
