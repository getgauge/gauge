/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package version

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CurrentGaugeVersion represents the current version of Gauge
var CurrentGaugeVersion = &Version{1, 6, 22}

// BuildMetadata represents build information of current release (e.g, nightly build information)
var BuildMetadata = ""
var CommitHash = ""

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
		return nil, fmt.Errorf("incorrect version format. version should be in the form 1.5.7")
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
	return fmt.Errorf("error parsing %s version %s to integer. %s", level, text, err.Error())
}

func (Version *Version) IsBetween(lower, greater *Version) bool {
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

func CompareVersions(first, second *Version, compareFunc func(int, int) bool) bool {
	if compareFunc(first.Major, second.Major) {
		return true
	} else if IsEqual(first.Major, second.Major) {
		if compareFunc(first.Minor, second.Minor) {
			return true
		} else if IsEqual(first.Minor, second.Minor) {
			return compareFunc(first.Patch, second.Patch)
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
		return fmt.Errorf("invalid minimum support version %s. : %s. ", versionSupport.Minimum, err.Error())
	}
	if versionSupport.Maximum != "" {
		maxSupportVersion, err := ParseVersion(versionSupport.Maximum)
		if err != nil {
			return fmt.Errorf("invalid maximum support version %s. : %s. ", versionSupport.Maximum, err.Error())
		}
		if currentVersion.IsBetween(minSupportVersion, maxSupportVersion) {
			return nil
		}
		return fmt.Errorf("version %s is not between %s and %s", currentVersion, minSupportVersion, maxSupportVersion)
	}

	if minSupportVersion.IsLesserThanEqualTo(currentVersion) {
		return nil
	}
	return fmt.Errorf("incompatible version. Minimum support version %s is higher than current version %s", minSupportVersion, currentVersion)
}
