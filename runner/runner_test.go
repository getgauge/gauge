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

package runner

import (
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

	want := "path1:path2"
	if !strings.Contains(env[0], want) {
		t.Errorf("getCleanEnv failed. Did not append to path.\n\tWanted PATH to contain: `%s`", want)
	}
}
