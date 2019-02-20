package mmdatformat

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/anexia-it/bitmap"
	"github.com/anexia-it/geodbtools"
)

type readerCountryNode struct {
	depth   uint
	offset  int64
	bitMask []byte
}

type readerCountry struct {
	source geodbtools.ReaderSource
	dbType DatabaseTypeID

	recordTreeMu sync.Mutex
	recordTree   *geodbtools.RecordTree
}

func (r *readerCountry) buildTree() (err error) {
	maxDepth := int(127)
	recordBelongsRight := geodbtools.RecordBelongsRightIPv6
	if r.dbType == DatabaseTypeIDCountryEdition {
		maxDepth = 31
		recordBelongsRight = bitmap.IsSet
	}

	rootBitMask := make([]byte, (maxDepth+1)/8)
	rootNode := &readerCountryNode{
		depth:   0,
		offset:  0,
		bitMask: rootBitMask,
	}

	nodes := []*readerCountryNode{
		rootNode,
	}

	var records []geodbtools.Record

	curData := make([]byte, 6)
	for len(nodes) > 0 {
		cur := nodes[0]
		nodes = nodes[1:]

		if _, err = r.source.ReadAt(curData, cur.offset); err != nil {
			return
		}

		var left, right uint32
		if left, err = DecodeRecordUint32(curData, 3); err != nil {
			return
		} else if right, err = DecodeRecordUint32(curData[3:], 3); err != nil {
			return
		}

		if left < countryBegin {
			bitMask := make([]byte, len(cur.bitMask))
			copy(bitMask, cur.bitMask)
			bitmap.Clear(bitMask, uint(maxDepth)-(cur.depth+1))
			nodes = append(nodes, &readerCountryNode{
				depth:   cur.depth + 1,
				offset:  int64(left) * 6,
				bitMask: bitMask,
			})
		} else {
			recordCountryCode, _ := GetISO2CountryCodeString(int(left - countryBegin))
			ip := make([]byte, len(cur.bitMask))
			copy(ip, cur.bitMask)
			recordNet := &net.IPNet{
				IP:   net.IP(ip),
				Mask: net.CIDRMask(int(cur.depth+1), int(maxDepth+1)),
			}
			records = append(records, &countryRecord{
				network:     recordNet,
				countryCode: recordCountryCode,
			})
		}

		if right < countryBegin {
			bitMask := make([]byte, len(cur.bitMask))
			copy(bitMask, cur.bitMask)
			bitmap.Set(bitMask, uint(maxDepth)-cur.depth)
			nodes = append(nodes, &readerCountryNode{
				depth:   cur.depth + 1,
				offset:  int64(right) * 6,
				bitMask: bitMask,
			})
		} else {
			recordCountryCode, _ := GetISO2CountryCodeString(int(right - countryBegin))
			bitMask := cur.bitMask[:]
			bitmap.Set(bitMask, uint(maxDepth)-cur.depth)
			ip := make([]byte, len(cur.bitMask))
			copy(ip, cur.bitMask)
			recordNet := &net.IPNet{
				IP:   net.IP(ip),
				Mask: net.CIDRMask(int(cur.depth+1), int(maxDepth+1)),
			}
			records = append(records, &countryRecord{
				network:     recordNet,
				countryCode: recordCountryCode,
			})
		}
	}

	r.recordTree, err = geodbtools.NewRecordTree(uint(maxDepth), records, recordBelongsRight)
	return
}

func (r *readerCountry) RecordTree(ipVersion geodbtools.IPVersion) (tree *geodbtools.RecordTree, err error) {
	r.recordTreeMu.Lock()
	defer r.recordTreeMu.Unlock()
	if r.recordTree == nil {
		err = r.buildTree()
	}

	if err == nil {
		tree = r.recordTree
	}
	return
}

func (r *readerCountry) LookupIP(ip net.IP) (record geodbtools.Record, err error) {
	maxDepth := uint(31)
	recordBelongsRight := bitmap.IsSet

	if r.dbType == DatabaseTypeIDCountryEditionV6 {
		maxDepth = 127
		recordBelongsRight = geodbtools.RecordBelongsRightIPv6
		ip = ip.To16()
	} else if ip = ip.To4(); ip == nil {
		// checking a non-v4 address in a v4 tree does not make any sense
		err = geodbtools.ErrRecordNotFound
		return
	}

	rootBitMask := make([]byte, (maxDepth+1)/8)
	current := &readerCountryNode{
		depth:   0,
		offset:  0,
		bitMask: rootBitMask,
	}

	memSize := uint32(r.source.Size())
	for depth := int(maxDepth); depth >= 0; depth-- {
		next := make([]byte, 3)
		nextBitMask := make([]byte, len(current.bitMask))
		copy(nextBitMask, current.bitMask)

		if recordBelongsRight(ip, uint(depth)) {
			if _, err = r.source.ReadAt(next, current.offset+3); err != nil {
				return
			}
			bitmap.Set(nextBitMask, current.depth)
		} else {
			if _, err = r.source.ReadAt(next, current.offset); err != nil {
				return
			}
			bitmap.Clear(nextBitMask, current.depth)
		}

		var nextVal uint32
		if nextVal, err = DecodeRecordUint32(next, 3); err != nil {
			return
		}

		if nextVal >= countryBegin && nextVal <= countryBegin+255 {
			recordCountryCode, _ := GetISO2CountryCodeString(int(nextVal - countryBegin))

			cidrMask := net.CIDRMask((int(maxDepth)-depth)+1, int(maxDepth)+1)

			var maskedIP []byte
			if maskedIP, err = bitmap.Mask(ip, cidrMask); err != nil {
				return
			}

			matchingNetwork := &net.IPNet{
				IP:   maskedIP,
				Mask: cidrMask,
			}

			// country record found, report it back
			record = &countryRecord{
				network:     matchingNetwork,
				countryCode: recordCountryCode,
			}
			return
		}

		nextOffset := int64(nextVal * 6)

		if uint32(nextOffset+6) >= memSize {
			err = geodbtools.ErrDatabaseInvalid
			return
		}

		current = &readerCountryNode{
			depth:   current.depth + 1,
			offset:  nextOffset,
			bitMask: nextBitMask,
		}
	}

	err = geodbtools.ErrRecordNotFound
	return
}

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

func (t countryType) DatabaseType() geodbtools.DatabaseType {
	return geodbtools.DatabaseTypeCountry
}

func (t countryType) NewReader(source geodbtools.ReaderSource, dbType DatabaseTypeID, dbInfo string, buildTime *time.Time) (reader geodbtools.Reader, meta geodbtools.Metadata, err error) {
	if buildTime == nil {
		now := time.Now()
		buildTime = &now
	}

	switch dbType {
	case DatabaseTypeIDCountryEdition, DatabaseTypeIDCountryEditionV6:
		ipVersion := geodbtools.IPVersion4
		if dbType == DatabaseTypeIDCountryEditionV6 {
			ipVersion = geodbtools.IPVersion6
		}

		meta = geodbtools.Metadata{
			Type:               geodbtools.DatabaseTypeCountry,
			BuildTime:          *buildTime,
			Description:        dbInfo,
			MajorFormatVersion: 1,
			MinorFormatVersion: 0,
			IPVersion:          ipVersion,
		}

		reader = &readerCountry{
			source: source,
			dbType: dbType,
		}
		return
	}

	err = geodbtools.ErrUnsupportedDatabaseType
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
