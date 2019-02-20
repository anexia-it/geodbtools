package mmdatformat

import (
	"fmt"
	"net"

	"github.com/anexia-it/geodbtools"
)

var _ geodbtools.CountryRecord = (*countryRecord)(nil)

type countryRecord struct {
	network     *net.IPNet
	countryCode string
}

func (r *countryRecord) GetRecordKey() []byte {
	return r.network.IP
}

func (r *countryRecord) GetNetwork() *net.IPNet {
	return r.network
}

func (r *countryRecord) GetCountryCode() string {
	return r.countryCode
}

func (r *countryRecord) String() string {
	return fmt.Sprintf("%s: country code %s", r.network, r.countryCode)
}
