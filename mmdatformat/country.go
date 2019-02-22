package mmdatformat

import (
	"io"
	"net"

	"github.com/anexia-it/geodbtools"
)

var _ Type = countryType{}

type countryType struct{}

func (t countryType) NewWriter(w io.Writer, ipVersion geodbtools.IPVersion) (writer geodbtools.Writer, err error) {
	var typeID DatabaseTypeID

	switch ipVersion {
	case geodbtools.IPVersion4:
		typeID = DatabaseTypeIDCountryEdition
	case geodbtools.IPVersion6:
		typeID = DatabaseTypeIDCountryEditionV6
	default:
		err = geodbtools.ErrUnsupportedDatabaseType
		return
	}

	return NewWriter(w, t, typeID), nil
}

func (countryType) DatabaseType() geodbtools.DatabaseType {
	return geodbtools.DatabaseTypeCountry
}

func (countryType) IPVersion(dbTypeID DatabaseTypeID) geodbtools.IPVersion {
	switch dbTypeID {
	case DatabaseTypeIDCountryEdition:
		return geodbtools.IPVersion4
	case DatabaseTypeIDCountryEditionV6:
		return geodbtools.IPVersion6
	}

	return geodbtools.IPVersionUndefined
}

func (countryType) RecordLength(dbTypeID DatabaseTypeID) uint {
	switch dbTypeID {
	case DatabaseTypeIDCountryEdition, DatabaseTypeIDCountryEditionV6:
		return 3
	}

	return 0
}

func (countryType) DatabaseSegmentOffset(_ geodbtools.ReaderSource, _ DatabaseTypeID, _ int64) uint32 {
	return countryBegin
}

func (countryType) NewRecord(_ geodbtools.ReaderSource, matchingNetwork *net.IPNet, value uint32) (record geodbtools.Record, err error) {
	var recordCountryCode string
	if recordCountryCode, err = GetISO2CountryCodeString(int(value)); err != nil {
		return
	}

	record = &countryRecord{
		network:     matchingNetwork,
		countryCode: recordCountryCode,
	}
	return
}

func (countryType) EncodeTreeNode(position *uint32, node *geodbtools.RecordTree) (b []byte, additionalNodes []*geodbtools.RecordTree, err error) {
	b = make([]byte, 0, 6)

	var next *geodbtools.RecordTree
	if b, next, err = encodeCountryRecord(position, b, node.Left()); err != nil {
		return
	} else if next != nil {
		additionalNodes = append(additionalNodes, next)
	}

	if b, next, err = encodeCountryRecord(position, b, node.Right()); err != nil {
		return
	} else if next != nil {
		additionalNodes = append(additionalNodes, next)
	}
	return
}

func encodeCountryRecord(position *uint32, b []byte, node *geodbtools.RecordTree) (updatedB []byte, next *geodbtools.RecordTree, err error) {
	var value uint32
	if node != nil {
		if leaf := node.Leaf(); leaf != nil {
			countryRecord, ok := leaf.(geodbtools.CountryRecord)
			if !ok {
				err = ErrUnsupportedRecordType
				return
			}

			var idx int
			if idx, err = GetISO2CountryCodeIndex(countryRecord.GetCountryCode()); err != nil {
				return
			}
			value = uint32(idx) + countryBegin
		} else {
			*position = *position + 1
			value = *position
			next = node
		}
	} else {
		// unknown country
		value = countryBegin
	}

	var rec []byte
	if rec, err = EncodeRecord(value, 3); err != nil {
		return
	}

	updatedB = append(b, rec...)
	return
}

func init() {
	MustRegisterType(DatabaseTypeIDCountryEdition, countryType{})
	MustRegisterType(DatabaseTypeIDCountryEditionV6, countryType{})
}
