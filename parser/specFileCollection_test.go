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

	got := []string{collection.Next(), collection.Next(), collection.Next()}
	want := []string{f1, f2, f3}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Spec Collection Failed\n\tWant: %v\n\t Got:%v", want, got)
	}
}
