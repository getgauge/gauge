package gauge

import (
	"github.com/getgauge/gauge/gauge_messages"
	. "github.com/go-check/check"
)

func (s *MySuite) TestGetAllDynamicParams(c *C) {
	heading := &Heading{
		HeadingType: ScenarioHeading,
		LineNo:      1,
		Value:       "simple heading",
	}

	fragment1 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Text:         "print ",
	}

	fragment2 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Parameter: &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Dynamic,
			Name:          "name",
		},
	}

	fragment3 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Parameter: &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Dynamic,
			Name:          "id",
		},
	}

	fragment4 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Parameter: &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Static,
			Value:         "abc",
		},
	}

	fragments1 := []*gauge_messages.Fragment{fragment1, fragment2, fragment3, fragment4}
	fragments2 := []*gauge_messages.Fragment{fragment1, fragment4}

	step1 := &Step{
		Fragments: fragments1,
	}

	step2 := &Step{
		Fragments: fragments2,
	}

	steps := []*Step{step1, step2}

	scenario := Scenario{
		Heading: heading,
		Steps:   steps,
	}

	dynamicParams := scenario.IsUsingDynamicParamInScenario()

	c.Assert(2, Equals, len(dynamicParams))
	c.Assert(dynamicParams[0], Equals, "name")
	c.Assert(dynamicParams[1], Equals, "id")

}
