package mmdbformat

import (
	"errors"
	"io"
	"sync"

	"github.com/anexia-it/geodbtools"
	"github.com/oschwald/maxminddb-golang"
)

// DatabaseTypeID represents a MMDB database type ID
type DatabaseTypeID string

const (
	// DatabaseTypeIDGeoLite2Country defines the database type of GeoLite2-Country databases
	DatabaseTypeIDGeoLite2Country DatabaseTypeID = "GeoLite2-Country"
	// DatabaseTypeIDGeoIP2Country defines the database type of GeoIP2-Country databases
	DatabaseTypeIDGeoIP2Country DatabaseTypeID = "GeoIP2-Country"

	// DatabaseTypeIDGeoLite2City defines the database type of GeoLite2-City databases
	DatabaseTypeIDGeoLite2City DatabaseTypeID = "GeoLite2-City"
	// DatabaseTypeIDGeoIP2City defines the database type of GeoIP2-City databases
	DatabaseTypeIDGeoIP2City DatabaseTypeID = "GeoIP2-City"
)

// Type describes a database type
type Type interface {
	// DatabaseType returns the database type
	DatabaseType() geodbtools.DatabaseType

	// NewReader returns a new database reader
	NewReader(dbReader *maxminddb.Reader) (reader geodbtools.Reader, err error)

	// NewWriter returns a new database writer
	NewWriter(w io.Writer, ipVersion geodbtools.IPVersion) (writer geodbtools.Writer, err error)
}

var typeRegistryMu sync.RWMutex
var typeRegistry = map[DatabaseTypeID]Type{}

var (
	// ErrTypeRegistered indicates that the database type has already been registered
	ErrTypeRegistered = errors.New("database type is registered")

	// ErrTypeNotFound indicates that the database type has not been found
	ErrTypeNotFound = errors.New("database type not found")
)

// RegisterType registers a database type
func RegisterType(typeID DatabaseTypeID, t Type) (err error) {
	typeRegistryMu.Lock()
	defer typeRegistryMu.Unlock()

	if _, exists := typeRegistry[typeID]; exists {
		err = ErrTypeRegistered
		return
	}

	typeRegistry[typeID] = t
	return
}

// MustRegisterType registers a database type and panics on error
func MustRegisterType(typeID DatabaseTypeID, t Type) {
	if err := RegisterType(typeID, t); err != nil {
		panic(err)
	}
}

// LookupType retrieves the type for a given geodbtools.DatabaseType string
func LookupType(dbType geodbtools.DatabaseType) (t Type, typeID DatabaseTypeID, err error) {
	typeRegistryMu.RLock()
	defer typeRegistryMu.RUnlock()

	for typeID, t = range typeRegistry {
		if t.DatabaseType() == dbType {
			return
		}
	}

	t = nil
	typeID = ""
	err = ErrTypeNotFound
	return
}

// LookupTypeByDatabaseType retrieves the type for a given DatabaseTypeID constant
func LookupTypeByDatabaseType(typeID DatabaseTypeID) (t Type, err error) {
	typeRegistryMu.RLock()
	defer typeRegistryMu.RUnlock()

	var found bool
	if t, found = typeRegistry[typeID]; !found {
		t = nil
		err = ErrTypeNotFound
	}

	return
}
