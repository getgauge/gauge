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

func createNode(typ nodeType, value, rawText string, lineNum int, children []*Node) *Node {
	n := newNode(typ, value, rawText, lineNum)
	n.children = children
	return n
}

var conceptParseTests = []conceptParseTest{
	{"simple concept", "# This is a concept heading\n* This is the first step\n* This is the second step\n",
		createNode(nodeConcept, "This is a concept heading", "# This is a concept heading", 1, []*Node{
			newNode(nodeStep, "This is the first step", "* This is the first step", 2),
			newNode(nodeStep, "This is the second step", "* This is the second step", 3),
		})},
	{"simple underline concept", "This is a concept heading\n=======================\n* This is the first step\n* This is the second step",
		createNode(nodeConcept, "This is a concept heading", "This is a concept heading\n=======================", 1, []*Node{
			newNode(nodeStep, "This is the first step", "* This is the first step", 3),
			newNode(nodeStep, "This is the second step", "* This is the second step", 4),
		})},
	{"simple concept with extra newlines", "# This is a concept heading\n\n\n* This is the first step\n\n\n* This is the second step\n\n\n",
		createNode(nodeConcept, "This is a concept heading", "# This is a concept heading", 1, []*Node{
			newNode(nodeStep, "This is the first step", "* This is the first step", 4),
			newNode(nodeStep, "This is the second step", "* This is the second step", 7),
		})},
	{"simple underline concept with extra newlines", "This is a concept heading\n=======================\n\n\n* This is the first step\n\n\n* This is the second step\n\n\n",
		createNode(nodeConcept, "This is a concept heading", "This is a concept heading\n=======================", 1, []*Node{
			newNode(nodeStep, "This is the first step", "* This is the first step", 5),
			newNode(nodeStep, "This is the second step", "* This is the second step", 8),
		})},
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
	if root1.rawText != root2.rawText {
		return false
	}
	if root1.lineNum != root2.lineNum {
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
