/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package api

import (
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/gauge"
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
			Errs: []parser.ParseError{{Message: "Scenario1 not found"}, {Message: "Scenario2 not found"}},
		},
		{
			Spec: &gauge.Specification{},
			Errs: []parser.ParseError{{Message: "Scenarios not found"}},
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
