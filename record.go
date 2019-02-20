package geodbtools

import (
	"fmt"
	"net"

	"github.com/anexia-it/bitmap"
)

// Record describes a database record
type Record interface {
	fmt.Stringer

	// GetNetwork returns the network represented by the record
	GetNetwork() *net.IPNet
}

// CountryRecord describes a database record holding country-specific information
type CountryRecord interface {
	Record

	// GetCountryCode returns the 2-character ISO country code
	GetCountryCode() string
}

// CityRecord describes a database record holding city-specific information
type CityRecord interface {
	CountryRecord

	// GetCityName returns the name of the city
	GetCityName() string
}

// RecordBelongsRightIPv6 defines the "belongs right" test function for IPv6 addresses
func RecordBelongsRightIPv6(b []byte, depth uint) bool {
	if len(b) < 16 {
		return false
	}

	targetByte := b[((127 - depth) >> 3)]
	checkBit := ^(127 - depth) & 7
	return bitmap.IsSet([]byte{targetByte}, checkBit)
}
