package mmdbformat

import (
	"errors"
	"io"
	"net"

	"github.com/anexia-it/geodbtools"
	"github.com/oschwald/maxminddb-golang"
)

var _ geodbtools.Reader = (*countryReader)(nil)

type countryReader struct {
	r *maxminddb.Reader
}

func (r *countryReader) RecordTree(ipVersion geodbtools.IPVersion) (tree *geodbtools.RecordTree, err error) {
	tree, err = BuildRecordTree(r.r, ipVersion, func() Record {
		return &countryRecord{}
	})

	return
}

func (r *countryReader) LookupIP(ip net.IP) (record geodbtools.Record, err error) {
	rec := &countryRecord{}

	if err = r.r.Lookup(ip, rec); err != nil {
		return
	}
	rec.network = &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(len(ip)*8, len(ip)*8),
	}

	record = rec
	return
}

type countryType struct {
}

func (countryType) DatabaseType() geodbtools.DatabaseType {
	return geodbtools.DatabaseTypeCountry
}

func (countryType) NewReader(dbReader *maxminddb.Reader) (reader geodbtools.Reader, err error) {
	reader = &countryReader{
		r: dbReader,
	}
	return
}

func (countryType) NewWriter(w io.Writer, ipVersion geodbtools.IPVersion) (writer geodbtools.Writer, err error) {
	err = errors.New("not implemented")
	return
}

func init() {
	MustRegisterType(DatabaseTypeIDGeoLite2Country, countryType{})
	MustRegisterType(DatabaseTypeIDGeoIP2Country, countryType{})
	// TODO: remove the two registrations below once we also support city database
	MustRegisterType(DatabaseTypeIDGeoLite2City, countryType{})
	MustRegisterType(DatabaseTypeIDGeoIP2City, countryType{})
}
