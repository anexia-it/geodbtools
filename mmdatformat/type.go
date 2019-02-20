package mmdatformat

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/anexia-it/geodbtools"
)

// DatabaseTypeID represents a DAT database type ID
type DatabaseTypeID uint8

const (
	// DatabaseTypeIDBase defines the base value for database types
	// Newer databases have this offset added to their actual type value, older not.
	DatabaseTypeIDBase DatabaseTypeID = 105
	// DatabaseTypeIDCountryEdition is an IPv4 country database
	DatabaseTypeIDCountryEdition = DatabaseTypeIDBase + 1
	// DatabaseTypeIDCountryEditionV6 is an IPv6 country database
	DatabaseTypeIDCountryEditionV6 = DatabaseTypeIDBase + 12
)

const (
	countryBegin uint32 = 16776960
)

// Type describes a database type
type Type interface {
	// DatabaseType returns the database type
	DatabaseType() geodbtools.DatabaseType

	// EncodeTreeNode encodes a given tree node and returns its representation as a byte-slice, along with additional
	// nodes that need processing
	EncodeTreeNode(position *uint32, node *geodbtools.RecordTree) (b []byte, additionalNodes []*geodbtools.RecordTree, err error)

	// NewReader returns a new database reader, given generic information obtained from the source
	NewReader(source geodbtools.ReaderSource, dbType DatabaseTypeID, dbInfo string, buildTime *time.Time) (reader geodbtools.Reader, meta geodbtools.Metadata, err error)

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
	typeID = 0
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
