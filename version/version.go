// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either Version 3 of the License, or
// (at your option) any later Version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package version

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CurrentGaugeVersion represents the current version of Gauge
var CurrentGaugeVersion = &Version{0, 8, 5}

// BuildMetadata represents build information of current release (e.g, nightly build information)
var BuildMetadata = ""

type Version struct {
	Major int
	Minor int
	Patch int
}

type VersionSupport struct {
	Minimum string
	Maximum string
}

func ParseVersion(versionText string) (*Version, error) {
	splits := strings.Split(versionText, ".")
	if len(splits) != 3 {
		return nil, fmt.Errorf("Incorrect Version format. Version should be in the form 1.5.7")
	}
	Major, err := strconv.Atoi(splits[0])
	if err != nil {
		return nil, VersionError("major", splits[0], err)
	}
	Minor, err := strconv.Atoi(splits[1])
	if err != nil {
		return nil, VersionError("minor", splits[1], err)
	}
	Patch, err := strconv.Atoi(splits[2])
	if err != nil {
		return nil, VersionError("patch", splits[2], err)
	}

	return &Version{Major, Minor, Patch}, nil
}

func VersionError(level, text string, err error) error {
	return fmt.Errorf("Error parsing %s Version %s to integer. %s", level, text, err.Error())
}

func (Version *Version) IsBetween(lower *Version, greater *Version) bool {
	return Version.IsGreaterThanEqualTo(lower) && Version.IsLesserThanEqualTo(greater)
}

func (Version *Version) IsLesserThan(version1 *Version) bool {
	return CompareVersions(Version, version1, LesserThanFunc)
}

func (Version *Version) IsGreaterThan(version1 *Version) bool {
	return CompareVersions(Version, version1, GreaterThanFunc)
}

func (Version *Version) IsLesserThanEqualTo(version1 *Version) bool {
	return Version.IsLesserThan(version1) || Version.IsEqualTo(version1)
}

func (Version *Version) IsGreaterThanEqualTo(version1 *Version) bool {
	return Version.IsGreaterThan(version1) || Version.IsEqualTo(version1)
}

func (Version *Version) IsEqualTo(version1 *Version) bool {
	return IsEqual(Version.Major, version1.Major) && IsEqual(Version.Minor, version1.Minor) && IsEqual(Version.Patch, version1.Patch)
}

func CompareVersions(first *Version, second *Version, compareFunc func(int, int) bool) bool {
	if compareFunc(first.Major, second.Major) {
		return true
	} else if IsEqual(first.Major, second.Major) {
		if compareFunc(first.Minor, second.Minor) {
			return true
		} else if IsEqual(first.Minor, second.Minor) {
			if compareFunc(first.Patch, second.Patch) {
				return true
			}
			return false
		}
	}
	return false
}

func LesserThanFunc(first, second int) bool {
	return first < second
}

func GreaterThanFunc(first, second int) bool {
	return first > second
}

func IsEqual(first, second int) bool {
	return first == second
}

func (Version *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", Version.Major, Version.Minor, Version.Patch)
}

// FullVersion returns the CurrentGaugeVersion including build metadata.
func FullVersion() string {
	var metadata string
	if BuildMetadata != "" {
		metadata = fmt.Sprintf(".%s", BuildMetadata)
	}
	return fmt.Sprintf("%s%s", CurrentGaugeVersion.String(), metadata)
}

type byDecreasingVersion []*Version

func (a byDecreasingVersion) Len() int      { return len(a) }
func (a byDecreasingVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDecreasingVersion) Less(i, j int) bool {
	return a[i].IsGreaterThan(a[j])
}

func GetLatestVersion(versions []*Version) *Version {
	sort.Sort(byDecreasingVersion(versions))
	return versions[0]
}

func CheckCompatibility(currentVersion *Version, versionSupport *VersionSupport) error {
	minSupportVersion, err := ParseVersion(versionSupport.Minimum)
	if err != nil {
		return fmt.Errorf("Invalid minimum support version %s. : %s. ", versionSupport.Minimum, err.Error())
	}
	if versionSupport.Maximum != "" {
		maxSupportVersion, err := ParseVersion(versionSupport.Maximum)
		if err != nil {
			return fmt.Errorf("Invalid maximum support version %s. : %s. ", versionSupport.Maximum, err.Error())
		}
		if currentVersion.IsBetween(minSupportVersion, maxSupportVersion) {
			return nil
		}
		return fmt.Errorf("Version %s is not between %s and %s", currentVersion, minSupportVersion, maxSupportVersion)
	}

	if minSupportVersion.IsLesserThanEqualTo(currentVersion) {
		return nil
	}
	return fmt.Errorf("Incompatible version. Minimum support version %s is higher than current version %s", minSupportVersion, currentVersion)
}
