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

package api

import (
	"testing"

	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestCreateSpecsResponseMessageFor(c *C) {
	h := &gaugeAPIMessageHandler{}
	m := h.createSpecsResponseMessageFor([]*infoGatherer.SpecDetail{
		{
			Spec: &gauge.Specification{Heading: &gauge.Heading{Value: "Spec heading 1"}},
			Errs: []*parser.ParseError{{Message: "Scenario1 not found"}, {Message: "Scenario2 not found"}},
		},
		{
			Spec: &gauge.Specification{},
			Errs: []*parser.ParseError{{Message: "Scenarios not found"}},
		},
		{
			Spec: &gauge.Specification{Heading: &gauge.Heading{Value: "Spec heading 2"}},
		},
	})

	var nilSpec *gauge_messages.ProtoSpec

	c.Assert(len(m.GetDetails()), Equals, 3)
	c.Assert(len(m.GetDetails()[0].ParseErrors), Equals, 2)
	c.Assert(m.GetDetails()[0].Spec.GetSpecHeading(), Equals, "Spec heading 1")
	c.Assert(len(m.GetDetails()[1].ParseErrors), Equals, 1)
	c.Assert(m.GetDetails()[1].GetSpec(), Equals, nilSpec)
	c.Assert(len(m.GetDetails()[2].ParseErrors), Equals, 0)
	c.Assert(m.GetDetails()[2].Spec.GetSpecHeading(), Equals, "Spec heading 2")
}
