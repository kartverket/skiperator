package k8sfeatures

import (
	"fmt"
	"k8s.io/apimachinery/pkg/version"
	"os"
	"strconv"
)

type VersionInfo struct {
	*version.Info
	MajorInt int
	MinorInt int
}

func NewVersionInfo(info *version.Info) {
	if info == nil {
		fmt.Println("no version information available")
		os.Exit(1)
	}

	major, errOne := strconv.Atoi(info.Major)
	minor, errTwo := strconv.Atoi(info.Minor)
	if errOne != nil || errTwo != nil {
		fmt.Println("fatal error during version parsing")
		os.Exit(1)
	}

	currentVersion = &VersionInfo{
		info,
		major,
		minor,
	}
}

func (v VersionInfo) greaterOrEqualToMinorVersion(desiredMinorVersion int) bool {
	return v.MinorInt >= desiredMinorVersion
}

// Please do not overwrite runtime :D
var currentVersion *VersionInfo

// EnhancedPDBAvailable checks whether the server version is equal or above 1.27 in order to
// support the new field `spec.unhealthyPodEvictionPolicy`.
func EnhancedPDBAvailable() bool {
	return currentVersion.greaterOrEqualToMinorVersion(27)
}
