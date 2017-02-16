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

package gauge

import . "github.com/go-check/check"

func (s *MySuite) TestDeepCopyOfConcept(c *C) {
	step := &Step{Value: "test concept step 1",
		LineText:  "dsfdsfdsf",
		IsConcept: true, LineNo: 2,
		ConceptSteps: []*Step{
			&Step{Value: "sfd",
				LineText:  "sfd",
				IsConcept: false},
			&Step{Value: "sdfsdf" + "T",
				LineText:  "sdfsdf" + "T",
				IsConcept: false}}}

	concept := &Concept{ConceptStep: step, FileName: "file.cpt"}

	copiedTopLevelConcept := concept.deepCopy()
	verifyCopiedConcept(copiedTopLevelConcept, concept, c)
}

func verifyCopiedConcept(copiedConcept *Concept, actualConcept *Concept, c *C) {
	c.Assert(&copiedConcept, Not(Equals), &actualConcept)
	c.Assert(copiedConcept, DeepEquals, actualConcept)
}
