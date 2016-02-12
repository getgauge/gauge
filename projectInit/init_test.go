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
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestGetTemplateRegexMatches(c *C) {
	input := `<html>
<head>
<script>
//1452251363769
function navi(e){
	location.href = e.target.href.replace('/:','/'); e.preventDefault();
}
</script>
</head>
<body>
<pre><a onclick="navi(event)" href=":java.zip" rel="nofollow">java.zip</a></pre>
<pre><a onclick="navi(event)" href=":java_maven_selenium.zip" rel="nofollow">maven_123.zip</a></pre>
<pre><a onclick="navi(event)" href=":java_maven.zip" rel="nofollow">java_maven.zip</a></pre>
<pre><a onclick="navi(event)" href=":java_maven_selenium.zip" rel="nofollow">java_m@#_selenium.zip</a></pre>
<pre><a onclick="navi(event)" href=":java_maven_selenium.zip" rel="nofollow">java_maven_selenium.zip</a></pre>
</body>
</html>
`
	names := getTemplateNames(input)
	c.Assert(names[0], Equals, "csharp")
	c.Assert(names[1], Equals, "java")
	c.Assert(names[2], Equals, "java_maven")
	c.Assert(names[3], Equals, "java_maven_selenium")
	c.Assert(names[4], Equals, "maven_123")
	c.Assert(names[5], Equals, "ruby")
}

func (s *MySuite) TestGetTemplateLanguage(c *C) {
	c.Assert(getTemplateLangauge("java"), Equals, "java")
	c.Assert(getTemplateLangauge("java_maven"), Equals, "java")
	c.Assert(getTemplateLangauge("java_maven_selenium"), Equals, "java")
}
