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

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Spec Collection Failed\n\tWant: %v\n\t Got:%v", want, got)
	}

}
