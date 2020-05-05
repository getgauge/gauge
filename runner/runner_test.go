/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package runner

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/getgauge/common"
)

func TestGetCleanEnvRemovesGAUGE_INTERNAL_PORTAndSetsPortNumber(t *testing.T) {
	HELLO := "HELLO"
	portVariable := common.GaugeInternalPortEnvName + "=1234"
	PORT_NAME_WITH_EXTRA_WORD := "b" + common.GaugeInternalPortEnvName
	PORT_NAME_WITH_SPACES := "      " + common.GaugeInternalPortEnvName + "         "
	env := []string{HELLO, common.GaugeInternalPortEnvName + "=", common.GaugeInternalPortEnvName,
		PORT_NAME_WITH_SPACES, PORT_NAME_WITH_EXTRA_WORD}
	want := []string{HELLO, portVariable, portVariable, portVariable, PORT_NAME_WITH_EXTRA_WORD}
	got := getCleanEnv("1234", env, false, []string{})

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Did not clean env.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestGetCleanEnvWithDebugging(t *testing.T) {
	env := getCleanEnv("1234", []string{}, true, []string{})

	if env[1] != "debugging=true" {
		t.Errorf("getCleanEnv failed. Did not add debugging env")
	}
}

func TestGetCleanEnvAddsToPath(t *testing.T) {
	env := getCleanEnv("1234", []string{"PATH=PATH"}, false, []string{"path1", "path2"})

	want := "path1" + string(os.PathListSeparator) + "path2"
	if !strings.Contains(env[0], want) {
		t.Errorf("getCleanEnv failed. Did not append to path.\n\tWanted PATH to contain: `%s`", want)
	}
}
