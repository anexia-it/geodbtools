// Package mmdatformat implements the version 1 MaxMind database (dat) format
package mmdatformat

import (
	"errors"
	"io"

	"github.com/anexia-it/geodbtools"
)

var (
	// ErrUnsupportedRecordType indicates that a record type is unsupported
	ErrUnsupportedRecordType = errors.New("unsupported record type")
	// ErrDatabaseInfoNotFound indicates that the database information could not be found
	ErrDatabaseInfoNotFound = errors.New("database information not found")
)

const (
	structureInfoMaxSize = 20
	databaseInfoMaxSize  = 100
	maxRecordLength      = 4
)

var _ geodbtools.Format = format{}

type format struct{}

func (format) FormatName() string {
	return "mmdat"
}

func (format) NewReaderAt(r geodbtools.ReaderSource) (reader geodbtools.Reader, meta geodbtools.Metadata, err error) {
	return NewReader(r)
}

func (format) NewWriter(w io.Writer, dbType geodbtools.DatabaseType, ipVersion geodbtools.IPVersion) (writer geodbtools.Writer, err error) {
	var t Type

	if t, _, err = LookupType(dbType); err != nil {
		return
	}

	return t.NewWriter(w, ipVersion)
}

func (f format) DetectFormat(r geodbtools.ReaderSource) (isFormat bool) {
	_, _, err := f.NewReaderAt(r)
	return err == nil
}

func init() {
	geodbtools.MustRegisterFormat(format{})
}
