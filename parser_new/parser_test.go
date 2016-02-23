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

package parse

import "testing"

type conceptParseTest struct {
	name string
	text string
	root *Node
}

func createNode(typ nodeType, value string, children []*Node) *Node {
	n := newNode(typ, value)
	n.children = children
	return n
}

var conceptParseTests = []conceptParseTest{
	{"simple concept", "# This is a concept heading\n* This is the first step\n* This is the second step\n", createNode(nodeConcept, "This is a concept heading", []*Node{newNode(nodeStep, "This is the first step"), newNode(nodeStep, "This is the second step")})},
	{"simple underline concept", "This is a concept heading\n=======================\n* This is the first step\n* This is the second step", createNode(nodeConcept, "This is a concept heading", []*Node{newNode(nodeStep, "This is the first step"), newNode(nodeStep, "This is the second step")})},
}

func TestConceptParsing(t *testing.T) {
	for _, test := range conceptParseTests {
		cptNode := Concept(test.name, test.text)
		if !equals(cptNode, test.root) {
			t.Errorf("%s: \ngot\n\t%+v\nexpected\n\t%v", test.name, cptNode, test.root)
		}
	}
}

func equals(root1, root2 *Node) bool {
	if root1.nodeType != root2.nodeType {
		return false
	}
	if root1.value != root2.value {
		return false
	}
	if len(root1.children) != len(root2.children) {
		return false
	}

	isEqual := true
	for i1 := range root1.children {
		for i2 := range root2.children {
			if equals(root1.children[i1], root2.children[i2]) {
				isEqual = true
				break
			} else {
				isEqual = false
			}
		}
		if !isEqual {
			break
		}
	}
	return isEqual
}
