package result

import (

	gc "github.com/go-check/check"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
)

func (s *MySuite) TestAddNonTableRelatedScenarioResult(c *gc.C) {
	specResult := SpecResult{}

	heading := "Scenario heading"

	item1

	items := []*gauge_messages.ProtoItem{item1}

	scenarioResult := NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading:heading, ScenarioItems:items })

	specResult.AddNonTableRelatedScenarioResult(scenarioResult)

}