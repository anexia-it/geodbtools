package mmdatformat

import (
	"bytes"
	"strings"
	"time"
	"unicode"

	"github.com/anexia-it/geodbtools"
)

type metaReader struct {
	source    geodbtools.ReaderSource
	dbType    DatabaseTypeID
	dbInfo    string
	buildTime *time.Time
}

func (mr *metaReader) setupDatabaseInfo(dataSize int64) (err error) {
	if dataSize < databaseInfoMaxSize {
		err = geodbtools.ErrDatabaseInvalid
		return
	}

	dbInfoBytes := make([]byte, databaseInfoMaxSize)
	if _, err = mr.source.ReadAt(dbInfoBytes, dataSize-databaseInfoMaxSize); err != nil {
		return
	}

	dbInfoStart := int64(bytes.LastIndex(dbInfoBytes, []byte{0x00, 0x00, 0x00}))
	if dbInfoStart < 0 || (dbInfoStart+3) >= dataSize {
		err = ErrDatabaseInfoNotFound
		return
	}

	dbInfoStart += 3
	dbInfoBytes = dbInfoBytes[dbInfoStart:]
	var dbInfoEnd int

	for dbInfoEnd = 0; dbInfoEnd < len(dbInfoBytes); dbInfoEnd++ {
		r := rune(dbInfoBytes[dbInfoEnd])
		if !unicode.IsPrint(r) || r >= 0x7f {
			break
		}
	}

	mr.dbInfo = strings.TrimRightFunc(string(dbInfoBytes[:dbInfoEnd]), func(r rune) bool {
		return !unicode.IsPrint(r) || r >= 0x7f
	})

	if mr.dbInfo == "" {
		err = ErrDatabaseInfoNotFound
		return
	}

	// parse build time out of database info string, if possible
	for _, dbInfoPart := range strings.Split(mr.dbInfo, " ") {
		if len(dbInfoPart) == 8 && ContainsOnlyNumericCharacters(dbInfoPart) {
			if buildTime, err := time.Parse("20060102", dbInfoPart); err == nil {
				mr.buildTime = &buildTime
				break
			}
		}
	}

	if mr.buildTime == nil {
		now := time.Now()
		mr.buildTime = &now
	}

	return
}

func (mr *metaReader) setupStructInfo(dataSize int64) (structureInfoOffset int64, err error) {
	structInfoBytes := make([]byte, structureInfoMaxSize)

	if _, err = mr.source.ReadAt(structInfoBytes, dataSize-structureInfoMaxSize); err != nil {
		return
	}

	structureInfoOffset = -1
	structInfoStart := int64(bytes.LastIndex(structInfoBytes, []byte{0xff, 0xff, 0xff}))
	if structInfoStart >= 0 && (structInfoStart+3) < dataSize {
		mr.dbType = DatabaseTypeID(structInfoBytes[structInfoStart+3])
		if (structInfoStart + 4) < dataSize {
			structureInfoOffset = structInfoStart + 3
		}
	}

	return
}

func (mr *metaReader) setup() (reader geodbtools.Reader, meta geodbtools.Metadata, err error) {
	mr.dbType = DatabaseTypeIDCountryEdition
	dataSize := mr.source.Size()

	if err = mr.setupDatabaseInfo(dataSize); err != nil {
		return
	}

	var structureInfoOffset int64
	if structureInfoOffset, err = mr.setupStructInfo(dataSize); err != nil {
		return
	}

	var t Type
	if t, err = LookupTypeByDatabaseType(mr.dbType); err != nil {
		if t, err = LookupTypeByDatabaseType(mr.dbType + DatabaseTypeIDBase); err != nil {
			return
		}
		mr.dbType += DatabaseTypeIDBase
	}

	ipVersion := t.IPVersion(mr.dbType)
	if ipVersion == geodbtools.IPVersionUndefined {
		err = geodbtools.ErrUnsupportedIPVersion
		return
	}

	if reader, err = NewGenericReader(mr.source, t, mr.dbType, structureInfoOffset, ipVersion == geodbtools.IPVersion6); err != nil {
		return
	}

	meta = geodbtools.Metadata{
		Type:               t.DatabaseType(),
		BuildTime:          *mr.buildTime,
		Description:        mr.dbInfo,
		MajorFormatVersion: 1,
		MinorFormatVersion: 0,
		IPVersion:          ipVersion,
	}

	return
}

// NewReader initializes a new reader
func NewReader(r geodbtools.ReaderSource) (reader geodbtools.Reader, meta geodbtools.Metadata, err error) {
	mr := &metaReader{
		source: r,
	}

	return mr.setup()
}
