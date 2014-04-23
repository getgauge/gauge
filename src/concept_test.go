package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestConceptDictionaryAdd(c *C) {
	dictionary := new(conceptDictionary)
	step1 := &step{value: "test step 1"}
	step2 := &step{value: "test step 2"}

	err := dictionary.add([]*step{step1, step2}, "file.cpt")

	c.Assert(err, IsNil)
	c.Assert(dictionary.conceptsMap["test step 1"].conceptStep, Equals, step1)
	c.Assert(dictionary.conceptsMap["test step 1"].fileName, Equals, "file.cpt")
	c.Assert(dictionary.conceptsMap["test step 2"].conceptStep, Equals, step2)
	c.Assert(dictionary.conceptsMap["test step 2"].fileName, Equals, "file.cpt")
}

func (s *MySuite) TestConceptDictionaryAddDuplicateConcept(c *C) {
	dictionary := new(conceptDictionary)
	step1 := &step{value: "test step {}", lineText: "test step <first>"}
	step2 := &step{value: "test step {}", lineText: "test step <second>"}

	err := dictionary.add([]*step{step1, step2}, "file.cpt")

	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Duplicate concept definition found")
}

func (s *MySuite) TestConceptDictionarySearch(c *C) {
	dictionary := new(conceptDictionary)
	step1 := &step{value: "test step 1"}
	step2 := &step{value: "test step 2"}

	dictionary.add([]*step{step1, step2}, "file.cpt")

	c.Assert(dictionary.search(step1.value).conceptStep, Equals, step1)
	c.Assert(dictionary.search(step1.value).fileName, Equals, "file.cpt")
	c.Assert(dictionary.search(step2.value).conceptStep, Equals, step2)
	c.Assert(dictionary.search(step2.value).fileName, Equals, "file.cpt")
}

func (s *MySuite) TestParsingSimpleConcept(c *C) {
	parser := new(conceptParser)
	concepts, err := parser.parse("# my concept \n * first step \n * second step ")

	c.Assert(err, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.isConcept, Equals, true)
	c.Assert(len(concept.conceptSteps), Equals, 2)
	c.Assert(concept.conceptSteps[0].value, Equals, "first step")
	c.Assert(concept.conceptSteps[1].value, Equals, "second step")

}

func (s *MySuite) TestErrorParsingConceptHeadingWithStaticOrSpecialParameter(c *C) {
	parser := new(conceptParser)
	_, err := parser.parse("# my concept with \"paratemer\" \n * first step \n * second step ")
	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Concept heading can have only Dynamic Parameters")

	_, err = parser.parse("# my concept with <table: foo> \n * first step \n * second step ")
	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Concept heading can have only Dynamic Parameters")

}

func (s *MySuite) TestErrorParsingConceptWithoutHeading(c *C) {
	parser := new(conceptParser)

	_, err := parser.parse("* first step \n * second step ")

	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Step is not defined inside a concept heading")
}

func (s *MySuite) TestErrorParsingConceptWithoutSteps(c *C) {
	parser := new(conceptParser)

	_, err := parser.parse("# my concept with \n")

	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Concept should have atleast one step")
}

func (s *MySuite) TestParsingSimpleConceptWithParameters(c *C) {
	parser := new(conceptParser)
	concepts, err := parser.parse("# my concept with <param0> and <param1> \n * first step using <param0> \n * second step using \"value\" and <param1> ")

	c.Assert(err, IsNil)
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
	c.Assert(firstConcept.args[0].argType, Equals, dynamic)
	c.Assert(firstConcept.args[0].value, Equals, "param0")

	secondConcept := concept.conceptSteps[1]
	c.Assert(secondConcept.value, Equals, "second step using {} and {}")
	c.Assert(len(secondConcept.args), Equals, 2)
	c.Assert(secondConcept.args[0].argType, Equals, static)
	c.Assert(secondConcept.args[0].value, Equals, "value")
	c.Assert(secondConcept.args[1].argType, Equals, dynamic)
	c.Assert(secondConcept.args[1].value, Equals, "param1")

}

func (s *MySuite) TestErrorParsingConceptStepWithInvalidParameters(c *C) {
	parser := new(conceptParser)
	_, err := parser.parse("# my concept with <param0> and <param1> \n * first step using <param3> \n * second step using \"value\" and <param1> ")

	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Dynamic parameter <param3> could not be resolved")
}

func (s *MySuite) TestParsingMultipleConcept(c *C) {
	parser := new(conceptParser)
	concepts, err := parser.parse("# my concept \n * first step \n * second step \n# my second concept \n* next step\n # my third concept <param0>\n * next step <param0> and \"value\"\n  ")

	c.Assert(err, IsNil)
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
	c.Assert(thirdConcept.conceptSteps[0].args[0].argType, Equals, dynamic)
	c.Assert(thirdConcept.conceptSteps[0].args[1].argType, Equals, static)

	c.Assert(len(thirdConcept.lookup.paramValue), Equals, 1)
	c.Assert(thirdConcept.lookup.containsArg("param0"), Equals, true)

}

func (s *MySuite) TestParsingConceptStepWithInlineTable(c *C) {
	parser := new(conceptParser)
	concepts, err := parser.parse("# my concept \n * first step with inline table\n |id|name|\n|1|vishnu|\n|2|prateek|\n")

	c.Assert(err, IsNil)
	c.Assert(len(concepts), Equals, 1)

	concept := concepts[0]

	c.Assert(concept.isConcept, Equals, true)
	c.Assert(len(concept.conceptSteps), Equals, 1)
	c.Assert(concept.conceptSteps[0].value, Equals, "first step with inline table")
	inlineTable := concept.conceptSteps[0].inlineTable
	c.Assert(inlineTable.isInitialized(), Equals, true)
	c.Assert(len(inlineTable.get("id")), Equals, 2)
	c.Assert(len(inlineTable.get("name")), Equals, 2)
	c.Assert(inlineTable.get("id")[0], Equals, "1")
	c.Assert(inlineTable.get("id")[1], Equals, "2")
	c.Assert(inlineTable.get("name")[0], Equals, "vishnu")
	c.Assert(inlineTable.get("name")[1], Equals, "prateek")
}

func (s *MySuite) TestErrorParsingConceptWithInvalidInlineTable(c *C) {
	parser := new(conceptParser)
	_, err := parser.parse("# my concept \n |id|name|\n|1|vishnu|\n|2|prateek|\n")

	c.Assert(err, NotNil)
	c.Assert(err.message, Equals, "Table doesn't belong to any step")
}
