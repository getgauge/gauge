// Copyright 2019 ThoughtWorks, Inc.

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

package parser

import (
	"reflect"
	"testing"
)

func TestSpecFileCollection(t *testing.T) {
	f1 := "filename1"
	f2 := "filename2"
	f3 := "filename3"

	collection := NewSpecFileCollection([]string{f1, f2, f3})

	i1, _ := collection.Next()
	i2, _ := collection.Next()
	i3, _ := collection.Next()

	got := []string{i1, i2, i3}
	want := []string{f1, f2, f3}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Spec Collection Failed\n\tWant: %v\n\t Got:%v", want, got)
	}
}

func TestSpecFileCollectionWithItems(t *testing.T) {
	collection := NewSpecFileCollection([]string{})

	i, err := collection.Next()
	if i != "" {
		t.Errorf("Spec Collection Failed\n\tWant: %v\n\t Got:%v", "", i)
	}
	errMessage := "no files in collection"
	if !reflect.DeepEqual(err.Error(), errMessage) {
		t.Errorf("Expected error to be - \n%s\nBut got -\n%s", errMessage, err.Error())
	}
}
