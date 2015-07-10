package version

import (
	"fmt"
	"errors"
)

func CheckCompatibility(currentVersion *Version, versionSupport *VersionSupport) error {
	minSupportVersion, err := ParseVersion(versionSupport.Minimum)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid minimum support version %s. : %s. ", versionSupport.Minimum, err))
	}
	if versionSupport.Maximum != "" {
		maxSupportVersion, err := ParseVersion(versionSupport.Maximum)
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid maximum support version %s. : %s. ", versionSupport.Maximum, err))
		}
		if currentVersion.IsBetween(minSupportVersion, maxSupportVersion) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Version %s is not between %s and %s", currentVersion, minSupportVersion, maxSupportVersion))
		}
	}

	if minSupportVersion.IsLesserThanEqualTo(currentVersion) {
		return nil
	}
	return errors.New(fmt.Sprintf("Incompatible version. Minimum support version %s is higher than current version %s", minSupportVersion, currentVersion))
}
