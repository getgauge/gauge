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

package filter

import (
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestToShuffleSpecsToRandomize(c *C) {
	var specs []*parser.Specification
	specs = append(specs, &parser.Specification{FileName: "a"}, &parser.Specification{FileName: "b"}, &parser.Specification{FileName: "c"}, &parser.Specification{FileName: "d"},
		&parser.Specification{FileName: "e"}, &parser.Specification{FileName: "f"}, &parser.Specification{FileName: "g"}, &parser.Specification{FileName: "h"})
	shuffledSpecs := shuffleSpecs(specs)
	for i, spec := range shuffledSpecs {
		if spec.FileName != specs[i].FileName {
			c.Succeed()
		}
	}
}

func (s *MySuite) TestToRunSpecificSetOfSpecs(c *C) {
	var specs []*parser.Specification
	spec1 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING1"}}
	spec2 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING2"}}
	heading3 := &parser.Heading{Value: "SPECHEADING3"}
	spec3 := &parser.Specification{Heading: heading3}
	spec4 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING4"}}
	spec5 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING5"}}
	spec6 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING6"}}
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = append(specs, spec4)
	specs = append(specs, spec5)
	specs = append(specs, spec6)

	value := 6
	value1 := 3

	groupFilter := &specsGroupFilter{value1, value}
	specsToExecute := groupFilter.filter(specs)

	c.Assert(len(specsToExecute), Equals, 1)
	c.Assert(specsToExecute[0].Heading, Equals, heading3)

}

func (s *MySuite) TestToRunSpecificSetOfSpecsGivesSameSpecsEverytime(c *C) {
	spec1 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING1"}}
	spec2 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING2"}}
	spec3 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING3"}}
	spec4 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING4"}}
	heading5 := &parser.Heading{Value: "SPECHEADING5"}
	spec5 := &parser.Specification{Heading: heading5}
	heading6 := &parser.Heading{Value: "SPECHEADING6"}
	spec6 := &parser.Specification{Heading: heading6}
	var specs []*parser.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = append(specs, spec4)
	specs = append(specs, spec5)
	specs = append(specs, spec6)

	value := 3

	groupFilter := &specsGroupFilter{value, value}
	specsToExecute1 := groupFilter.filter(specs)
	c.Assert(len(specsToExecute1), Equals, 2)

	specsToExecute2 := groupFilter.filter(specs)
	c.Assert(len(specsToExecute2), Equals, 2)

	specsToExecute3 := groupFilter.filter(specs)
	c.Assert(len(specsToExecute3), Equals, 2)

	c.Assert(specsToExecute2[0].Heading, Equals, specsToExecute1[0].Heading)
	c.Assert(specsToExecute2[1].Heading, Equals, specsToExecute1[1].Heading)
	c.Assert(specsToExecute3[0].Heading, Equals, specsToExecute1[0].Heading)
	c.Assert(specsToExecute3[1].Heading, Equals, specsToExecute1[1].Heading)
}

func (s *MySuite) TestToRunNonExistingSpecificSetOfSpecs(c *C) {
	spec1 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING1"}}
	var specs []*parser.Specification
	specs = append(specs, spec1)
	value := 3
	groupFilter := &specsGroupFilter{value, value}
	specsToExecute := groupFilter.filter(specs)
	c.Assert(len(specsToExecute), Equals, 0)
}

func (s *MySuite) TestToRunSpecificSetOfSpecsGivesEmptySpecsIfDistributableNumberIsNotValid(c *C) {
	spec1 := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING1"}}
	var specs []*parser.Specification
	specs = append(specs, spec1)

	value := 1
	value1 := 3
	groupFilter := &specsGroupFilter{value1, value}
	specsToExecute1 := groupFilter.filter(specs)
	c.Assert(len(specsToExecute1), Equals, 0)

	value = 1
	value1 = -3
	groupFilter = &specsGroupFilter{value1, value}
	specsToExecute1 = groupFilter.filter(specs)
	c.Assert(len(specsToExecute1), Equals, 0)
}
