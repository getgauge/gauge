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

package projectInit

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestTemplateZipName(c *C) {
	t := template{Name: "gauge.zip"}
	c.Assert("gauge", Equals, t.GetName())
}

func (s *MySuite) TestTemplateName(c *C) {
	t := template{Name: "gauge"}
	c.Assert("gauge", Equals, t.GetName())
}
