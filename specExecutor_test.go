package main

import . "launchpad.net/gocheck"

func (s *MySuite) TestResolveToProtoConceptItem(c *C) {
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign phone <user-phone>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil)
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()
	c.Assert(protoConcept.GetSteps()[0].GetItemType(), Equals, ProtoItem_Concept)

	nestedConcept := protoConcept.GetSteps()[0].GetConcept()
	firstNestedStep := nestedConcept.GetSteps()[0].GetStep()
	params := getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "456")

	secondNestedStep := nestedConcept.GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "foo")

	c.Assert(protoConcept.GetSteps()[1].GetItemType(), Equals, ProtoItem_Step)
	secondStepInConcept := protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStepInConcept.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "9900")

}

func (s *MySuite) TestResolveToProtoConceptItemWithDataTable(c *C) {
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "8800").
		tableHeader("456", "bar", "9800").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign phone <user-phone>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil)

	// For first row
	specExecutor.dataTableIndex = 0
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()
	c.Assert(protoConcept.GetSteps()[0].GetItemType(), Equals, ProtoItem_Concept)

	nestedConcept := protoConcept.GetSteps()[0].GetConcept()
	firstNestedStep := nestedConcept.GetSteps()[0].GetStep()
	params := getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "123")

	secondNestedStep := nestedConcept.GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "foo")

	c.Assert(protoConcept.GetSteps()[1].GetItemType(), Equals, ProtoItem_Step)
	secondStepInConcept := protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStepInConcept.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "8800")

	// For second row
	specExecutor.dataTableIndex = 1
	protoConcept = specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()
	c.Assert(protoConcept.GetSteps()[0].GetItemType(), Equals, ProtoItem_Concept)

	nestedConcept = protoConcept.GetSteps()[0].GetConcept()
	firstNestedStep = nestedConcept.GetSteps()[0].GetStep()
	params = getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "456")

	secondNestedStep = nestedConcept.GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "bar")

	c.Assert(protoConcept.GetSteps()[1].GetItemType(), Equals, ProtoItem_Step)
	secondStepInConcept = protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStepInConcept.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "9800")
}
