// Package geodbtools provides functions for working with GeoIP databases
package geodbtools

import "fmt"

// VersionMajor defines the major version number of geodbtools
const (
	// VersionMajor defines the major version number of geodbtools
	VersionMajor = 1
	// VersionMinor defines the minor version number of geodbtools
	VersionMinor = 0
	// VersionPatch defines the patch version number of geodbtools
	VersionPatch = 1
)

// VersionString returns the complete version as a string
func VersionString() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}
