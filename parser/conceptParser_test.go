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
	step1 := &Step{value: "test step 1"}
	step2 := &Step{value: "test step 2"}

	err := dictionary.add([]*Step{step1, step2}, "file.cpt")

	c.Assert(err, IsNil)
	c.Assert(dictionary.conceptsMap["test step 1"].ConceptStep, Equals, step1)
	c.Assert(dictionary.conceptsMap["test step 1"].FileName, Equals, "file.cpt")
	c.Assert(dictionary.conceptsMap["test step 2"].ConceptStep, Equals, step2)
	c.Assert(dictionary.conceptsMap["test step 2"].FileName, Equals, "file.cpt")
}

func (s *MySuite) TestConceptDictionaryAddDuplicateConcept(c *C) {
	dictionary := new(ConceptDictionary)
	step1 := &Step{value: "test step {}", lineText: "test step <first>"}
	step2 := &Step{value: "test step {}", lineText: "test step <second>"}

	err := dictionary.add([]*Step{step1, step2}, "file.cpt")

	c.Assert(err, NotNil)
	c.Assert(err.Message, Equals, "Duplicate concept definition found")
}

func (s *MySuite) TestConceptDictionaryWithNestedConcepts(c *C) {
	dictionary := new(ConceptDictionary)
	normalStep1 := &Step{value: "normal step 1", lineText: "normal step 1"}
	normalStep2 := &Step{value: "normal step 2", lineText: "normal step 2"}
	nestedConceptStep := &Step{value: "nested concept", lineText: "nested concept"}
	nestedConcept := &Step{value: "nested concept", lineText: "nested concept", isConcept: true, conceptSteps: []*Step{normalStep2}}

	topLevelConcept := &Step{value: "top level concept", isConcept: true, conceptSteps: []*Step{nestedConceptStep, normalStep1}}

	dictionary.add([]*Step{nestedConcept}, "file1.cpt")
	dictionary.add([]*Step{topLevelConcept}, "file2.cpt")

	concept := dictionary.search("top level concept")
	c.Assert(len(concept.ConceptStep.conceptSteps), Equals, 2)
	actualnestedConcept := concept.ConceptStep.conceptSteps[0]
	c.Assert(actualnestedConcept.isConcept, Equals, true)
	c.Assert(len(actualnestedConcept.conceptSteps), Equals, 1)
	c.Assert(actualnestedConcept.conceptSteps[0].value, Equals, normalStep2.value)
	c.Assert(concept.ConceptStep.conceptSteps[1].value, Equals, normalStep1.value)
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithParameter(c *C) {
	conceptDictionary := new(conceptDictionary)
	conceptText := SpecBuilder().
		specHeading("assign id").
		step("add id").
		specHeading("create user").
		step("assign id").
		step("assign id").String()
	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	concept := conceptDictionary.search("create user")
	c.Assert(concept.conceptStep.conceptSteps[0].value, Equals, "assign id")
	c.Assert(concept.conceptStep.conceptSteps[0].isConcept, Equals, true)
	c.Assert(concept.conceptStep.conceptSteps[1].value, Equals, "assign id")
	c.Assert(concept.conceptStep.conceptSteps[1].isConcept, Equals, true)
}

func (s *MySuite) TestConceptDictionaryWithNestedConceptsWithDefinitionAfterUsage(c *C) {
	conceptDictionary := new(conceptDictionary)
	conceptText := SpecBuilder().
		specHeading("create user").
		step("assign id").
		step("assign id").
		step("assign id").
		specHeading("assign id").
		step("add id").String()
	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	concept := conceptDictionary.search("create user")
	c.Assert(concept.conceptStep.conceptSteps[0].value, Equals, "assign id")
	c.Assert(concept.conceptStep.conceptSteps[1].value, Equals, "assign id")
	c.Assert(concept.conceptStep.conceptSteps[2].value, Equals, "assign id")
	c.Assert(concept.conceptStep.conceptSteps[0].isConcept, Equals, true)
	c.Assert(concept.conceptStep.conceptSteps[1].isConcept, Equals, true)
	c.Assert(concept.conceptStep.conceptSteps[2].isConcept, Equals, true)
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
	conceptDictionary.add(concepts, "file.cpt")

	concept := conceptDictionary.search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.conceptSteps), Equals, 1)
	actualnestedConcept := concept.ConceptStep.conceptSteps[0]
	c.Assert(actualnestedConcept.isConcept, Equals, true)

	c.Assert(len(actualnestedConcept.conceptSteps), Equals, 2)
	c.Assert(actualnestedConcept.conceptSteps[0].value, Equals, "add id {}")
	c.Assert(actualnestedConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.conceptSteps[0].args[0].Value, Equals, "userid")
	c.Assert(len(concepts[0].items), Equals, 2)

	c.Assert(actualnestedConcept.conceptSteps[1].value, Equals, "add name {}")
	c.Assert(actualnestedConcept.conceptSteps[1].args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.conceptSteps[1].args[0].Value, Equals, "username")
	c.Assert(len(concepts[1].items), Equals, 3)
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
	conceptDictionary.add(concepts, "file.cpt")

	concept := conceptDictionary.search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.conceptSteps), Equals, 1)
	actualNestedConcept := concept.ConceptStep.conceptSteps[0]
	c.Assert(actualNestedConcept.isConcept, Equals, true)

	c.Assert(actualNestedConcept.args[0].ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.args[0].Value, Equals, "user-id")

	c.Assert(actualNestedConcept.args[1].ArgType, Equals, Static)
	c.Assert(actualNestedConcept.args[1].Value, Equals, "static-value")
	c.Assert(actualNestedConcept.lookup.getArg("userid").Value, Equals, "user-id")
	c.Assert(actualNestedConcept.lookup.getArg("userid").ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.lookup.getArg("username").Value, Equals, "static-value")
	c.Assert(actualNestedConcept.lookup.getArg("username").ArgType, Equals, Static)

	c.Assert(len(actualNestedConcept.conceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.conceptSteps[0].value, Equals, "add id {}")
	c.Assert(actualNestedConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.conceptSteps[0].args[0].Value, Equals, "userid")
	c.Assert(len(concepts[0].items), Equals, 3)

	c.Assert(actualNestedConcept.conceptSteps[1].value, Equals, "add name {}")
	c.Assert(actualNestedConcept.conceptSteps[1].args[0].ArgType, Equals, Dynamic)
	c.Assert(actualNestedConcept.conceptSteps[1].args[0].Value, Equals, "username")
	c.Assert(len(concepts[1].items), Equals, 2)
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
	conceptDictionary.add(concepts, "file.cpt")

	concept := conceptDictionary.search("create user {} {} and {}")
	c.Assert(len(concept.ConceptStep.conceptSteps), Equals, 1)
	actualnestedConcept := concept.ConceptStep.conceptSteps[0]
	c.Assert(actualnestedConcept.isConcept, Equals, true)

	c.Assert(len(actualnestedConcept.conceptSteps), Equals, 2)
	c.Assert(actualnestedConcept.conceptSteps[0].value, Equals, "add id {}")
	c.Assert(actualnestedConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.conceptSteps[0].args[0].Value, Equals, "userid")
	c.Assert(len(concepts[0].items), Equals, 3)
	c.Assert(concepts[0].items[2].(*Comment).value, Equals, "Comments")

	c.Assert(actualnestedConcept.conceptSteps[1].value, Equals, "add name {}")
	c.Assert(actualnestedConcept.conceptSteps[1].args[0].ArgType, Equals, Dynamic)
	c.Assert(actualnestedConcept.conceptSteps[1].args[0].Value, Equals, "username")
	c.Assert(len(concepts[1].items), Equals, 4)
}

func (s *MySuite) TestConceptHavingItemsWithTablesAndPreComments(c *C) {
	conceptDictionary := new(ConceptDictionary)
	concepts, _ := new(ConceptParser).parse("COMMENT\n# my concept <foo> \n * first step with <foo> and inline table\n |id|name|\n|1|vishnu|\n|2|prateek|\n comment")
	conceptDictionary.add(concepts, "concept.cpt")

	c.Assert(len(concepts[0].items), Equals, 3)
	c.Assert(len(concepts[0].preComments), Equals, 1)
	c.Assert(concepts[0].preComments[0].value, Equals, "COMMENT")
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
	normalStep1 := &Step{value: "normal step 1", lineText: "normal step 1"}
	normalStep2 := &Step{value: "normal step 2", lineText: "normal step 2"}

	nestedConceptStep := &Step{value: "nested concept", lineText: "nested concept"}

	topLevelConcept := &Step{value: "top level concept", isConcept: true, conceptSteps: []*Step{nestedConceptStep, normalStep1}}
	anotherTopLevelConcept := &Step{value: "top level concept 2", isConcept: true, conceptSteps: []*Step{nestedConceptStep}}
	nestedConcept := &Step{value: "nested concept", lineText: "nested concept", isConcept: true, conceptSteps: []*Step{normalStep2}}

	dictionary.add([]*Step{topLevelConcept}, "file1.cpt")
	dictionary.add([]*Step{anotherTopLevelConcept}, "file1.cpt")
	dictionary.add([]*Step{nestedConcept}, "file2.cpt")

	concept := dictionary.search("top level concept")
	c.Assert(len(concept.ConceptStep.conceptSteps), Equals, 2)
	actualNestedConcept := concept.ConceptStep.conceptSteps[0]
	c.Assert(actualNestedConcept.isConcept, Equals, true)
	c.Assert(len(actualNestedConcept.conceptSteps), Equals, 1)
	c.Assert(actualNestedConcept.conceptSteps[0].value, Equals, normalStep2.value)
	c.Assert(concept.ConceptStep.conceptSteps[1].value, Equals, normalStep1.value)

	topLevelConcept2 := dictionary.search("top level concept 2")
	c.Assert(len(topLevelConcept2.ConceptStep.conceptSteps), Equals, 1)
	actualNestedConcept = topLevelConcept2.ConceptStep.conceptSteps[0]
	c.Assert(actualNestedConcept.isConcept, Equals, true)
	c.Assert(len(actualNestedConcept.conceptSteps), Equals, 1)
	c.Assert(actualNestedConcept.conceptSteps[0].value, Equals, normalStep2.value)
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
	normalStep1 := &Step{value: "normal step 1", lineText: "normal step 1"}
	normalStep2 := &Step{value: "normal step 2", lineText: "normal step 2"}
	normalStep3 := &Step{value: "normal step 3", lineText: "normal step 3"}
	nestedConceptStep := &Step{value: "nested concept", lineText: "nested concept"}

	topLevelConcept := &Step{value: "top level concept", isConcept: true, conceptSteps: []*Step{nestedConceptStep, normalStep1}}
	anotherNestedConcept := &Step{value: "another nested concept", isConcept: true, conceptSteps: []*Step{normalStep3}}
	nestedConcept := &Step{value: "nested concept", isConcept: true, conceptSteps: []*Step{anotherNestedConcept, normalStep2}}

	dictionary.add([]*Step{topLevelConcept}, "file1.cpt")
	dictionary.add([]*Step{anotherNestedConcept}, "file1.cpt")
	dictionary.add([]*Step{nestedConcept}, "file1.cpt")

	actualTopLevelConcept := dictionary.search("top level concept")
	c.Assert(len(actualTopLevelConcept.ConceptStep.conceptSteps), Equals, 2)
	actualNestedConcept := actualTopLevelConcept.ConceptStep.conceptSteps[0]
	c.Assert(actualNestedConcept.isConcept, Equals, true)
	c.Assert(len(actualNestedConcept.conceptSteps), Equals, 2)
	c.Assert(actualNestedConcept.conceptSteps[0].value, Equals, anotherNestedConcept.value)
	c.Assert(actualNestedConcept.conceptSteps[1].value, Equals, normalStep2.value)
	c.Assert(actualTopLevelConcept.ConceptStep.conceptSteps[1].value, Equals, normalStep1.value)

	actualAnotherNestedConcept := dictionary.search("another nested concept")
	c.Assert(len(actualAnotherNestedConcept.ConceptStep.conceptSteps), Equals, 1)
	step := actualAnotherNestedConcept.ConceptStep.conceptSteps[0]
	c.Assert(step.isConcept, Equals, false)
	c.Assert(step.value, Equals, normalStep3.value)

	nestedConcept2 := dictionary.search("nested concept")
	c.Assert(len(nestedConcept2.ConceptStep.conceptSteps), Equals, 2)
	actualAnotherNestedConcept2 := nestedConcept2.ConceptStep.conceptSteps[0]
	c.Assert(actualAnotherNestedConcept2.isConcept, Equals, true)
	c.Assert(len(actualAnotherNestedConcept2.conceptSteps), Equals, 1)
	c.Assert(actualAnotherNestedConcept2.conceptSteps[0].value, Equals, normalStep3.value)
	c.Assert(nestedConcept2.ConceptStep.conceptSteps[1].value, Equals, normalStep2.value)

}

func (s *MySuite) TestConceptDictionarySearch(c *C) {
	dictionary := new(ConceptDictionary)
	step1 := &Step{value: "test step 1"}
	step2 := &Step{value: "test step 2"}

	dictionary.add([]*Step{step1, step2}, "file.cpt")

	c.Assert(dictionary.search(step1.value).ConceptStep, Equals, step1)
	c.Assert(dictionary.search(step1.value).FileName, Equals, "file.cpt")
	c.Assert(dictionary.search(step2.value).ConceptStep, Equals, step2)
	c.Assert(dictionary.search(step2.value).FileName, Equals, "file.cpt")
}

func (s *MySuite) TestParsingSimpleConcept(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept \n * first step \n * second step ")

	c.Assert(parseRes.error, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.isConcept, Equals, true)
	c.Assert(len(concept.conceptSteps), Equals, 2)
	c.Assert(concept.conceptSteps[0].value, Equals, "first step")
	c.Assert(concept.conceptSteps[1].value, Equals, "second step")

}

func (s *MySuite) TestErrorParsingConceptHeadingWithStaticOrSpecialParameter(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept with \"paratemer\" \n * first step \n * second step ")
	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Concept heading can have only Dynamic Parameters")

	_, parseRes = parser.parse("# my concept with <table: foo> \n * first step \n * second step ")
	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Dynamic parameter <table: foo> could not be resolved")

}

func (s *MySuite) TestErrorParsingConceptWithoutHeading(c *C) {
	parser := new(ConceptParser)

	_, parseRes := parser.parse("* first step \n * second step ")

	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Step is not defined inside a concept heading")
}

func (s *MySuite) TestErrorParsingConceptWithoutSteps(c *C) {
	parser := new(ConceptParser)

	_, parseRes := parser.parse("# my concept with \n")

	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Concept should have atleast one step")
}

func (s *MySuite) TestParsingSimpleConceptWithParameters(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept with <param0> and <param1> \n * first step using <param0> \n * second step using \"value\" and <param1> ")

	c.Assert(parseRes.error, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]
	c.Assert(concept.isConcept, Equals, true)
	c.Assert(len(concept.conceptSteps), Equals, 2)
	c.Assert(len(concept.lookup.paramValue), Equals, 2)
	c.Assert(concept.lookup.containsArg("param0"), Equals, true)
	c.Assert(concept.lookup.containsArg("param1"), Equals, true)

	firstConcept := concept.conceptSteps[0]
	c.Assert(firstConcept.value, Equals, "first step using {}")
	c.Assert(len(firstConcept.args), Equals, 1)
	c.Assert(firstConcept.args[0].ArgType, Equals, Dynamic)
	c.Assert(firstConcept.args[0].Value, Equals, "param0")

	secondConcept := concept.conceptSteps[1]
	c.Assert(secondConcept.value, Equals, "second step using {} and {}")
	c.Assert(len(secondConcept.args), Equals, 2)
	c.Assert(secondConcept.args[0].ArgType, Equals, Static)
	c.Assert(secondConcept.args[0].Value, Equals, "value")
	c.Assert(secondConcept.args[1].ArgType, Equals, Dynamic)
	c.Assert(secondConcept.args[1].Value, Equals, "param1")

}

func (s *MySuite) TestErrorParsingConceptWithRecursiveCallToConcept(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept \n * first step using \n * my concept ")

	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Cyclic dependancy found. Step is calling concept again.")
}

func (s *MySuite) TestErrorParsingConceptStepWithInvalidParameters(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept with <param0> and <param1> \n * first step using <param3> \n * second step using \"value\" and <param1> ")

	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Dynamic parameter <param3> could not be resolved")
}

func (s *MySuite) TestParsingMultipleConcept(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept \n * first step \n * second step \n# my second concept \n* next step\n # my third concept <param0>\n * next step <param0> and \"value\"\n  ")

	c.Assert(parseRes.error, IsNil)
	c.Assert(len(concepts), Equals, 3)

	firstConcept := concepts[0]
	secondConcept := concepts[1]
	thirdConcept := concepts[2]

	c.Assert(firstConcept.isConcept, Equals, true)
	c.Assert(len(firstConcept.conceptSteps), Equals, 2)
	c.Assert(firstConcept.conceptSteps[0].value, Equals, "first step")
	c.Assert(firstConcept.conceptSteps[1].value, Equals, "second step")

	c.Assert(secondConcept.isConcept, Equals, true)
	c.Assert(len(secondConcept.conceptSteps), Equals, 1)
	c.Assert(secondConcept.conceptSteps[0].value, Equals, "next step")

	c.Assert(thirdConcept.isConcept, Equals, true)
	c.Assert(len(thirdConcept.conceptSteps), Equals, 1)
	c.Assert(thirdConcept.conceptSteps[0].value, Equals, "next step {} and {}")
	c.Assert(len(thirdConcept.conceptSteps[0].args), Equals, 2)
	c.Assert(thirdConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(thirdConcept.conceptSteps[0].args[1].ArgType, Equals, Static)

	c.Assert(len(thirdConcept.lookup.paramValue), Equals, 1)
	c.Assert(thirdConcept.lookup.containsArg("param0"), Equals, true)

}

func (s *MySuite) TestParsingConceptStepWithInlineTable(c *C) {
	parser := new(ConceptParser)
	concepts, parseRes := parser.parse("# my concept <foo> \n * first step with <foo> and inline table\n |id|name|\n|1|vishnu|\n|2|prateek|\n")

	c.Assert(parseRes.error, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.isConcept, Equals, true)
	c.Assert(len(concept.conceptSteps), Equals, 1)
	c.Assert(concept.conceptSteps[0].value, Equals, "first step with {} and inline table {}")

	tableArgument := concept.conceptSteps[0].args[1]
	c.Assert(tableArgument.ArgType, Equals, TableArg)

	inlineTable := tableArgument.Table
	c.Assert(inlineTable.isInitialized(), Equals, true)
	c.Assert(len(inlineTable.Get("id")), Equals, 2)
	c.Assert(len(inlineTable.Get("name")), Equals, 2)
	c.Assert(inlineTable.Get("id")[0].value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[0].value, Equals, "vishnu")
	c.Assert(inlineTable.Get("name")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[1].value, Equals, "prateek")
	c.Assert(inlineTable.Get("name")[1].cellType, Equals, Static)
}

func (s *MySuite) TestErrorParsingConceptWithInvalidInlineTable(c *C) {
	parser := new(ConceptParser)
	_, parseRes := parser.parse("# my concept \n |id|name|\n|1|vishnu|\n|2|prateek|\n")

	c.Assert(parseRes.error, NotNil)
	c.Assert(parseRes.error.Message, Equals, "Table doesn't belong to any step")
}

func (s *MySuite) TestDeepCopyOfConcept(c *C) {
	dictionary := new(ConceptDictionary)
	normalStep1 := &Step{value: "normal step 1", lineText: "normal step 1"}
	normalStep2 := &Step{value: "normal step 2", lineText: "normal step 2"}

	nestedConceptStep := &Step{value: "nested concept", lineText: "nested concept"}

	topLevelConcept := &Step{value: "top level concept", isConcept: true, conceptSteps: []*Step{nestedConceptStep, normalStep1}}
	nestedConcept := &Step{value: "nested concept", lineText: "nested concept", isConcept: true, conceptSteps: []*Step{normalStep2}}

	dictionary.add([]*Step{topLevelConcept}, "file1.cpt")
	dictionary.add([]*Step{nestedConcept}, "file2.cpt")

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

	conceptDictionary.add(concepts, "file.cpt")
	tokens, _ := parser.generateTokens(specText)
	spec, parseResult := parser.createSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstStepInSpec := spec.scenarios[0].steps[0]
	nestedConcept := firstStepInSpec.conceptSteps[0]
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

	conceptDictionary.add(concepts, "file.cpt")
	tokens, _ := parser.generateTokens(specText)
	spec, parseResult := parser.createSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstLevelConcept := spec.scenarios[0].steps[0]
	c.Assert(firstLevelConcept.getArg("user-id").Value, Equals, "foo")
	c.Assert(firstLevelConcept.getArg("user-name").Value, Equals, "prateek")
	c.Assert(firstLevelConcept.getArg("user-phone").Value, Equals, "007")

	nestedConcept := firstLevelConcept.conceptSteps[0]

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

	conceptDictionary.add(concepts, "file.cpt")
	tokens, _ := parser.generateTokens(specText)
	spec, parseResult := parser.createSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstLevelConcept := spec.scenarios[0].steps[0]
	c.Assert(firstLevelConcept.getArg("user-id").Value, Equals, "foo")
	c.Assert(firstLevelConcept.getArg("user-name").Value, Equals, "prateek")
	c.Assert(firstLevelConcept.getArg("user-phone").Value, Equals, "007")

	nestedConcept := firstLevelConcept.conceptSteps[0]
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

	conceptDictionary.add(concepts, "file.cpt")
	tokens, _ := parser.generateTokens(specText)
	spec, parseResult := parser.createSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)

	firstStepInSpec := spec.scenarios[0].steps[0]
	c.Assert(firstStepInSpec.isConcept, Equals, true)
	c.Assert(firstStepInSpec.getArg("user-id").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-name").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-phone").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-id").Value, Equals, "id")
	c.Assert(firstStepInSpec.getArg("user-name").Value, Equals, "name")
	c.Assert(firstStepInSpec.getArg("user-phone").Value, Equals, "phone")

	nestedConcept := firstStepInSpec.conceptSteps[0]
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

	conceptDictionary.add(concepts, "file.cpt")
	tokens, _ := parser.generateTokens(specText)
	spec, parseResult := parser.createSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)

	firstStepInSpec := spec.scenarios[0].steps[0]
	c.Assert(firstStepInSpec.isConcept, Equals, true)
	c.Assert(firstStepInSpec.getArg("user-id").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-name").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-phone").ArgType, Equals, Dynamic)
	c.Assert(firstStepInSpec.getArg("user-id").Value, Equals, "id")
	c.Assert(firstStepInSpec.getArg("user-name").Value, Equals, "name")
	c.Assert(firstStepInSpec.getArg("user-phone").Value, Equals, "phone")

	nestedConcept := firstStepInSpec.conceptSteps[0]
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
	c.Assert(parseErr.error, IsNil)
	err := conceptDictionary.add(concepts, "file.cpt")
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
	c.Assert(parseRes.error, IsNil)
	concepts2, parseRes := new(ConceptParser).parse(secondConceptText)

	err := conceptDictionary.add(concepts1, "file.cpt")
	c.Assert(err, IsNil)

	err = conceptDictionary.add(concepts2, "file2.cpt")
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.Contains(err.Message, "Circular reference found in concept"))
}

func (s *MySuite) TestConceptHavingDynamicParameters(c *C) {
	conceptText := SpecBuilder().
		specHeading("create user <user:id> <user:name> and <file>").
		step("a step <user:id>").String()
	step, _ := new(ConceptParser).parse(conceptText)
	c.Assert(step[0].lineText, Equals, "create user <user:id> <user:name> and <file>")
	c.Assert(step[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(step[0].args[1].ArgType, Equals, Dynamic)
	c.Assert(step[0].args[2].ArgType, Equals, Dynamic)
}

func (s *MySuite) TestConceptHavingInvalidSpecialParameters(c *C) {
	conceptText := SpecBuilder().
		specHeading("create user <user:id> <table:name> and <file>").
		step("a step <user:id>").String()
	_, parseRes := new(ConceptParser).parse(conceptText)
	c.Assert(parseRes.error.Message, Equals, "Dynamic parameter <table:name> could not be resolved")
}

func (s *MySuite) TestConceptHavingStaticParameters(c *C) {
	conceptText := SpecBuilder().
		specHeading("create user <user:id> \"abc\" and <file>").
		step("a step <user:id>").String()
	_, parseRes := new(ConceptParser).parse(conceptText)
	c.Assert(parseRes.error.Message, Equals, "Concept heading can have only Dynamic Parameters")
}
