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
	"fmt"

	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestDistributionOfSpecs(c *C) {
	specs := createSpecsList(10)
	specCollections := DistributeSpecs(specs, 10)
	c.Assert(len(specCollections), Equals, 10)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	specCollections = DistributeSpecs(specs, 5)
	c.Assert(len(specCollections), Equals, 5)
	verifySpecCollectionsForSize(c, 2, specCollections...)

	specCollections = DistributeSpecs(specs, 4)
	c.Assert(len(specCollections), Equals, 4)
	verifySpecCollectionsForSize(c, 3, specCollections[:2]...)
	verifySpecCollectionsForSize(c, 2, specCollections[2:]...)

	specCollections = DistributeSpecs(specs, 3)
	c.Assert(len(specCollections), Equals, 3)
	verifySpecCollectionsForSize(c, 4, specCollections[0])
	verifySpecCollectionsForSize(c, 3, specCollections[1:]...)

	specs = createSpecsList(0)
	specCollections = DistributeSpecs(specs, 0)
	c.Assert(len(specCollections), Equals, 0)
}

func verifySpecCollectionsForSize(c *C, size int, specCollections ...*gauge.SpecCollection) {
	for _, collection := range specCollections {
		c.Assert(len(collection.Specs()), Equals, size)
	}
}

func createSpecsList(number int) []*gauge.Specification {
	var specs []*gauge.Specification
	for i := 0; i < number; i++ {
		specs = append(specs, &gauge.Specification{FileName: fmt.Sprint("spec", i)})
	}
	return specs
}

func (s *MySuite) TestToRunSpecificSetOfSpecs(c *C) {
	var specs []*gauge.Specification
	spec1 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING1"}}
	spec2 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING2"}}
	heading3 := &gauge.Heading{Value: "SPECHEADING3"}
	spec3 := &gauge.Specification{Heading: heading3}
	spec4 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING4"}}
	spec5 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING5"}}
	spec6 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING6"}}
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
	spec1 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING1"}}
	spec2 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING2"}}
	spec3 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING3"}}
	spec4 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING4"}}
	heading5 := &gauge.Heading{Value: "SPECHEADING5"}
	spec5 := &gauge.Specification{Heading: heading5}
	heading6 := &gauge.Heading{Value: "SPECHEADING6"}
	spec6 := &gauge.Specification{Heading: heading6}
	var specs []*gauge.Specification
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
	spec1 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING1"}}
	var specs []*gauge.Specification
	specs = append(specs, spec1)
	value := 3
	groupFilter := &specsGroupFilter{value, value}
	specsToExecute := groupFilter.filter(specs)
	c.Assert(len(specsToExecute), Equals, 0)
}

func (s *MySuite) TestToRunSpecificSetOfSpecsGivesEmptySpecsIfDistributableNumberIsNotValid(c *C) {
	spec1 := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING1"}}
	var specs []*gauge.Specification
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
