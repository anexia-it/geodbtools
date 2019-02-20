package geodbtools

import "time"

// DatabaseType defines the type of a database
type DatabaseType string

const (
	// DatabaseTypeCountry defines the country database type
	DatabaseTypeCountry DatabaseType = "country"
)

// IPVersion defines an IP version
type IPVersion uint

const (
	// IPVersionUndefined is used for non-IP databases
	IPVersionUndefined IPVersion = 0
	// IPVersion4 is used for IPv4 databases
	IPVersion4 IPVersion = 4
	// IPVersion6 is used for IPv6 databases
	IPVersion6 IPVersion = 6
)

// Metadata represents a database's metadata
type Metadata struct {
	// Type holds the database type
	Type DatabaseType
	// BuildTime holds the database's build time
	BuildTime time.Time
	// Description holds the human-readable database description
	Description string

	// MajorFormatVersion holds the major version number of the database format
	MajorFormatVersion uint
	// MinorFormatVersion holds the minor version number of the database format
	MinorFormatVersion uint

	// IPVersion holds the IP version represented by the database.
	// This may be IPVersionUndefined for databases non-IP databases
	IPVersion IPVersion
}
