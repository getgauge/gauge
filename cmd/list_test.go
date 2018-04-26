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

package cmd

import (
	"reflect"
	"testing"
)

func TestUniqueElementsOfFiltersIdenticalNames(t *testing.T) {
	unique := uniqueNonEmptyElementsOf([]string{"Peralta", "Santiago", "Boyle", "Peralta", "Scully"})
	want := []string{"Peralta", "Santiago", "Boyle", "Scully"}
	if !reflect.DeepEqual(unique, want) {
		t.Errorf("want: `%s`,\n got: `%s` ", unique, want)
	}
}

func TestUniqueElementsOfRemainsUniqueList(t *testing.T) {
	names := []string{"Peralta", "Santiago", "Diaz"}
	unique := uniqueNonEmptyElementsOf(names)
	if !reflect.DeepEqual(unique, names) {
		t.Errorf("want: `%s`,\n got: `%s` ", unique, names)
	}
}
