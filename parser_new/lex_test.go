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

func stringEquals(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

// func TestLine(t *testing.T) {
// 	input := "hi, hello, how are you?\ni am fine\rthank you!"
// 	l := New("", input)
// 	want := []string{"hi, hello, how are you?\n", "i am fine\r", "thank you!"}
// 	got := make([]string, 0)

// 	got = append(got, l.line())
// 	got = append(got, l.line())
// 	got = append(got, l.line())

// 	if !stringEquals(got, want) {
// 		t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", "test line", got, want)
// 	}
// }

// func TestPeekLine(t *testing.T) {
// 	input := "hi, hello, how are you?\ni am fine\rthank you!"
// 	l := New("", input)
// 	want := []string{"i am fine\r", "hi, hello, how are you?\n", "thank you!", "i am fine\r", "", "thank you!"}
// 	got := make([]string, 0)

// 	got = append(got, l.peekNextLine())
// 	got = append(got, l.line())
// 	got = append(got, l.peekNextLine())
// 	got = append(got, l.line())
// 	got = append(got, l.peekNextLine())
// 	got = append(got, l.line())

// 	if !stringEquals(got, want) {
// 		t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", "test peek line", got, want)
// 	}
// }

type lexTest struct {
	name  string
	input string
	items []item
}

// var (
// 	{itemEOF, 0, 0, ""} = item{itemEOF, 0, 0, ""}
// )

var lexTests = []lexTest{
	{"empty", "", []item{item{itemEOF, 0, 1, ""}}},
	{"only spaces", "    \n     \r    \t    \n ", []item{item{itemEOF, 0, 1, ""}}},
	{"concept heading with hash", "# This is a concept heading  \n", []item{
		{itemH1Hash, 0, 1, "#"},
		{itemText, 0, 1, " This is a concept heading  "},
		{itemNewline, 0, 1, "\n"},
		{itemEOF, 0, 2, ""},
	}},
	{"concept heading with double underline", "This is a simple concept\n========================", []item{
		{itemText, 0, 1, "This is a simple concept"},
		{itemNewline, 0, 1, "\n"},
		{itemDoubleUnderline, 0, 2, "========================"},
		{itemEOF, 0, 2, ""},
	}},
	{"concept heading with double underline and newline", "This is a simple concept\n========================\n", []item{
		{itemText, 0, 1, "This is a simple concept"},
		{itemNewline, 0, 1, "\n"},
		{itemDoubleUnderline, 0, 2, "========================"},
		{itemNewline, 0, 2, "\n"},
		{itemEOF, 0, 3, ""},
	}},
	{"step without params with hashed concept heading", "# This is a concept heading  \n\n* This is the first step", []item{
		{itemH1Hash, 0, 1, "#"},
		{itemText, 0, 1, " This is a concept heading  "},
		{itemNewline, 0, 1, "\n"},
		{itemNewline, 0, 2, "\n"},
		{itemAsterisk, 0, 3, "*"},
		{itemText, 0, 3, " This is the first step"},
		{itemEOF, 0, 3, ""},
	}},
	{"concept heading and step without newline in between", "# This is a concept heading\n* This is the first step", []item{
		{itemH1Hash, 0, 1, "#"},
		{itemText, 0, 1, " This is a concept heading"},
		{itemNewline, 0, 1, "\n"},
		{itemAsterisk, 0, 2, "*"},
		{itemText, 0, 2, " This is the first step"},
		{itemEOF, 0, 2, ""},
	}},
	{"step ending in newline", "# This is a concept heading\n* This is the first step\n", []item{
		{itemH1Hash, 0, 1, "#"},
		{itemText, 0, 1, " This is a concept heading"},
		{itemNewline, 0, 1, "\n"},
		{itemAsterisk, 0, 2, "*"},
		{itemText, 0, 2, " This is the first step"},
		{itemNewline, 0, 2, "\n"},
		{itemEOF, 0, 3, ""},
	}},
	{"step without params with underlined concept heading", "This is a simple concept\n========================\n\n* This is the first step", []item{
		{itemText, 0, 1, "This is a simple concept"},
		{itemNewline, 0, 1, "\n"},
		{itemDoubleUnderline, 0, 2, "========================"},
		{itemNewline, 0, 2, "\n"},
		{itemNewline, 0, 3, "\n"},
		{itemAsterisk, 0, 4, "*"},
		{itemText, 0, 4, " This is the first step"},
		{itemEOF, 0, 4, ""},
	}},
	{"concept with multiple simple steps", "# This is a concept heading\n* This is the first step\n* This is the second step\n", []item{
		{itemH1Hash, 0, 1, "#"},
		{itemText, 0, 1, " This is a concept heading"},
		{itemNewline, 0, 1, "\n"},
		{itemAsterisk, 0, 2, "*"},
		{itemText, 0, 2, " This is the first step"},
		{itemNewline, 0, 2, "\n"},
		{itemAsterisk, 0, 3, "*"},
		{itemText, 0, 3, " This is the second step"},
		{itemNewline, 0, 3, "\n"},
		{itemEOF, 0, 4, ""},
	}},
	{"underline concept with multiple simple steps", "This is a concept heading\n========================\n* This is the first step\n* This is the second step\n", []item{
		{itemText, 0, 1, "This is a concept heading"},
		{itemNewline, 0, 1, "\n"},
		{itemDoubleUnderline, 0, 2, "========================"},
		{itemNewline, 0, 2, "\n"},
		{itemAsterisk, 0, 3, "*"},
		{itemText, 0, 3, " This is the first step"},
		{itemNewline, 0, 3, "\n"},
		{itemAsterisk, 0, 4, "*"},
		{itemText, 0, 4, " This is the second step"},
		{itemNewline, 0, 4, "\n"},
		{itemEOF, 0, 5, ""},
	}},
}

func itemEquals(slice1, slice2 []item, ignorePos bool) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i].typ != slice2[i].typ {
			return false
		}
		if !ignorePos {
			if slice1[i].pos != slice2[i].pos {
				return false
			}
		}
		if slice1[i].lineNum != slice2[i].lineNum {
			return false
		}
		if slice1[i].val != slice2[i].val {
			return false
		}
	}
	return true
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := newLexer(t.name, t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF {
			break
		}
	}
	return
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !itemEquals(items, test.items, true) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}
