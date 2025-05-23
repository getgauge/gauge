/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"path/filepath"
	"strings"

	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func assertStepEqual(c *C, expected, actual *gauge.Step) {
	c.Assert(expected.LineNo, Equals, actual.LineNo)
	c.Assert(expected.Value, Equals, actual.Value)
	c.Assert(expected.LineText, Equals, actual.LineText)
}

func (s *MySuite) TearDownTest(c *C) {
	config.ProjectRoot = ""
}

func (s *MySuite) TestConceptDictionaryAdd(c *C) {
	dictionary := gauge.NewConceptDictionary()
	step1Text := "test concept step 1"
	step2Text := "test concept step 2"
	step1 := &gauge.Step{Value: step1Text, LineNo: 1, IsConcept: true, LineText: step1Text}
	step2 := &gauge.Step{Value: step2Text, LineNo: 4, IsConcept: true, LineText: step2Text}
	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))

	concepts, errs, err := AddConcepts([]string{path}, dictionary)

	c.Assert(err, IsNil)
	c.Assert(len(concepts), Equals, 2)
	c.Assert(len(errs), Equals, 0)
	assertStepEqual(c, dictionary.ConceptsMap[step1Text].ConceptStep, step1)
	c.Assert(dictionary.ConceptsMap[step1Text].FileName, Equals, path)
	assertStepEqual(c, dictionary.ConceptsMap[step2Text].ConceptStep, step2)
	c.Assert(dictionary.ConceptsMap[step2Text].FileName, Equals, path)
}

func (s *MySuite) TestConceptDictionaryAddDuplicateConcept(c *C) {
	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "err", "cpt", "duplicate_concept.cpt"))

	concepts, errs, err := AddConcepts([]string{path}, dictionary)
	c.Assert(err, IsNil)

	c.Assert(len(concepts), Equals, 2)
	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(hasParseError("Duplicate concept definition found", path, 1, errs), Equals, true)
	c.Assert(hasParseError("Duplicate concept definition found", path, 4, errs), Equals, true)
}

func hasParseError(eMessage, fileName string, lineNo int, errs []ParseError) bool {
	for _, e := range errs {
		if e.Message == eMessage && e.FileName == fileName && e.LineNo == lineNo {
			return true
		}
	}
	return false
}

func (s *MySuite) TestDuplicateConceptsinMultipleFile(c *C) {
	dictionary := gauge.NewConceptDictionary()
	cpt1, _ := filepath.Abs(filepath.Join("testdata", "err", "cpt", "concept.cpt"))
	cpt2, _ := filepath.Abs(filepath.Join("testdata", "err", "cpt", "duplicate.cpt"))

	_, _, err := AddConcepts([]string{cpt1}, dictionary)
	c.Assert(err, IsNil)
	concepts, errs, err := AddConcepts([]string{cpt2}, dictionary)
	c.Assert(err, IsNil)

	c.Assert(len(concepts), Equals, 2)
	c.Assert(len(errs), Equals, 4)
	c.Assert(hasParseError("Duplicate concept definition found", cpt1, 1, errs), Equals, true)
	c.Assert(hasParseError("Duplicate concept definition found", cpt1, 4, errs), Equals, true)
	c.Assert(hasParseError("Duplicate concept definition found", cpt2, 1, errs), Equals, true)
	c.Assert(hasParseError("Duplicate concept definition found", cpt2, 4, errs), Equals, true)
}

func (s *MySuite) TestCreateConceptDictionaryGivesAllParseErrors(c *C) {
	config.ProjectRoot, _ = filepath.Abs(filepath.Join("testdata", "err", "cpt"))

	_, res, err := CreateConceptsDictionary()

	c.Assert(err, IsNil)
	c.Assert(res.Ok, Equals, false)
	c.Assert(len(res.ParseErrors), Equals, 9)
}

func (s *MySuite) TestCreateConceptDictionary(c *C) {
	config.ProjectRoot, _ = filepath.Abs(filepath.Join("testdata", "dir1"))

	dict, res, err := CreateConceptsDictionary()

	c.Assert(err, IsNil)
	c.Assert(res.Ok, Equals, true)
	c.Assert(dict, NotNil)
	c.Assert(len(dict.ConceptsMap), Equals, 1)
}

func (s *MySuite) TestConceptDictionaryWithNestedConcepts(c *C) {
	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "nested_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, dictionary)
	c.Assert(err, IsNil)
	concept := dictionary.Search("test concept step 1")

	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 1)

	actualNestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)
	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 1)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, "step 2")
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithDynamicParameters(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	concept := conceptDictionary.Search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 1)
	actualNestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)

	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(actualNestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(actualNestedConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")

	c.Assert(actualNestedConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(actualNestedConcept.ConceptSteps[1].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(actualNestedConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithStaticParameters(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "static_param_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	concept := conceptDictionary.Search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 2)
	actualNestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)

	c.Assert(actualNestedConcept.Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(actualNestedConcept.Args[0].Value, Equals, "user-id")

	c.Assert(actualNestedConcept.Args[1].ArgType, Equals, gauge.Static)
	c.Assert(actualNestedConcept.Args[1].Value, Equals, "static-value")
	useridArg, _ := actualNestedConcept.Lookup.GetArg("userid")
	usernameArg, _ := actualNestedConcept.Lookup.GetArg("username")
	c.Assert(useridArg.Value, Equals, "user-id")
	c.Assert(useridArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(usernameArg.Value, Equals, "static-value")
	c.Assert(usernameArg.ArgType, Equals, gauge.Static)

	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(actualNestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(actualNestedConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")

	c.Assert(actualNestedConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(actualNestedConcept.ConceptSteps[1].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(actualNestedConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
}

func (s *MySuite) TestConceptHavingItemsWithComments(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	concept := conceptDictionary.Search("create user {} {} and {}")

	c.Assert(len(concept.ConceptStep.Items), Equals, 3)
	c.Assert(concept.ConceptStep.Items[2].(*gauge.Comment).Value, Equals, "Comments")

	concept = conceptDictionary.Search("assign id {} and name {}")

	c.Assert(len(concept.ConceptStep.Items), Equals, 4)
	c.Assert(concept.ConceptStep.Items[3].(*gauge.Comment).Value, Equals, "Comment1")
}

func (s *MySuite) TestConceptHavingItemsComments(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "tabular_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)

	concept := conceptDictionary.Search("my concept {}")
	c.Assert(len(concept.ConceptStep.Items), Equals, 3)
	c.Assert(len(concept.ConceptStep.PreComments), Equals, 1)
	c.Assert(concept.ConceptStep.PreComments[0].Value, Equals, "COMMENT")
	c.Assert(concept.ConceptStep.Items[2].(*gauge.Comment).Value, Equals, "   comment")
}

func TestConceptHavingItemsWithTables(t *testing.T) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "tabular_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	if err != nil {
		t.Error(err)
	}

	concept := conceptDictionary.Search("my concept {}")
	if len(concept.ConceptStep.Items) != 3 {
		t.Errorf("Incorrect number of items; want %d, got %d", 3, len(concept.ConceptStep.Items))
	}
	if got := concept.ConceptStep.Items[1].Kind(); got != gauge.StepKind {
		t.Errorf("Incorrect concept step item kind; want %d, got %d", gauge.StepKind, got)
	}
	if concept.ConceptStep.Items[1].(*gauge.Step).HasInlineTable == false {
		t.Errorf("Concept Step does not have inline table")
	}
}

func TestConceptHavingConceptStepWithInlineTable(t *testing.T) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "tabular_concept2.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	if err != nil {
		t.Error(err)
	}

	concept := conceptDictionary.Search("my concept")
	if got := len(concept.ConceptStep.Items); got != 2 {
		t.Errorf("Incorrect number of concept step items; want %d, got %d", 2, got)
	}
	anotherConceptStep := concept.ConceptStep.Items[1].(*gauge.Step)
	if anotherConceptStep.HasInlineTable == false {
		t.Errorf("Expected first Item to have inline table")
	}
	if anotherConceptStep.IsConcept == false {
		t.Errorf("Expected a nested concept step")
	}
	if len(anotherConceptStep.Args) != 2 {
		t.Errorf("Incorrect number of Args for concept step")
	}
	if anotherConceptStep.Args[0].ArgValue() != "bar" {
		t.Errorf("Incorrect first param value; want %s, got %s", "bar", anotherConceptStep.Args[0].ArgValue())
	}
	if anotherConceptStep.Args[1].ArgType != gauge.TableArg {
		t.Errorf("Incorrect second param value; want %s, got %s", gauge.TableArg, anotherConceptStep.Args[1].ArgType)
	}
}

func (s *MySuite) TestMultiLevelConcept(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "nested_concept2.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	actualTopLevelConcept := conceptDictionary.Search("top level concept")
	c.Assert(len(actualTopLevelConcept.ConceptStep.ConceptSteps), Equals, 2)
	actualNestedConcept := actualTopLevelConcept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)
	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, "another nested concept")
	c.Assert(actualNestedConcept.ConceptSteps[1].Value, Equals, "normal step 2")
	c.Assert(actualTopLevelConcept.ConceptStep.ConceptSteps[1].Value, Equals, "normal step 1")

	actualAnotherNestedConcept := conceptDictionary.Search("another nested concept")
	c.Assert(len(actualAnotherNestedConcept.ConceptStep.ConceptSteps), Equals, 1)
	step := actualAnotherNestedConcept.ConceptStep.ConceptSteps[0]
	c.Assert(step.IsConcept, Equals, false)
	c.Assert(step.Value, Equals, "normal step 3")

	nestedConcept2 := conceptDictionary.Search("nested concept")
	c.Assert(len(nestedConcept2.ConceptStep.ConceptSteps), Equals, 2)
	actualAnotherNestedConcept2 := nestedConcept2.ConceptStep.ConceptSteps[0]
	c.Assert(actualAnotherNestedConcept2.IsConcept, Equals, true)
	c.Assert(len(actualAnotherNestedConcept2.ConceptSteps), Equals, 1)
	c.Assert(actualAnotherNestedConcept2.ConceptSteps[0].Value, Equals, "normal step 3")
	c.Assert(nestedConcept2.ConceptStep.ConceptSteps[1].Value, Equals, "normal step 2")
}

func (s *MySuite) TestParsingSimpleConcept(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.Parse("# my concept \n * first step \n * second step ", "")

	c.Assert(parseRes.Ok, Equals, true)
	c.Assert(len(parseRes.ParseErrors), Equals, 0)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 2)
	c.Assert(concept.ConceptSteps[0].Value, Equals, "first step")
	c.Assert(concept.ConceptSteps[1].Value, Equals, "second step")
}

func (s *MySuite) TestParsingConceptRetainsStepSuffix(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.Parse("# my concept \n * first step \n * second step \n\n", "")

	c.Assert(len(parseRes.ParseErrors), Equals, 0)
	c.Assert(parseRes.Ok, Equals, true)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 2)
	c.Assert(concept.ConceptSteps[0].Value, Equals, "first step")
	c.Assert(concept.ConceptSteps[1].Value, Equals, "second step")
	c.Assert(concept.ConceptSteps[0].Suffix, Equals, "")
	c.Assert(concept.ConceptSteps[1].Suffix, Equals, "\n")
}

func (s *MySuite) TestErrorParsingConceptHeadingWithStaticOrSpecialParameter(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.Parse("# my concept with \"parameter\" \n * first step \n * second step ", "foo.spec")
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.ParseErrors[0].Error(), Equals, "foo.spec:1 Concept heading can have only Dynamic Parameters => 'my concept with \"parameter\"'")

	_, parseRes = parser.Parse("# my concept with <table: foo> \n * first step \n * second step ", "foo2.spec")
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.ParseErrors[0].Error(), Equals, "foo2.spec:1 Dynamic parameter <table: foo> could not be resolved => 'my concept with <table: foo>'")
}

func (s *MySuite) TestErrorParsingConceptWithoutHeading(c *C) {
	parser := new(ConceptParser)

	_, parseRes := parser.Parse("* first step \n * second step ", "")

	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Step is not defined inside a concept heading")
}

func (s *MySuite) TestErrorParsingConceptWithoutSteps(c *C) {
	parser := new(ConceptParser)

	_, parseRes := parser.Parse("# my concept with \n", "")

	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Concept should have at least one step")
}

func (s *MySuite) TestParsingSimpleConceptWithParameters(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.Parse("# my concept with <param0> and <param1> \n * first step using <param0> \n * second step using \"value\" and <param1> ", "")

	c.Assert(len(parseRes.ParseErrors), Equals, 0)
	c.Assert(parseRes.Ok, Equals, true)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]
	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 2)
	// c.Assert(len(concept.Lookup.paramValue), Equals, 2)
	c.Assert(concept.Lookup.ContainsArg("param0"), Equals, true)
	c.Assert(concept.Lookup.ContainsArg("param1"), Equals, true)

	firstConcept := concept.ConceptSteps[0]
	c.Assert(firstConcept.Value, Equals, "first step using {}")
	c.Assert(len(firstConcept.Args), Equals, 1)
	c.Assert(firstConcept.Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(firstConcept.Args[0].Value, Equals, "param0")

	secondConcept := concept.ConceptSteps[1]
	c.Assert(secondConcept.Value, Equals, "second step using {} and {}")
	c.Assert(len(secondConcept.Args), Equals, 2)
	c.Assert(secondConcept.Args[0].ArgType, Equals, gauge.Static)
	c.Assert(secondConcept.Args[0].Value, Equals, "value")
	c.Assert(secondConcept.Args[1].ArgType, Equals, gauge.Dynamic)
	c.Assert(secondConcept.Args[1].Value, Equals, "param1")

}

func (s *MySuite) TestErrorParsingConceptStepWithInvalidParameters(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.Parse("# my concept with <param0> and <param1> \n * first step using <param3> \n * second step using \"value\" and <param1> ", "")

	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Dynamic parameter <param3> could not be resolved")
}

func (s *MySuite) TestParsingMultipleConcept(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.Parse("# my concept \n * first step \n * second step \n# my second concept \n* next step\n # my third concept <param0>\n * next step <param0> and \"value\"\n  ", "")

	c.Assert(len(parseRes.ParseErrors), Equals, 0)
	c.Assert(parseRes.Ok, Equals, true)
	c.Assert(len(concepts), Equals, 3)

	firstConcept := concepts[0]
	secondConcept := concepts[1]
	thirdConcept := concepts[2]

	c.Assert(firstConcept.IsConcept, Equals, true)
	c.Assert(len(firstConcept.ConceptSteps), Equals, 2)
	c.Assert(firstConcept.ConceptSteps[0].Value, Equals, "first step")
	c.Assert(firstConcept.ConceptSteps[1].Value, Equals, "second step")

	c.Assert(secondConcept.IsConcept, Equals, true)
	c.Assert(len(secondConcept.ConceptSteps), Equals, 1)
	c.Assert(secondConcept.ConceptSteps[0].Value, Equals, "next step")

	c.Assert(thirdConcept.IsConcept, Equals, true)
	c.Assert(len(thirdConcept.ConceptSteps), Equals, 1)
	c.Assert(thirdConcept.ConceptSteps[0].Value, Equals, "next step {} and {}")
	c.Assert(len(thirdConcept.ConceptSteps[0].Args), Equals, 2)
	c.Assert(thirdConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(thirdConcept.ConceptSteps[0].Args[1].ArgType, Equals, gauge.Static)

	// c.Assert(len(thirdConcept.Lookup.paramValue), Equals, 1)
	c.Assert(thirdConcept.Lookup.ContainsArg("param0"), Equals, true)

}

func (s *MySuite) TestParsingConceptStepWithInlineTable(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.Parse("# my concept <foo> \n * first step with <foo> and inline table\n |id|name|\n|1|vishnu|\n|2|prateek|\n", "")

	c.Assert(len(parseRes.ParseErrors), Equals, 0)
	c.Assert(parseRes.Ok, Equals, true)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 1)
	c.Assert(concept.ConceptSteps[0].Value, Equals, "first step with {} and inline table {}")

	tableArgument := concept.ConceptSteps[0].Args[1]
	c.Assert(tableArgument.ArgType, Equals, gauge.TableArg)

	inlineTable := tableArgument.Table
	c.Assert(inlineTable.IsInitialized(), Equals, true)
	idCells, _ := inlineTable.Get("id")
	nameCells, _ := inlineTable.Get("name")
	c.Assert(len(idCells), Equals, 2)
	c.Assert(len(nameCells), Equals, 2)
	c.Assert(idCells[0].Value, Equals, "1")
	c.Assert(idCells[0].CellType, Equals, gauge.Static)
	c.Assert(idCells[1].Value, Equals, "2")
	c.Assert(idCells[1].CellType, Equals, gauge.Static)
	c.Assert(nameCells[0].Value, Equals, "vishnu")
	c.Assert(nameCells[0].CellType, Equals, gauge.Static)
	c.Assert(nameCells[1].Value, Equals, "prateek")
	c.Assert(nameCells[1].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestErrorParsingConceptWithInvalidInlineTable(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.Parse("# my concept \n |id|name|\n|1|vishnu|\n|2|prateek|\n", "")

	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Table doesn't belong to any step")
}

func (s *MySuite) TestNestedConceptLooksUpArgsFromParent(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First flow").
		step("create user \"foo\" \"doo\"").
		step("another step").String()

	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "param_nested_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, dictionary)
	c.Assert(err, IsNil)
	tokens, _ := parser.GenerateTokens(specText, "")
	spec, parseResult, _ := parser.CreateSpecification(tokens, dictionary, "")

	c.Assert(parseResult.Ok, Equals, true)
	firstStepInSpec := spec.Scenarios[0].Steps[0]
	nestedConcept := firstStepInSpec.ConceptSteps[0]
	nestedConceptArg1, _ := nestedConcept.GetArg("baz")
	c.Assert(nestedConceptArg1.Value, Equals, "foo")
	nestedConceptArg2, _ := nestedConcept.GetArg("boo")
	c.Assert(nestedConceptArg2.Value, Equals, "doo")
}

func (s *MySuite) TestNestedConceptLooksUpDataTableArgs(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		tableHeader("456", "apoorva", "9800").
		tableHeader("789", "srikanth", "7900").
		scenarioHeading("First scenario").
		step("create user <id> <name>").
		step("another step").String()

	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "param_nested_concept.cpt"))

	_, _, err := AddConcepts([]string{path}, dictionary)
	c.Assert(err, IsNil)

	tokens, _ := parser.GenerateTokens(specText, "")
	spec, parseResult, _ := parser.CreateSpecification(tokens, dictionary, "")

	c.Assert(parseResult.Ok, Equals, true)

	firstStepInSpec := spec.Scenarios[0].Steps[0]
	c.Assert(firstStepInSpec.IsConcept, Equals, true)
	barArg, _ := firstStepInSpec.GetArg("bar")
	farArg, _ := firstStepInSpec.GetArg("far")
	c.Assert(barArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(farArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(barArg.Value, Equals, "id")
	c.Assert(farArg.Value, Equals, "name")

	nestedConcept := firstStepInSpec.ConceptSteps[0]
	bazArg, _ := nestedConcept.GetArg("baz")
	booArg, _ := nestedConcept.GetArg("boo")
	c.Assert(bazArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(booArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(bazArg.Value, Equals, "id")
	c.Assert(booArg.Value, Equals, "name")

}

func (s *MySuite) TestNestedConceptLooksUpWhenParameterPlaceholdersAreSame(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		tableHeader("456", "apoorva", "9800").
		tableHeader("789", "srikanth", "7900").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		step("another step").String()

	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "param_nested_concept2.cpt"))

	_, _, err := AddConcepts([]string{path}, dictionary)
	c.Assert(err, IsNil)

	tokens, _ := parser.GenerateTokens(specText, "")
	spec, parseResult, _ := parser.CreateSpecification(tokens, dictionary, "")

	c.Assert(parseResult.Ok, Equals, true)

	firstStepInSpec := spec.Scenarios[0].Steps[0]
	c.Assert(firstStepInSpec.IsConcept, Equals, true)
	useridArg, _ := firstStepInSpec.GetArg("user-id")
	usernameArg, _ := firstStepInSpec.GetArg("user-name")
	userphoneArg, _ := firstStepInSpec.GetArg("user-phone")
	c.Assert(useridArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(usernameArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(userphoneArg.ArgType, Equals, gauge.Dynamic)
	c.Assert(useridArg.Value, Equals, "id")
	c.Assert(usernameArg.Value, Equals, "name")
	c.Assert(userphoneArg.Value, Equals, "phone")

	nestedConcept := firstStepInSpec.ConceptSteps[0]
	useridArg2, _ := nestedConcept.GetArg("user-id")
	usernameArg2, _ := nestedConcept.GetArg("user-name")
	c.Assert(useridArg2.ArgType, Equals, gauge.Dynamic)
	c.Assert(usernameArg2.ArgType, Equals, gauge.Dynamic)
	c.Assert(useridArg2.Value, Equals, "id")
	c.Assert(usernameArg2.Value, Equals, "name")

}

func (s *MySuite) TestErrorOnCircularReferenceInConcept(c *C) {
	cd := gauge.NewConceptDictionary()
	cd.ConceptsMap["concept"] = &gauge.Concept{ConceptStep: &gauge.Step{LineText: "concept", Value: "concept", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: true}}}, FileName: "filename.cpt"}

	res := ValidateConcepts(cd)

	c.Assert(containsAny(res.ParseErrors, "Circular reference found"), Equals, true)
}

func (s *MySuite) TestValidateConceptShouldRemoveCircularConceptsConceptStepFromDictionary(c *C) {
	cd := gauge.NewConceptDictionary()
	cd.ConceptsMap["concept"] = &gauge.Concept{ConceptStep: &gauge.Step{LineText: "concept", Value: "concept", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: true}}}, FileName: "filename.cpt"}
	cd.ConceptsMap["concept2"] = &gauge.Concept{ConceptStep: &gauge.Step{LineText: "concept2", Value: "concept2", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: true}}}, FileName: "filename.cpt"}

	res := ValidateConcepts(cd)

	c.Assert(cd.ConceptsMap["concept"], Equals, (*gauge.Concept)(nil))
	c.Assert(len(cd.ConceptsMap["concept2"].ConceptStep.ConceptSteps), Equals, 0)
	c.Assert(len(res.ParseErrors), Equals, 2)
	c.Assert(strings.Contains(res.ParseErrors[0].Message, "Circular reference found"), Equals, true)
	c.Assert(strings.Contains(res.ParseErrors[1].Message, "Circular reference found"), Equals, true)
}

func (s *MySuite) TestValidateConceptShouldRemoveCircularConceptsFromDictionary(c *C) {
	cd := gauge.NewConceptDictionary()
	c1 := &gauge.Step{LineText: "concept", Value: "concept", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept2", Value: "concept2", IsConcept: true}}}
	c2 := &gauge.Step{LineText: "concept2", Value: "concept2", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: true}}}
	_, err := AddConcept([]*gauge.Step{c1, c2}, "filename.cpt", cd)
	c.Assert(err, IsNil)

	res := ValidateConcepts(cd)

	c.Assert(cd.ConceptsMap["concept"], Equals, (*gauge.Concept)(nil))
	c.Assert(cd.ConceptsMap["concept2"], Equals, (*gauge.Concept)(nil))
	c.Assert(len(res.ParseErrors), Equals, 2)
	c.Assert(strings.Contains(res.ParseErrors[0].Message, "Circular reference found"), Equals, true)
	c.Assert(strings.Contains(res.ParseErrors[1].Message, "Circular reference found"), Equals, true)
}

func (s *MySuite) TestRemoveAllReferences(c *C) {
	cd := gauge.NewConceptDictionary()
	cpt1 := &gauge.Concept{ConceptStep: &gauge.Step{LineText: "concept", Value: "concept", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: true}}}, FileName: "filename.cpt"}
	cd.ConceptsMap["concept"] = cpt1
	cd.ConceptsMap["concept2"] = &gauge.Concept{ConceptStep: &gauge.Step{LineText: "concept2", Value: "concept2", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: true}}}, FileName: "filename.cpt"}

	c.Assert(len(cd.ConceptsMap["concept2"].ConceptStep.ConceptSteps), Equals, 1)

	removeAllReferences(cd, cpt1)

	c.Assert(len(cd.ConceptsMap["concept2"].ConceptStep.ConceptSteps), Equals, 0)
}

func (s *MySuite) TestReplaceNestedConceptsWithCircularReference(c *C) {
	lookup := gauge.ArgLookup{}
	lookup.AddArgName("a")
	err := lookup.AddArgValue("a", &gauge.StepArg{Name: "", Value: "a", ArgType: gauge.Static})
	c.Assert(err, IsNil)
	lookup.ParamIndexMap = make(map[string]int)
	lookup.ParamIndexMap["a"] = 0

	cd := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "err", "cpt", "circular_concept.cpt"))

	_, _, err = AddConcepts([]string{path}, cd)
	c.Assert(err, IsNil)
	concept := cd.Search("concept1 {}")

	c.Assert(concept.ConceptStep.ConceptSteps[0].Lookup, DeepEquals, lookup)
}

func (s *MySuite) TestErrorParsingConceptWithRecursiveCallToConcept(c *C) {
	cd := gauge.NewConceptDictionary()
	cd.ConceptsMap["concept"] = &gauge.Concept{ConceptStep: &gauge.Step{LineText: "concept", Value: "concept", IsConcept: true, ConceptSteps: []*gauge.Step{&gauge.Step{LineText: "concept", Value: "concept", IsConcept: false}}}, FileName: "filename.cpt"}

	res := ValidateConcepts(cd)

	c.Assert(len(res.ParseErrors), Not(Equals), 0)
	c.Assert(containsAny(res.ParseErrors, "Circular reference found"), Equals, true)
}

func (s *MySuite) TestConceptHavingDynamicParameters(c *C) {
	conceptText := newSpecBuilder().
		specHeading("create user <user:id> <user:name> and <file>").
		step("a step <user:id>").String()
	step, _ := new(ConceptParser).Parse(conceptText, "")
	c.Assert(step[0].LineText, Equals, "create user <user:id> <user:name> and <file>")
	c.Assert(step[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(step[0].Args[1].ArgType, Equals, gauge.Dynamic)
	c.Assert(step[0].Args[2].ArgType, Equals, gauge.Dynamic)
}

func (s *MySuite) TestConceptHavingInvalidSpecialParameters(c *C) {
	conceptText := newSpecBuilder().
		specHeading("create user <user:id> <table:name> and <file>").
		step("a step <user:id>").String()
	_, parseRes := new(ConceptParser).Parse(conceptText, "")
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Dynamic parameter <table:name> could not be resolved")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestConceptHavingStaticParameters(c *C) {
	conceptText := newSpecBuilder().
		specHeading("create user <user:id> \"abc\" and <file>").
		step("a step <user:id>").String()
	_, parseRes := new(ConceptParser).Parse(conceptText, "")
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Concept heading can have only Dynamic Parameters")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestConceptFileHavingScenarioHeadingGivesParseError(c *C) {
	conceptText := newSpecBuilder().
		specHeading("create user").
		step("a step").
		scenarioHeading("Scenario Heading").
		step("a step1").
		String()

	scenarioHeading := newSpecBuilder().
		scenarioHeading("Scenario Heading").
		String()
	_, parseRes := new(ConceptParser).Parse(conceptText, "")

	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Scenario Heading is not allowed in concept file")
	c.Assert(parseRes.ParseErrors[0].LineText, Equals, strings.TrimSpace(scenarioHeading))
}

func (s *MySuite) TestConceptFileHavingStaticParamsInHeadingShouldGiveParseError(c *C) {
	conceptText := newSpecBuilder().
		specHeading("Concept Heading37a").
		step("a step").
		specHeading("testinghjk \"sdf\"").
		step("a step1").
		String()

	_, parseRes := new(ConceptParser).Parse(conceptText, "")

	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Concept heading can have only Dynamic Parameters")
	c.Assert(parseRes.ParseErrors[0].LineText, Equals, "testinghjk \"sdf\"")
}

func (s *MySuite) TestConceptFileHavingTableAfterConceptHeadingShouldGiveParseError(c *C) {
	conceptText := newSpecBuilder().
		specHeading("Concept Heading37a").
		step("a step").
		specHeading("testinghjk ").
		text("|sdfsdf|").
		text("|----|").
		text("|wer|").
		step("a step1").
		String()

	_, parseRes := new(ConceptParser).Parse(conceptText, "")

	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Table doesn't belong to any step")
	c.Assert(parseRes.ParseErrors[0].LineText, Equals, "|sdfsdf|")
}

func (s *MySuite) TestMultipleConceptsInAFileHavingErrorsShouldBeConsolidated(c *C) {
	conceptText := newSpecBuilder().
		specHeading("1<werwer>").
		step("self <werwe1r>").
		specHeading("2 <werwer> two").
		step("self <werwer>").
		String()

	_, parseRes := new(ConceptParser).Parse(conceptText, "")

	c.Assert(len(parseRes.ParseErrors), Not(Equals), 0)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Dynamic parameter <werwe1r> could not be resolved")
	c.Assert(parseRes.ParseErrors[0].LineText, Equals, "self <werwe1r>")
}

func (s *MySuite) TestConceptFileHavingItemsWithDuplicateTableHeaders(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "tabular_concept1.cpt"))

	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	concept := conceptDictionary.Search("my concept {}")
	concept1 := conceptDictionary.Search("my {}")

	c.Assert(concept, Not(Equals), nil)
	c.Assert(concept1, Not(Equals), nil)
}

func (s *MySuite) TestConceptParserShouldNotAddTableAsArgIfCommentsArePresentBetweenStepAndTable(c *C) {
	conceptText := newSpecBuilder().
		specHeading("create user").
		step("a step").
		text("").
		text("adasdasd\n\n").
		text("|sdfsdf|").
		text("|----|").
		text("|wer|").
		step("a step1").
		String()
	steps, _ := new(ConceptParser).Parse(conceptText, "")
	c.Assert(steps[0].ConceptSteps[0].GetLineText(), Equals, "a step")
}

func (s *MySuite) TestErrorParsingConceptWithNoSteps(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.Parse("# my concept\n# second concept\n* first step ", "foo.cpt")
	c.Assert(len(parseRes.ParseErrors), Equals, 1)
	c.Assert(parseRes.Ok, Equals, false)
	c.Assert(parseRes.ParseErrors[0].Error(), Equals, "foo.cpt:1 Concept should have at least one step => 'my concept'")
}

func containsAny(errs []ParseError, msg string) bool {
	for _, err := range errs {
		if strings.Contains(err.Message, msg) {
			return true
		}
	}
	return false
}
