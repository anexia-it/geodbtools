package mmdbformat

import (
	"fmt"
	"net"

	"github.com/anexia-it/geodbtools"
)

var _ geodbtools.CountryRecord = (*countryRecord)(nil)
var _ Record = (*countryRecord)(nil)

// countryRecord represents a record with country information
type countryRecord struct {
	network *net.IPNet

	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func (r *countryRecord) SetNetwork(network *net.IPNet) {
	r.network = network
}

func (r *countryRecord) String() string {
	return fmt.Sprintf("%s: country code %s", r.network, r.Country.ISOCode)
}

func (r *countryRecord) GetRecordKey() []byte {
	if r.network != nil {
		return r.network.IP
	}
	return net.IPv6zero
}

func (r *countryRecord) GetNetwork() *net.IPNet {
	return r.network
}

func (r *countryRecord) GetCountryCode() string {
	return r.Country.ISOCode
}
