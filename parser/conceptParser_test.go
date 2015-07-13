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

package parser

import (
	. "gopkg.in/check.v1"
	"strings"
)

func (s *MySuite) TestConceptDictionaryAdd(c *C) {
	dictionary := new(ConceptDictionary)
	step1 := &Step{Value: "test step 1"}
	step2 := &Step{Value: "test step 2"}

	err := dictionary.Add([]*Step{step1, step2}, "file.cpt")

	c.Assert(err, IsNil)
	c.Assert(dictionary.ConceptsMap["test step 1"].ConceptStep, Equals, step1)
	c.Assert(dictionary.ConceptsMap["test step 1"].FileName, Equals, "file.cpt")
	c.Assert(dictionary.ConceptsMap["test step 2"].ConceptStep, Equals, step2)
	c.Assert(dictionary.ConceptsMap["test step 2"].FileName, Equals, "file.cpt")
}

func (s *MySuite) TestConceptDictionaryAddDuplicateConcept(c *C) {
	dictionary := new(ConceptDictionary)
	step1 := &Step{Value: "test step {}", LineText: "test step <first>"}
	step2 := &Step{Value: "test step {}", LineText: "test step <second>"}

	err := dictionary.Add([]*Step{step1, step2}, "file.cpt")

	c.Assert(err, NotNil)
	c.Assert(err.Message, Equals, "Duplicate concept definition found")
}

func (s *MySuite) TestConceptDictionaryWithNestedConcepts(c *C) {
	dictionary := new(ConceptDictionary)
	normalStep1 := &Step{Value: "normal step 1", LineText: "normal step 1"}
	normalStep2 := &Step{Value: "normal step 2", LineText: "normal step 2"}
	nestedConceptStep := &Step{Value: "nested concept", LineText: "nested concept"}
	nestedConcept := &Step{Value: "nested concept", LineText: "nested concept", IsConcept: true, ConceptSteps: []*Step{normalStep2}}

	topLevelConcept := &Step{Value: "top level concept", IsConcept: true, ConceptSteps: []*Step{nestedConceptStep, normalStep1}}

	dictionary.Add([]*Step{nestedConcept}, "file1.cpt")
	dictionary.Add([]*Step{topLevelConcept}, "file2.cpt")

	concept := dictionary.search("top level concept")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 2)
	actualnestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualnestedConcept.IsConcept, Equals, true)
	c.Assert(len(actualnestedConcept.ConceptSteps), Equals, 1)
	c.Assert(actualnestedConcept.ConceptSteps[0].Value, Equals, normalStep2.Value)
	c.Assert(concept.ConceptStep.ConceptSteps[1].Value, Equals, normalStep1.Value)
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithParameter(c *C) {
	conceptDictionary := new(ConceptDictionary)
	conceptText := SpecBuilder().
		specHeading("assign id").
		step("add id").
		specHeading("create user").
		step("assign id").
		step("assign id").String()
	concepts, _ := new(ConceptParser).parse(conceptText)
	conceptDictionary.Add(concepts, "file.cpt")
	concept := conceptDictionary.search("create user")
	c.Assert(concept.ConceptStep.ConceptSteps[0].Value, Equals, "assign id")
	c.Assert(concept.ConceptStep.ConceptSteps[0].IsConcept, Equals, true)
	c.Assert(concept.ConceptStep.ConceptSteps[1].Value, Equals, "assign id")
	c.Assert(concept.ConceptStep.ConceptSteps[1].IsConcept, Equals, true)
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithDefinitionAfterUsage(c *C) {
	conceptDictionary := new(ConceptDictionary)
	conceptText := SpecBuilder().
		specHeading("create user").
		step("assign id").
		step("assign id").
		step("assign id").
		specHeading("assign id").
		step("add id").String()
	concepts, _ := new(ConceptParser).parse(conceptText)
	conceptDictionary.Add(concepts, "file.cpt")
	concept := conceptDictionary.search("create user")
	c.Assert(concept.ConceptStep.ConceptSteps[0].Value, Equals, "assign id")
	c.Assert(concept.ConceptStep.ConceptSteps[1].Value, Equals, "assign id")
	c.Assert(concept.ConceptStep.ConceptSteps[2].Value, Equals, "assign id")
	c.Assert(concept.ConceptStep.ConceptSteps[0].IsConcept, Equals, true)
	c.Assert(concept.ConceptStep.ConceptSteps[1].IsConcept, Equals, true)
	c.Assert(concept.ConceptStep.ConceptSteps[2].IsConcept, Equals, true)
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithParameters(c *C) {
	conceptDictionary := new(ConceptDictionary)
	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()
	concepts, _ := new(ConceptParser).parse(conceptText)
	conceptDictionary.Add(concepts, "file.cpt")

	concept := conceptDictionary.search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 1)
	actualnestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualnestedConcept.IsConcept, Equals, true)

	c.Assert(len(actualnestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualnestedConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(actualnestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(len(concepts[0].Items), Equals, 2)

	c.Assert(actualnestedConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(actualnestedConcept.ConceptSteps[1].Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(len(concepts[1].Items), Equals, 3)
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithStaticParameters(c *C) {
	conceptDictionary := new(ConceptDictionary)
	conceptText := SpecBuilder().
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name \"static-value\"").String()
	concepts, _ := new(ConceptParser).parse(conceptText)
	conceptDictionary.Add(concepts, "file.cpt")

	concept := conceptDictionary.search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 1)
	actualNestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)

	c.Assert(actualNestedConcept.Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.Args[0].Value, Equals, "user-id")

	c.Assert(actualNestedConcept.Args[1].ArgType, Equals, Static)
	c.Assert(actualNestedConcept.Args[1].Value, Equals, "static-value")
	c.Assert(actualNestedConcept.Lookup.getArg("userid").Value, Equals, "user-id")
	c.Assert(actualNestedConcept.Lookup.getArg("userid").ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.Lookup.getArg("username").Value, Equals, "static-value")
	c.Assert(actualNestedConcept.Lookup.getArg("username").ArgType, Equals, Static)

	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(actualNestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(len(concepts[0].Items), Equals, 3)

	c.Assert(actualNestedConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(actualNestedConcept.ConceptSteps[1].Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(len(concepts[1].Items), Equals, 2)
}

func (s *MySuite) TestConceptHavingItemsWithComments(c *C) {
	conceptDictionary := new(ConceptDictionary)
	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		text("Comments").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").
		text("Comment1").String()
	concepts, _ := new(ConceptParser).parse(conceptText)
	conceptDictionary.Add(concepts, "file.cpt")

	concept := conceptDictionary.search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 1)
	actualnestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualnestedConcept.IsConcept, Equals, true)

	c.Assert(len(actualnestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualnestedConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(actualnestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(len(concepts[0].Items), Equals, 3)
	c.Assert(concepts[0].Items[2].(*Comment).Value, Equals, "Comments")

	c.Assert(actualnestedConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(actualnestedConcept.ConceptSteps[1].Args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(len(concepts[1].Items), Equals, 4)
}

func (s *MySuite) TestConceptHavingItemsWithTablesAndPreComments(c *C) {
	conceptDictionary := new(ConceptDictionary)
	concepts, _ := new(ConceptParser).parse("COMMENT\n# my concept <foo> \n * first step with <foo> and inline table\n |id|name|\n|1|vishnu|\n|2|prateek|\n comment")
	conceptDictionary.Add(concepts, "concept.cpt")

	c.Assert(len(concepts[0].Items), Equals, 3)
	c.Assert(len(concepts[0].PreComments), Equals, 1)
	c.Assert(concepts[0].PreComments[0].Value, Equals, "COMMENT")
}

/*
#top level concept
* nested concept
* normal step 1

#top level concept 2
* nested concept

# nested concept
* normal step 2
*/
func (s *MySuite) TestNestedConceptsWhenReferencedConceptParsedLater(c *C) {
	dictionary := new(ConceptDictionary)
	normalStep1 := &Step{Value: "normal step 1", LineText: "normal step 1"}
	normalStep2 := &Step{Value: "normal step 2", LineText: "normal step 2"}

	nestedConceptStep := &Step{Value: "nested concept", LineText: "nested concept"}

	topLevelConcept := &Step{Value: "top level concept", IsConcept: true, ConceptSteps: []*Step{nestedConceptStep, normalStep1}}
	anotherTopLevelConcept := &Step{Value: "top level concept 2", IsConcept: true, ConceptSteps: []*Step{nestedConceptStep}}
	nestedConcept := &Step{Value: "nested concept", LineText: "nested concept", IsConcept: true, ConceptSteps: []*Step{normalStep2}}

	dictionary.Add([]*Step{topLevelConcept}, "file1.cpt")
	dictionary.Add([]*Step{anotherTopLevelConcept}, "file1.cpt")
	dictionary.Add([]*Step{nestedConcept}, "file2.cpt")

	concept := dictionary.search("top level concept")
	c.Assert(len(concept.ConceptStep.ConceptSteps), Equals, 2)
	actualNestedConcept := concept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)
	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 1)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, normalStep2.Value)
	c.Assert(concept.ConceptStep.ConceptSteps[1].Value, Equals, normalStep1.Value)

	topLevelConcept2 := dictionary.search("top level concept 2")
	c.Assert(len(topLevelConcept2.ConceptStep.ConceptSteps), Equals, 1)
	actualNestedConcept = topLevelConcept2.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)
	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 1)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, normalStep2.Value)
}

/*
# top level concept
* nested concept
* normal step 1

# another nested concept
* normal step 3

# nested concept
* another nested concept
* normal step 2
*/
func (s *MySuite) TestMultiLevelConcept(c *C) {
	dictionary := new(ConceptDictionary)
	normalStep1 := &Step{Value: "normal step 1", LineText: "normal step 1"}
	normalStep2 := &Step{Value: "normal step 2", LineText: "normal step 2"}
	normalStep3 := &Step{Value: "normal step 3", LineText: "normal step 3"}
	nestedConceptStep := &Step{Value: "nested concept", LineText: "nested concept"}

	topLevelConcept := &Step{Value: "top level concept", IsConcept: true, ConceptSteps: []*Step{nestedConceptStep, normalStep1}}
	anotherNestedConcept := &Step{Value: "another nested concept", IsConcept: true, ConceptSteps: []*Step{normalStep3}}
	nestedConcept := &Step{Value: "nested concept", IsConcept: true, ConceptSteps: []*Step{anotherNestedConcept, normalStep2}}

	dictionary.Add([]*Step{topLevelConcept}, "file1.cpt")
	dictionary.Add([]*Step{anotherNestedConcept}, "file1.cpt")
	dictionary.Add([]*Step{nestedConcept}, "file1.cpt")

	actualTopLevelConcept := dictionary.search("top level concept")
	c.Assert(len(actualTopLevelConcept.ConceptStep.ConceptSteps), Equals, 2)
	actualNestedConcept := actualTopLevelConcept.ConceptStep.ConceptSteps[0]
	c.Assert(actualNestedConcept.IsConcept, Equals, true)
	c.Assert(len(actualNestedConcept.ConceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.ConceptSteps[0].Value, Equals, anotherNestedConcept.Value)
	c.Assert(actualNestedConcept.ConceptSteps[1].Value, Equals, normalStep2.Value)
	c.Assert(actualTopLevelConcept.ConceptStep.ConceptSteps[1].Value, Equals, normalStep1.Value)

	actualAnotherNestedConcept := dictionary.search("another nested concept")
	c.Assert(len(actualAnotherNestedConcept.ConceptStep.ConceptSteps), Equals, 1)
	step := actualAnotherNestedConcept.ConceptStep.ConceptSteps[0]
	c.Assert(step.IsConcept, Equals, false)
	c.Assert(step.Value, Equals, normalStep3.Value)

	nestedConcept2 := dictionary.search("nested concept")
	c.Assert(len(nestedConcept2.ConceptStep.ConceptSteps), Equals, 2)
	actualAnotherNestedConcept2 := nestedConcept2.ConceptStep.ConceptSteps[0]
	c.Assert(actualAnotherNestedConcept2.IsConcept, Equals, true)
	c.Assert(len(actualAnotherNestedConcept2.ConceptSteps), Equals, 1)
	c.Assert(actualAnotherNestedConcept2.ConceptSteps[0].Value, Equals, normalStep3.Value)
	c.Assert(nestedConcept2.ConceptStep.ConceptSteps[1].Value, Equals, normalStep2.Value)

}

func (s *MySuite) TestConceptDictionarySearch(c *C) {
	dictionary := new(ConceptDictionary)
	step1 := &Step{Value: "test step 1"}
	step2 := &Step{Value: "test step 2"}

	dictionary.Add([]*Step{step1, step2}, "file.cpt")

	c.Assert(dictionary.search(step1.Value).ConceptStep, Equals, step1)
	c.Assert(dictionary.search(step1.Value).FileName, Equals, "file.cpt")
	c.Assert(dictionary.search(step2.Value).ConceptStep, Equals, step2)
	c.Assert(dictionary.search(step2.Value).FileName, Equals, "file.cpt")
}

func (s *MySuite) TestParsingSimpleConcept(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept \n * first step \n * second step ")

	c.Assert(parseRes.Error, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 2)
	c.Assert(concept.ConceptSteps[0].Value, Equals, "first step")
	c.Assert(concept.ConceptSteps[1].Value, Equals, "second step")

}

func (s *MySuite) TestErrorParsingConceptHeadingWithStaticOrSpecialParameter(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept with \"paratemer\" \n * first step \n * second step ")
	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Concept heading can have only Dynamic Parameters")

	_, parseRes = parser.parse("# my concept with <table: foo> \n * first step \n * second step ")
	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Dynamic parameter <table: foo> could not be resolved")

}

func (s *MySuite) TestErrorParsingConceptWithoutHeading(c *C) {
	parser := new(ConceptParser)

	_, parseRes := parser.parse("* first step \n * second step ")

	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Step is not defined inside a concept heading")
}

func (s *MySuite) TestErrorParsingConceptWithoutSteps(c *C) {
	parser := new(ConceptParser)

	_, parseRes := parser.parse("# my concept with \n")

	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Concept should have atleast one step")
}

func (s *MySuite) TestParsingSimpleConceptWithParameters(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept with <param0> and <param1> \n * first step using <param0> \n * second step using \"value\" and <param1> ")

	c.Assert(parseRes.Error, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]
	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 2)
	c.Assert(len(concept.Lookup.paramValue), Equals, 2)
	c.Assert(concept.Lookup.containsArg("param0"), Equals, true)
	c.Assert(concept.Lookup.containsArg("param1"), Equals, true)

	firstConcept := concept.ConceptSteps[0]
	c.Assert(firstConcept.Value, Equals, "first step using {}")
	c.Assert(len(firstConcept.Args), Equals, 1)
	c.Assert(firstConcept.Args[0].ArgType, Equals, Dynamic)
	c.Assert(firstConcept.Args[0].Value, Equals, "param0")

	secondConcept := concept.ConceptSteps[1]
	c.Assert(secondConcept.Value, Equals, "second step using {} and {}")
	c.Assert(len(secondConcept.Args), Equals, 2)
	c.Assert(secondConcept.Args[0].ArgType, Equals, Static)
	c.Assert(secondConcept.Args[0].Value, Equals, "value")
	c.Assert(secondConcept.Args[1].ArgType, Equals, Dynamic)
	c.Assert(secondConcept.Args[1].Value, Equals, "param1")

}

func (s *MySuite) TestErrorParsingConceptWithRecursiveCallToConcept(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept \n * first step using \n * my concept ")

	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Cyclic dependancy found. Step is calling concept again.")
}

func (s *MySuite) TestErrorParsingConceptStepWithInvalidParameters(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept with <param0> and <param1> \n * first step using <param3> \n * second step using \"value\" and <param1> ")

	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Dynamic parameter <param3> could not be resolved")
}

func (s *MySuite) TestParsingMultipleConcept(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept \n * first step \n * second step \n# my second concept \n* next step\n # my third concept <param0>\n * next step <param0> and \"value\"\n  ")

	c.Assert(parseRes.Error, IsNil)
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
	c.Assert(thirdConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(thirdConcept.ConceptSteps[0].Args[1].ArgType, Equals, Static)

	c.Assert(len(thirdConcept.Lookup.paramValue), Equals, 1)
	c.Assert(thirdConcept.Lookup.containsArg("param0"), Equals, true)

}

func (s *MySuite) TestParsingConceptStepWithInlineTable(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept <foo> \n * first step with <foo> and inline table\n |id|name|\n|1|vishnu|\n|2|prateek|\n")

	c.Assert(parseRes.Error, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.IsConcept, Equals, true)
	c.Assert(len(concept.ConceptSteps), Equals, 1)
	c.Assert(concept.ConceptSteps[0].Value, Equals, "first step with {} and inline table {}")

	tableArgument := concept.ConceptSteps[0].Args[1]
	c.Assert(tableArgument.ArgType, Equals, TableArg)

	inlineTable := tableArgument.Table
	c.Assert(inlineTable.IsInitialized(), Equals, true)
	c.Assert(len(inlineTable.Get("id")), Equals, 2)
	c.Assert(len(inlineTable.Get("name")), Equals, 2)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "vishnu")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "prateek")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, Static)
}

func (s *MySuite) TestErrorParsingConceptWithInvalidInlineTable(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept \n |id|name|\n|1|vishnu|\n|2|prateek|\n")

	c.Assert(parseRes.Error, NotNil)
	c.Assert(parseRes.Error.Message, Equals, "Table doesn't belong to any step")
}

func (s *MySuite) TestDeepCopyOfConcept(c *C) {
	dictionary := new(ConceptDictionary)
	normalStep1 := &Step{Value: "normal step 1", LineText: "normal step 1"}
	normalStep2 := &Step{Value: "normal step 2", LineText: "normal step 2"}

	nestedConceptStep := &Step{Value: "nested concept", LineText: "nested concept"}

	topLevelConcept := &Step{Value: "top level concept", IsConcept: true, ConceptSteps: []*Step{nestedConceptStep, normalStep1}}
	nestedConcept := &Step{Value: "nested concept", LineText: "nested concept", IsConcept: true, ConceptSteps: []*Step{normalStep2}}

	dictionary.Add([]*Step{topLevelConcept}, "file1.cpt")
	dictionary.Add([]*Step{nestedConcept}, "file2.cpt")

	actualConcept := dictionary.search("top level concept")

	copiedTopLevelConcept := actualConcept.deepCopy()

	verifyCopiedConcept(copiedTopLevelConcept, actualConcept, c)
}

func verifyCopiedConcept(copiedConcept *Concept, actualConcept *Concept, c *C) {
	c.Assert(&copiedConcept, Not(Equals), &actualConcept)
	c.Assert(copiedConcept, DeepEquals, actualConcept)
}

func (s *MySuite) TestNestedConceptLooksUpArgsFromParent(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)
	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First flow").
		step("create user \"foo\" \"doo\"").
		step("another step").String()

	conceptText := SpecBuilder().
		specHeading("create user <bar> <far>").
		step("assign role <bar> <far>").
		step("step 2").
		specHeading("assign role <baz> <boo>").
		step("add admin rights <baz>").
		step("give root access").String()

	concepts, _ := new(ConceptParser).parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	tokens, _ := parser.GenerateTokens(specText)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstStepInSpec := spec.Scenarios[0].Steps[0]
	nestedConcept := firstStepInSpec.ConceptSteps[0]
	nestedConceptArg1 := nestedConcept.getArg("baz")
	c.Assert(nestedConceptArg1.Value, Equals, "foo")
	nestedConceptArg2 := nestedConcept.getArg("boo")
	c.Assert(nestedConceptArg2.Value, Equals, "doo")
}

func (s *MySuite) TestNestedConceptLooksUpArgsFromParentPresentWhenNestedConceptDefinedFirst(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"foo\" \"prateek\" and \"007\"").
		String()

	conceptText := SpecBuilder().
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name \"static-name\"").String()

	concepts, _ := new(ConceptParser).parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	tokens, _ := parser.GenerateTokens(specText)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstLevelConcept := spec.Scenarios[0].Steps[0]
	c.Assert(firstLevelConcept.getArg("user-id").Value, Equals, "foo")
	c.Assert(firstLevelConcept.getArg("user-name").Value, Equals, "prateek")
	c.Assert(firstLevelConcept.getArg("user-phone").Value, Equals, "007")

	nestedConcept := firstLevelConcept.ConceptSteps[0]

	c.Assert(nestedConcept.getArg("userid").Value, Equals, "foo")
	c.Assert(nestedConcept.getArg("username").Value, Equals, "static-name")

}

func (s *MySuite) TestNestedConceptLooksUpArgsFromParentPresentWhenNestedConceptDefinedSecond(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"foo\" \"prateek\" and \"007\"").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name \"static-name\"").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(ConceptParser).parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	tokens, _ := parser.GenerateTokens(specText)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstLevelConcept := spec.Scenarios[0].Steps[0]
	c.Assert(firstLevelConcept.getArg("user-id").Value, Equals, "foo")
	c.Assert(firstLevelConcept.getArg("user-name").Value, Equals, "prateek")
	c.Assert(firstLevelConcept.getArg("user-phone").Value, Equals, "007")

	nestedConcept := firstLevelConcept.ConceptSteps[0]
	c.Assert(nestedConcept.getArg("userid").Value, Equals, "foo")
	c.Assert(nestedConcept.getArg("username").Value, Equals, "static-name")

}

func (s *MySuite) TestNestedConceptLooksUpDataTableArgs(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)
	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		tableHeader("456", "apoorva", "9800").
		tableHeader("789", "srikanth", "7900").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		step("another step").String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign number <user-phone>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(ConceptParser).parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	tokens, _ := parser.GenerateTokens(specText)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)

	firstStepInSpec := spec.Scenarios[0].Steps[0]
	c.Assert(firstStepInSpec.IsConcept, Equals, true)
	c.Assert(firstStepInSpec.getArg("user-id").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-name").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-phone").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-id").Value, Equals, "id")
	c.Assert(firstStepInSpec.getArg("user-name").Value, Equals, "name")
	c.Assert(firstStepInSpec.getArg("user-phone").Value, Equals, "phone")

	nestedConcept := firstStepInSpec.ConceptSteps[0]
	c.Assert(nestedConcept.getArg("userid").ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.getArg("username").ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.getArg("userid").Value, Equals, "id")
	c.Assert(nestedConcept.getArg("username").Value, Equals, "name")

}

func (s *MySuite) TestNestedConceptLooksUpWhenParameterPlaceholdersAreSame(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)
	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		tableHeader("456", "apoorva", "9800").
		tableHeader("789", "srikanth", "7900").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		step("another step").String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign number <user-phone>").
		specHeading("assign id <user-id> and name <user-name>").
		step("add id <user-id>").
		step("add name <user-name>").String()

	concepts, _ := new(ConceptParser).parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	tokens, _ := parser.GenerateTokens(specText)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)

	firstStepInSpec := spec.Scenarios[0].Steps[0]
	c.Assert(firstStepInSpec.IsConcept, Equals, true)
	c.Assert(firstStepInSpec.getArg("user-id").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-name").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-phone").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-id").Value, Equals, "id")
	c.Assert(firstStepInSpec.getArg("user-name").Value, Equals, "name")
	c.Assert(firstStepInSpec.getArg("user-phone").Value, Equals, "phone")

	nestedConcept := firstStepInSpec.ConceptSteps[0]
	c.Assert(nestedConcept.getArg("user-id").ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.getArg("user-name").ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.getArg("user-id").Value, Equals, "id")
	c.Assert(nestedConcept.getArg("user-name").Value, Equals, "name")

}

func (s *MySuite) TestErrorOnCircularReferenceInConcept(c *C) {
	conceptDictionary := new(ConceptDictionary)

	conceptText := SpecBuilder().
		specHeading("another concept").
		step("second step").
		step("my concept").
		specHeading("my concept").
		step("first step").
		step("another concept").String()

	concepts, parseErr := new(ConceptParser).parse(conceptText)
	c.Assert(parseErr.Error, IsNil)
	err := conceptDictionary.Add(concepts, "file.cpt")
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.Contains(err.Message, "Circular reference found in concept"))
}

func (s *MySuite) TestErrorOnCircularReferenceInDeepNestedConceptConcept(c *C) {
	conceptDictionary := new(ConceptDictionary)
	conceptText := SpecBuilder().
		specHeading("first concept <a> and <b>").
		step("a step step").
		step("a nested concept <a>").
		specHeading("a nested concept <b>").
		step("second nested <b>").
		step("another step").String()

	secondConceptText := SpecBuilder().
		specHeading("second nested <c>").
		step("a nested concept <c>").String()

	concepts1, parseRes := new(ConceptParser).parse(conceptText)
	c.Assert(parseRes.Error, IsNil)
	concepts2, parseRes := new(ConceptParser).parse(secondConceptText)

	err := conceptDictionary.Add(concepts1, "file.cpt")
	c.Assert(err, IsNil)

	err = conceptDictionary.Add(concepts2, "file2.cpt")
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.Contains(err.Message, "Circular reference found in concept"))
}

func (s *MySuite) TestConceptHavingDynamicParameters(c *C) {
	conceptText := SpecBuilder().
		specHeading("create user <user:id> <user:name> and <file>").
		step("a step <user:id>").String()
	step, _ := new(ConceptParser).parse(conceptText)
	c.Assert(step[0].LineText, Equals, "create user <user:id> <user:name> and <file>")
	c.Assert(step[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(step[0].Args[1].ArgType, Equals, Dynamic)
	c.Assert(step[0].Args[2].ArgType, Equals, Dynamic)
}

func (s *MySuite) TestConceptHavingInvalidSpecialParameters(c *C) {
	conceptText := SpecBuilder().
		specHeading("create user <user:id> <table:name> and <file>").
		step("a step <user:id>").String()
	_, parseRes := new(ConceptParser).parse(conceptText)
	c.Assert(parseRes.Error.Message, Equals, "Dynamic parameter <table:name> could not be resolved")
}

func (s *MySuite) TestConceptHavingStaticParameters(c *C) {
	conceptText := SpecBuilder().
		specHeading("create user <user:id> \"abc\" and <file>").
		step("a step <user:id>").String()
	_, parseRes := new(ConceptParser).parse(conceptText)
	c.Assert(parseRes.Error.Message, Equals, "Concept heading can have only Dynamic Parameters")
}
