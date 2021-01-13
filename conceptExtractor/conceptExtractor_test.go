/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package conceptExtractor

import (
	"testing"

	"os"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {
	config.ProjectRoot, _ = os.Getwd()
}

func (s *MySuite) TestExtractConceptWithoutParameters(c *C) {
	STEP := "step that takes a table"
	name := "concept"
	conceptName := &gauge_messages.Step{Name: name}
	concept, conceptText, err := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP}}, "# sdfdsf\nsome comment\n* some step\n## sce\n* step", "")

	c.Assert(err, IsNil)
	c.Assert(concept, Equals, "# concept\n* step that takes a table\n")
	c.Assert(conceptText, Equals, "* concept")
}

func (s *MySuite) TestExtractConcept(c *C) {
	STEP := "step that takes a table \"arg\""
	name := "concept with \"arg\""
	conceptName := &gauge_messages.Step{Name: name}
	concept, conceptText, err := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP}}, "# sdfdsf\nsome comment\n* some step\n## sce\n* step", "")

	c.Assert(err, IsNil)
	c.Assert(concept, Equals, "# concept with <arg>\n* step that takes a table <arg>\n")
	c.Assert(conceptText, Equals, "* concept with \"arg\"")
}

func (s *MySuite) TestExtractConceptWithSkippedParameters(c *C) {
	STEP := "step that takes a table \"arg\" and \"hello again\" "
	name := "concept with \"arg\""
	conceptName := &gauge_messages.Step{Name: name}
	concept, conceptText, err := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP}}, "# sdfdsf\nsome comment\n* some step\n## sce\n* step", "")

	c.Assert(err, IsNil)
	c.Assert(concept, Equals, "# concept with <arg>\n* step that takes a table <arg> and \"hello again\"\n")
	c.Assert(conceptText, Equals, "* concept with \"arg\"")
}

func (s *MySuite) TestExtractConceptWithDynamicAndStaticParameters(c *C) {
	STEP := "step that takes a table \"arg\" and <hello again> "
	name := "concept with \"arg\" <hello again>"
	conceptName := &gauge_messages.Step{Name: name}
	concept, conceptText, err := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP}}, "# sdfdsf\n\n|hello again|name|\n|hey|hello|\n\n## sce\n* step", "")

	c.Assert(err, IsNil)
	c.Assert(concept, Equals, "# concept with <arg> <hello again>\n* step that takes a table <arg> and <hello again>\n")
	c.Assert(conceptText, Equals, "* concept with \"arg\" <hello again>")
}

func (s *MySuite) TestExtractConceptWithDynamicAndStaticParametersWithParamChar(c *C) {
	STEP := "step that takes a table \"arg <hello>\" and <hello again> "
	name := "concept with \"arg <hello>\" <hello again>"
	conceptName := &gauge_messages.Step{Name: name}
	concept, conceptText, err := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP}}, "# sdfdsf\n\n|hello again|name|\n|hey|hello|\n\n## sce\n* step", "")

	c.Assert(err, IsNil)
	c.Assert(concept, Equals, "# concept with <arg {hello}> <hello again>\n* step that takes a table <arg {hello}> and <hello again>\n")
	c.Assert(conceptText, Equals, "* concept with \"arg <hello>\" <hello again>")
}

func (s *MySuite) TestExtractConceptWithTableAsArg(c *C) {
	STEP := "step that takes a table"
	name := "concept with <table1>"
	conceptName := &gauge_messages.Step{Name: name}
	tableName := TABLE + "1"
	table := `	|id|name|
	|--|----|
	|1 |foo |
	|2 |bar |
	`
	concept, conceptText, err := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP, Table: table, ParamTableName: tableName},
		&gauge_messages.Step{Name: STEP, Table: table, ParamTableName: tableName}}, "# sdfdsf\nsome comment\n* some step\n## sce\n* step", "")

	c.Assert(err, IsNil)
	c.Assert(concept, Equals, "# concept with <table1>\n* step that takes a table <table1>\n* step that takes a table <table1>\n")
	c.Assert(conceptText, Equals, "* concept with "+`

   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |
`)
}

func (s *MySuite) TestExtractConceptWithTableAsArgAndTableWithDynamicArgs(c *C) {
	STEP := "step that takes a table"
	name := "concept with <table1>"
	conceptName := &gauge_messages.Step{Name: name}
	tableName := TABLE + "1"
	table := `	|id|name|
	|--|----|
	|1 |hello <foo> |
	|2 |bar |
	`
	concept, conceptText, _ := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP, Table: table, ParamTableName: tableName},
		&gauge_messages.Step{Name: STEP, Table: table, ParamTableName: tableName}}, "# sdfdsf\n\n|foo|name|\n|hey|hello|\n\n ##helloasdasdasd\n\n* step", "")

	c.Assert(concept, Equals, "# concept with <table1>\n* step that takes a table <table1>\n* step that takes a table <table1>\n")
	c.Assert(conceptText, Equals, "* concept with "+`

   |id|name       |
   |--|-----------|
   |1 |hello <foo>|
   |2 |bar        |
`)
}

func (s *MySuite) TestExtractConceptWithSkippedTableAsArg(c *C) {
	STEP := "step that takes a table"
	name := "concept with <table1>"
	conceptName := &gauge_messages.Step{Name: name}
	tableName := TABLE + "1"
	table := `	|id|name|
	|--|----|
	|1 |foo |
	|2 |bar |
	`
	concept, conceptText, _ := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP, Table: table, ParamTableName: tableName},
		&gauge_messages.Step{Name: STEP, Table: table, ParamTableName: tableName}, &gauge_messages.Step{Name: STEP, Table: table}}, "# sdfdsf\nsome comment\n* some step\n## sce\n* step", "")

	c.Assert(concept, Equals, "# concept with <table1>\n* step that takes a table <table1>\n* step that takes a table <table1>\n* step that takes a table"+`

   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |
`)
	c.Assert(conceptText, Equals, "* concept with "+`

   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |
`)
}

func (s *MySuite) TestExtractConceptWithTableWithDynamicArgs(c *C) {
	STEP := "step that takes a table"
	name := "concept with"
	conceptName := &gauge_messages.Step{Name: name}
	table := `|id|name|
	|--|----|
	|1 |<foo>|
	|2 |bar |
	`
	concept, conceptText, _ := getExtractedConcept(conceptName, []*gauge_messages.Step{&gauge_messages.Step{Name: STEP, Table: table}},
		"# sdfdsf\n\n|foo|name|\n|hey|hello|\n\n##helloasdasdasd\n\n* step", "")

	c.Assert(concept, Equals, "# concept with <foo>\n* step that takes a table"+`

   |id|name |
   |--|-----|
   |1 |<foo>|
   |2 |bar  |
`)
	c.Assert(conceptText, Equals, "* concept with <foo>")
}

func (s *MySuite) TestReplaceText(c *C) {
	content := `Copyright 2015 ThoughtWorks, Inc.

	This file is part of Gauge.

	Gauge is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Gauge is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Gauge.  If not, see <http://www.gnu.org/licenses/>.`

	replacement := `* concept with
     |id|name|
     |--|----|
     |1 |foo |
     |2 |bar |
`
	five := int32(5)
	ten := int32(10)
	info := &gauge_messages.TextInfo{StartingLineNo: five, EndLineNo: ten}
	finalText := replaceText(content, info, replacement)

	c.Assert(finalText, Equals, `Copyright 2015 ThoughtWorks, Inc.

	This file is part of Gauge.

* concept with
     |id|name|
     |--|----|
     |1 |foo |
     |2 |bar |

	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Gauge.  If not, see <http://www.gnu.org/licenses/>.`)
}

func (s *MySuite) TestIsDuplicateConcept(c *C) {
	concept := &gauge.Concept{ConceptStep: &gauge.Step{Value: "concept", IsConcept: true}, FileName: "sdfsdf.cpt"}
	dictionary := gauge.NewConceptDictionary()
	dictionary.ConceptsMap["sdfsdf.cpt"] = concept

	isDuplicate := isDuplicateConcept(&gauge.Step{Value: "concept"}, dictionary)

	c.Assert(isDuplicate, Equals, true)
}

func (s *MySuite) TestIsDuplicateConceptWithUniqueConcepts(c *C) {
	concept := &gauge.Concept{ConceptStep: &gauge.Step{Value: "concept", IsConcept: true}, FileName: "sdfsdf.cpt"}
	dictionary := gauge.NewConceptDictionary()
	dictionary.ConceptsMap["sdfsdf.cpt"] = concept

	isDuplicate := isDuplicateConcept(&gauge.Step{Value: "concept1"}, dictionary)

	c.Assert(isDuplicate, Equals, false)
}
