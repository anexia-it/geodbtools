// Package mmdbformat implements the version 2 MaxMind database (mmdb) format
package mmdbformat

import (
	"io"
	"time"

	"github.com/anexia-it/geodbtools"
	"github.com/oschwald/maxminddb-golang"
)

var _ geodbtools.Format = format{}

type format struct {
}

func (format) FormatName() string {
	return "mmdb"
}

func (format) NewReaderAt(r geodbtools.ReaderSource) (reader geodbtools.Reader, meta geodbtools.Metadata, err error) {
	buf := make([]byte, r.Size())
	if _, err = r.ReadAt(buf, 0); err != nil {
		return
	}

	var mmdbReader *maxminddb.Reader
	if mmdbReader, err = maxminddb.FromBytes(buf); err != nil {
		return
	}

	var t Type
	if t, err = LookupTypeByDatabaseType(DatabaseTypeID(mmdbReader.Metadata.DatabaseType)); err != nil {
		return
	}

	if reader, err = t.NewReader(mmdbReader); err != nil {
		return
	}

	buildTime := time.Unix(int64(mmdbReader.Metadata.BuildEpoch), 0)
	meta = geodbtools.Metadata{
		Type:               t.DatabaseType(),
		BuildTime:          buildTime,
		Description:        mmdbReader.Metadata.Description["en"],
		MajorFormatVersion: mmdbReader.Metadata.BinaryFormatMajorVersion,
		MinorFormatVersion: mmdbReader.Metadata.BinaryFormatMinorVersion,
		IPVersion:          geodbtools.IPVersion(mmdbReader.Metadata.IPVersion),
	}

	return
}

func (format) NewWriter(w io.Writer, dbType geodbtools.DatabaseType, ipVersion geodbtools.IPVersion) (writer geodbtools.Writer, err error) {
	var t Type
	if t, _, err = LookupType(dbType); err != nil {
		return
	}

	writer, err = t.NewWriter(w, ipVersion)
	return
}

func (f format) DetectFormat(r geodbtools.ReaderSource) (isFormat bool) {
	var err error
	_, _, err = f.NewReaderAt(r)
	return err == nil
}

func init() {
	geodbtools.MustRegisterFormat(format{})
}
