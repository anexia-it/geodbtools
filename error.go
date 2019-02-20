package geodbtools

import "errors"

var (
	// ErrUnsupportedDatabaseType indicates that the database type is unsupported
	ErrUnsupportedDatabaseType = errors.New("unsupported database type")

	// ErrDatabaseInvalid indicates that the database's contents are invalid
	ErrDatabaseInvalid = errors.New("database invalid")
	// ErrRecordNotFound indicates that the record for the given lookup could not be found
	ErrRecordNotFound = errors.New("record not found")
	// ErrUnsupportedIPVersion indicates that the desired IP version is not supported by the database
	ErrUnsupportedIPVersion = errors.New("requested IP version not supported by database")
)
