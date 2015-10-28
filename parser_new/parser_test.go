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

package parse

import (
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge"
)

type conceptParseTest struct {
	name    string
	text    string
	concept gauge.Concept
}

var conceptParseTests = []conceptParseTest{
	{"simple concept", "# This is a concept heading\n* This is the first step\n* This is the second step\n",
		gauge.Concept{
			FileName: "simple concept",
			LineNo:   0,
			Heading:  "This is a concept heading",
			Steps: []gauge.Step{
				{LineNo: 0, ActualText: "This is the first step"},
				{LineNo: 0, ActualText: "This is the second step"},
			},
		},
	},
	{"simple underline concept", "This is a concept heading\n=======================\n* This is the first step\n* This is the second step",
		gauge.Concept{
			FileName: "simple underline concept",
			LineNo:   0,
			Heading:  "This is a concept heading",
			Steps: []gauge.Step{
				{LineNo: 0, ActualText: "This is the first step"},
				{LineNo: 0, ActualText: "This is the second step"},
			},
		},
	},
}

func TestConceptParsing(t *testing.T) {
	for _, test := range conceptParseTests {
		cpt := ParseConcept(test.name, test.text)
		if !reflect.DeepEqual(cpt, test.concept) {
			t.Errorf("%s: \ngot\n\t%+v\nexpected\n\t%v", test.name, cpt, test.concept)
		}
	}
}
