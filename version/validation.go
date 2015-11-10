package version

import (
	"fmt"
)

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
		} else {
			return fmt.Errorf("Version %s is not between %s and %s", currentVersion, minSupportVersion, maxSupportVersion)
		}
	}

	if minSupportVersion.IsLesserThanEqualTo(currentVersion) {
		return nil
	}
	return fmt.Errorf("Incompatible version. Minimum support version %s is higher than current version %s", minSupportVersion, currentVersion)
}
