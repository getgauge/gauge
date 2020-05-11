// Copyright 2019 ThoughtWorks, Inc.

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
