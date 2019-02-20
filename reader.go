package geodbtools

import (
	"io"
	"net"
)

// ReaderSource defines the interface for reader source
type ReaderSource interface {
	io.ReaderAt

	// Size returns the overall size of the underlying ReaderAt
	Size() int64

	// Close frees up memory used by underlying source object
	Close() error
}

// Reader represents a database reader
type Reader interface {
	// RecordTree returns the database's underlying RecordTree instance, selecting the desired IP version.
	// Note that this is an expensive operation and is not needed for running lookups against the database.
	RecordTree(ipVersion IPVersion) (tree *RecordTree, err error)

	// LookupIP retrieves the record for the given IP address
	LookupIP(ip net.IP) (record Record, err error)
}
