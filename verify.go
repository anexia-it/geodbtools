package geodbtools

import (
	"fmt"
	"sync"

	"go.uber.org/multierr"
)

var equivalentCountryCodeMapMu sync.RWMutex
var equivalentCountryCodeMap = map[string]string{}

// RegisterEquivalentCountryCode registers a pair of country codes which are deemed as equivalent during verification
func RegisterEquivalentCountryCode(a, b string) {
	equivalentCountryCodeMapMu.Lock()
	defer equivalentCountryCodeMapMu.Unlock()
	equivalentCountryCodeMap[a] = b
}

// AreCountryCodesEqual checks if two country codes are considered equal
func AreCountryCodesEqual(a, b string) bool {
	equivalentCountryCodeMapMu.RLock()
	defer equivalentCountryCodeMapMu.RUnlock()

	return a == b || equivalentCountryCodeMap[a] == b || equivalentCountryCodeMap[b] == a
}

// VerificationError holds information relating to a verification error
type VerificationError struct {
	ExpectedRecord Record
	Record         Record
	LookupError    error
}

func (e *VerificationError) Error() string {
	if e.LookupError != nil {
		return fmt.Sprintf("expected record %s, received error %s", e.ExpectedRecord, e.LookupError)
	}

	return fmt.Sprintf("expected record %s, received record %s", e.ExpectedRecord, e.Record)
}

// VerificationProgress holds information regarding the status of the verification progress
type VerificationProgress struct {
	TotalRecords   int
	CheckedRecords int
}

// Verify tests if a given reader contains all records defined by the given tree.
func Verify(reader Reader, root *RecordTree, progress chan<- *VerificationProgress) (err error) {
	expectedRecords := root.Records()

	for i, expectedRecord := range expectedRecords {
		var record Record
		var lookupErr error

		if progress != nil {
			progress <- &VerificationProgress{
				TotalRecords:   len(expectedRecords),
				CheckedRecords: i,
			}
		}

		network := expectedRecord.GetNetwork()
		if network == nil {
			// ignore record without a network
			continue
		}

		if record, lookupErr = reader.LookupIP(network.IP); lookupErr != nil {
			err = multierr.Append(err, &VerificationError{
				ExpectedRecord: expectedRecord,
				LookupError:    lookupErr,
			})
			continue
		}

		if !RecordsEqual(expectedRecord, record) {
			err = multierr.Append(err, &VerificationError{
				ExpectedRecord: expectedRecord,
				Record:         record,
			})
		}
	}

	if progress != nil {
		progress <- &VerificationProgress{
			TotalRecords:   len(expectedRecords),
			CheckedRecords: len(expectedRecords),
		}
	}

	return
}

// RecordsEqual checks if two records are equal
func RecordsEqual(a, b Record) bool {
	switch recordA := a.(type) {
	case CityRecord:
		return CityRecordsEqual(recordA, b)
	case CountryRecord:
		return CountryRecordsEqual(recordA, b)
	}

	return false
}

// CountryRecordsEqual checks if two CountryRecord instances are equal
func CountryRecordsEqual(a CountryRecord, b Record) bool {
	recordB, isCountryRecord := b.(CountryRecord)
	if !isCountryRecord {
		return false
	}

	return AreCountryCodesEqual(a.GetCountryCode(), recordB.GetCountryCode())
}

// CityRecordsEqual checks if two CityRecord instances are equal
func CityRecordsEqual(a CityRecord, b Record) bool {
	recordB, isCityRecord := b.(CityRecord)
	if !isCityRecord {
		return false
	}

	if !CountryRecordsEqual(a, recordB) {
		return false
	}

	return a.GetCityName() == recordB.GetCityName()
}
