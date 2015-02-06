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

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	DOT = "."
)

var currentGaugeVersion = &version{0, 0, 3}

type version struct {
	major int
	minor int
	patch int
}

func parseVersion(versionText string) (*version, error) {
	splits := strings.Split(versionText, DOT)
	if len(splits) != 3 {
		return nil, errors.New("Incorrect number of '.' characters in Version. Version should be of the form 1.5.7")
	}
	major, err := strconv.Atoi(splits[0])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing major version number %s to integer. %s", splits[0], err.Error()))
	}
	minor, err := strconv.Atoi(splits[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing minor version number %s to integer. %s", splits[0], err.Error()))
	}
	patch, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing patch version number %s to integer. %s", splits[0], err.Error()))
	}

	return &version{major, minor, patch}, nil
}

func (version *version) isBetween(lower *version, greater *version) bool {
	return version.isGreaterThanEqualTo(lower) && version.isLesserThanEqualTo(greater)
}

func (version *version) isLesserThan(version1 *version) bool {
	return compareVersions(version, version1, lesserThanFunc)
}

func (version *version) isGreaterThan(version1 *version) bool {
	return compareVersions(version, version1, greaterThanFunc)
}

func (version *version) isLesserThanEqualTo(version1 *version) bool {
	return version.isLesserThan(version1) || version.isEqualTo(version1)
}

func (version *version) isGreaterThanEqualTo(version1 *version) bool {
	return version.isGreaterThan(version1) || version.isEqualTo(version1)
}

func (version *version) isEqualTo(version1 *version) bool {
	return isEqual(version.major, version1.major) && isEqual(version.minor, version1.minor) && isEqual(version.patch, version1.patch)
}

func compareVersions(first *version, second *version, compareFunc func(int, int) bool) bool {
	if compareFunc(first.major, second.major) {
		return true
	} else if isEqual(first.major, second.major) {
		if compareFunc(first.minor, second.minor) {
			return true
		} else if isEqual(first.minor, second.minor) {
			if compareFunc(first.patch, second.patch) {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func lesserThanFunc(first, second int) bool {
	return first < second
}

func greaterThanFunc(first, second int) bool {
	return first > second
}

func isEqual(first, second int) bool {
	return first == second
}

func (version *version) String() string {
	return fmt.Sprintf("%d.%d.%d", version.major, version.minor, version.patch)
}
