/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"path/filepath"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParsingFileSpecialType(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*gauge.StepArg, error) {
		return &gauge.StepArg{Value: "dummy", ArgType: gauge.Static}, nil
	}

	stepArg, _ := resolver.resolve("file:foo")
	c.Assert(stepArg.Value, Equals, "dummy")
	c.Assert(stepArg.ArgType, Equals, gauge.Static)
	c.Assert(stepArg.Name, Equals, "file:foo")
}

func (s *MySuite) TestParsingFileAsSpecialParamWithWindowsPathAsValue(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*gauge.StepArg, error) {
		return &gauge.StepArg{Value: "hello", ArgType: gauge.SpecialString}, nil
	}

	stepArg, _ := resolver.resolve(`file:C:\Users\abc`)
	c.Assert(stepArg.Value, Equals, "hello")
	c.Assert(stepArg.ArgType, Equals, gauge.SpecialString)
	if util.IsWindows() {
		c.Assert(stepArg.Name, Equals, `file:C:\\Users\\abc`)
	} else {
		c.Assert(stepArg.Name, Equals, `file:C:\Users\abc`)
	}
}

func (s *MySuite) TestParsingInvalidSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.resolve("unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestConvertCsvToTable(c *C) {
	table, _ := convertCsvToTable("id,name\n1,foo\n2,bar")

	idColumn, _ := table.Get("id")
	c.Assert(idColumn[0].Value, Equals, "1")
	c.Assert(idColumn[1].Value, Equals, "2")

	nameColumn, _ := table.Get("name")
	c.Assert(nameColumn[0].Value, Equals, "foo")
	c.Assert(nameColumn[1].Value, Equals, "bar")
}

func (s *MySuite) TestConvertEmptyCsvToTable(c *C) {
	table, _ := convertCsvToTable("")
	c.Assert(len(table.Columns), Equals, 0)
}

func (s *MySuite) TestParsingUnknownSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.getStepArg("unknown", "foo", "unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestPopulatingConceptLookup(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "888").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	spec, _, _ := parser.Parse(specText, conceptDictionary, "")
	concept := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup)
	err = dataTableLookup.ReadDataTableRow(spec.DataTable.Table, 0)
	c.Assert(err, IsNil)
	err = PopulateConceptDynamicParams(concept, dataTableLookup)
	c.Assert(err, IsNil)
	useridArg, _ := concept.GetArg("user-id")
	c.Assert(useridArg.Value, Equals, "123")
	usernameArg, _ := concept.GetArg("user-name")
	c.Assert(usernameArg.Value, Equals, "foo")
	userphoneArg, _ := concept.GetArg("user-phone")
	c.Assert(userphoneArg.Value, Equals, "888")

}

func (s *MySuite) TestPopulatingNestedConceptLookup(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	spec, _, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup)
	err = dataTableLookup.ReadDataTableRow(spec.DataTable.Table, 0)
	c.Assert(err, IsNil)
	err = PopulateConceptDynamicParams(concept1, dataTableLookup)
	c.Assert(err, IsNil)

	useridArg1, _ := concept1.GetArg("user-id")
	c.Assert(useridArg1.Value, Equals, "123")
	usernameArg1, _ := concept1.GetArg("user-name")
	c.Assert(usernameArg1.Value, Equals, "prateek")
	userphoneArg1, _ := concept1.GetArg("user-phone")
	c.Assert(userphoneArg1.Value, Equals, "8800")

	nestedConcept := concept1.ConceptSteps[0]
	useridArgN1, _ := nestedConcept.GetArg("userid")
	c.Assert(useridArgN1.Value, Equals, "123")
	usernameArgN1, _ := nestedConcept.GetArg("username")
	c.Assert(usernameArgN1.Value, Equals, "prateek")

	concept2 := spec.Scenarios[0].Steps[1]
	useridArg2, _ := concept2.GetArg("user-id")
	c.Assert(useridArg2.Value, Equals, "456")
	usernameArg2, _ := concept2.GetArg("user-name")
	c.Assert(usernameArg2.Value, Equals, "foo")
	userphoneArg2, _ := concept2.GetArg("user-phone")
	c.Assert(userphoneArg2.Value, Equals, "9900")

	nestedConcept2 := concept2.ConceptSteps[0]
	useridArgN2, _ := nestedConcept2.GetArg("userid")
	c.Assert(useridArgN2.Value, Equals, "456")
	usernameArgN2, _ := nestedConcept2.GetArg("username")
	c.Assert(usernameArgN2.Value, Equals, "foo")
}

func (s *MySuite) TestPopulatingNestedConceptsWithStaticParametersLookup(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"123456\"").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "static_param_concept.cpt"))
	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)

	spec, _, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup)
	err = dataTableLookup.ReadDataTableRow(spec.DataTable.Table, 0)
	c.Assert(err, IsNil)
	err = PopulateConceptDynamicParams(concept1, dataTableLookup)
	c.Assert(err, IsNil)
	useridArg1, _ := concept1.GetArg("user-id")
	c.Assert(useridArg1.Value, Equals, "456")
	usernameArg1, _ := concept1.GetArg("user-name")
	c.Assert(usernameArg1.Value, Equals, "foo")
	userphoneArg1, _ := concept1.GetArg("user-phone")
	c.Assert(userphoneArg1.Value, Equals, "123456")

	nestedConcept := concept1.ConceptSteps[0]
	useridArgN, _ := nestedConcept.GetArg("userid")
	c.Assert(useridArgN.Value, Equals, "456")
	usernameArgN, _ := nestedConcept.GetArg("username")
	c.Assert(usernameArgN.Value, Equals, "static-value")
}

func (s *MySuite) TestPopulatingConceptsWithDynamicParametersInTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("property").
		tableRow("something").
		scenarioHeading("First scenario").
		step("create user \"someone\" with ").
		tableHeader("name").
		tableRow("<property>").
		String()
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "table_param_concept.cpt"))
	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)

	spec, _, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup)
	err = dataTableLookup.ReadDataTableRow(spec.DataTable.Table, 0)
	c.Assert(err, IsNil)
	err = PopulateConceptDynamicParams(concept1, dataTableLookup)
	c.Assert(err, IsNil)
	tableArg, err := concept1.Lookup.GetArg("addresses")
	c.Assert(err, IsNil)
	v, err := tableArg.Table.Get("name")
	c.Assert(err, IsNil)
	c.Assert(v[0].CellType, Equals, gauge.Static)
	c.Assert(v[0].Value, Equals, "something")
}

func (s *MySuite) TestEachConceptUsageIsUpdatedWithRespectiveParams(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"sdf\" \"name\" and \"1234\"").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "static_param_concept.cpt"))
	_, _, err := AddConcepts([]string{path}, conceptDictionary)
	c.Assert(err, IsNil)
	spec, _, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	nestedConcept := concept1.ConceptSteps[0]
	nestedConcept1 := concept1.ConceptSteps[1]

	usernameArg, _ := nestedConcept.GetArg("username")
	c.Assert(usernameArg.Value, Equals, "static-value")
	usernameArg1, _ := nestedConcept1.GetArg("username")
	c.Assert(usernameArg1.Value, Equals, "static-value1")
	useridArg, _ := nestedConcept.GetArg("userid")
	c.Assert(useridArg.Value, Equals, "sdf")
	useridArg1, _ := nestedConcept1.GetArg("userid")
	c.Assert(useridArg1.Value, Equals, "sdf")
}

func (s *MySuite) TestGetResolveParameterFromTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec Heading").scenarioHeading("First scenario").step("my step").text("|name|id|").text("|---|---|").text("|john|123|").text("|james|<file:testdata/foo.txt>|").String()

	specs, _ := parser.ParseSpecText(specText, "")

	step := specs.Steps()[0]

	parameters, err := getResolvedParams(step, nil, nil)

	c.Assert(len(parameters), Equals, 1)
	c.Assert(parameters[0].Table.Rows[0].GetCells()[0], Equals, "john")
	c.Assert(parameters[0].Table.Rows[0].GetCells()[1], Equals, "123")
	c.Assert(parameters[0].Table.Rows[1].GetCells()[0], Equals, "james")
	c.Assert(parameters[0].Table.Rows[1].GetCells()[1], Equals, "007")

	c.Assert(err, IsNil)
}

func (s *MySuite) TestGetResolveParameterFromDataTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec Heading").text("|name|id|").text("|---|---|").text("|john|123|").text("|james|<file:testdata/foo.txt>|").scenarioHeading("First scenario").step("my step <id>").String()
	spec, _ := parser.ParseSpecText(specText, "")

	GetResolvedDataTablerows(spec.DataTable.Table)

	c.Assert(spec.DataTable.Table.Columns[0][0].Value, Equals, "john")
	c.Assert(spec.DataTable.Table.Columns[0][1].Value, Equals, "james")
	c.Assert(spec.DataTable.Table.Columns[1][0].Value, Equals, "123")
	c.Assert(spec.DataTable.Table.Columns[1][1].Value, Equals, "007")
}
