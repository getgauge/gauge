/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"reflect"
	"testing"
)

func TestSpecCollection(t *testing.T) {
	s1 := &Specification{FileName: "filename1"}
	s2 := &Specification{FileName: "filename2"}
	s3 := &Specification{FileName: "filename3"}

	collection := NewSpecCollection([]*Specification{s1, s2, s3}, false)

	got := [][]*Specification{collection.Next(), collection.Next(), collection.Next()}
	want := [][]*Specification{{s1}, {s2}, {s3}}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Spec Collection Failed\n\tWant: %v\n\t Got:%v", want, got)
	}
}

func TestGroupSpecCollection(t *testing.T) {
	s1 := &Specification{FileName: "filename1"}
	s3 := &Specification{FileName: "filename3"}

	collection := NewSpecCollection([]*Specification{s1, s3, s3}, true)

	got := [][]*Specification{collection.Next(), collection.Next()}
	want := [][]*Specification{{s1}, {s3, s3}}

	if !reflect.DeepEqual(set(want), set(got)) {
		t.Errorf("Spec Collection Failed\n\tWant: %v\n\t Got:%v", want, got)
	}

	for key, value := range got {
		if value[0].FileName != want[key][0].FileName {
			t.Errorf("Spec Collection order is not maintained\n\tWant: %v\n\t Got:%v", want[key], got[key])
		}
	}
}

func set(s [][]*Specification) map[int][]*Specification {
	specs := make(map[int][]*Specification)
	for _, sp := range s {
		specs[len(sp)] = sp
	}
	return specs
}
