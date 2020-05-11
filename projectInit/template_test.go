/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
