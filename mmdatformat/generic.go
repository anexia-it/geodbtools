package mmdatformat

import (
	"net"
	"sync"

	"github.com/anexia-it/bitmap"
	"github.com/anexia-it/geodbtools"
)

var _ geodbtools.Reader = (*GenericReader)(nil)

// GenericReader implements the generic reader functionality for DAT databases
type GenericReader struct {
	source          geodbtools.ReaderSource
	recordLength    uint
	dbSegmentOffset uint32
	isIPv6          bool
	readerType      Type
	recordTreeMu    sync.Mutex
	recordTree      *geodbtools.RecordTree
}

type readerTrieNode struct {
	depth   uint
	offset  int64
	bitMask []byte
}

// FindRecordValue walks the internal trie and returns the stored value for the given record
func (r *GenericReader) FindRecordValue(ip net.IP) (value uint32, matchingNetwork *net.IPNet, err error) {
	maxDepth := uint(31)
	recordBelongsRight := bitmap.IsSet

	if r.isIPv6 {
		maxDepth = 127
		recordBelongsRight = geodbtools.RecordBelongsRightIPv6
		ip = ip.To16()
	} else if ip = ip.To4(); ip == nil {
		// checking a non-v4 address in a v4 tree does not make any sense
		err = geodbtools.ErrRecordNotFound
		return
	}

	rootBitMask := make([]byte, (maxDepth+1)/8)
	current := &readerTrieNode{
		depth:   0,
		offset:  0,
		bitMask: rootBitMask,
	}

	memSize := uint32(r.source.Size())

	recordPairLength := 2 * r.recordLength
	for depth := int(maxDepth); depth >= 0; depth-- {
		next := make([]byte, r.recordLength)
		nextBitMask := make([]byte, len(current.bitMask))
		copy(nextBitMask, current.bitMask)

		if recordBelongsRight(ip, uint(depth)) {
			if _, err = r.source.ReadAt(next, current.offset+int64(r.recordLength)); err != nil {
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
		if nextVal, err = DecodeRecordUint32(next, int(r.recordLength)); err != nil {
			return
		}

		if nextVal >= r.dbSegmentOffset {
			cidrMask := net.CIDRMask((int(maxDepth)-depth)+1, int(maxDepth)+1)

			var maskedIP []byte
			if maskedIP, err = bitmap.Mask(ip, cidrMask); err != nil {
				return
			}

			matchingNetwork = &net.IPNet{
				IP:   maskedIP,
				Mask: cidrMask,
			}
			value = nextVal - r.dbSegmentOffset
			return
		}

		nextOffset := int64(nextVal) * int64(recordPairLength)

		if uint32(nextOffset)+uint32(recordPairLength) >= memSize {
			err = geodbtools.ErrDatabaseInvalid
			return
		}

		current = &readerTrieNode{
			depth:   current.depth + 1,
			offset:  nextOffset,
			bitMask: nextBitMask,
		}
	}

	err = geodbtools.ErrRecordNotFound
	return
}

// LookupIP retrieves the record for a given IP address
func (r *GenericReader) LookupIP(ip net.IP) (record geodbtools.Record, err error) {
	var recordValue uint32
	var matchingNetwork *net.IPNet

	if recordValue, matchingNetwork, err = r.FindRecordValue(ip); err != nil {
		return
	}

	record, err = r.readerType.NewRecord(r.source, matchingNetwork, recordValue)
	return
}

func (r *GenericReader) buildTree() (err error) {
	maxDepth := int(31)
	recordBelongsRight := bitmap.IsSet

	if r.isIPv6 {
		maxDepth = int(127)
		recordBelongsRight = geodbtools.RecordBelongsRightIPv6
	}

	rootBitMask := make([]byte, (maxDepth+1)/8)
	rootNode := &readerTrieNode{
		depth:   0,
		offset:  0,
		bitMask: rootBitMask,
	}

	nodes := []*readerTrieNode{
		rootNode,
	}

	var records []geodbtools.Record

	recordPairLength := int64(r.recordLength * 2)

	curData := make([]byte, recordPairLength)
	for len(nodes) > 0 {
		cur := nodes[0]
		nodes = nodes[1:]

		if _, err = r.source.ReadAt(curData, cur.offset); err != nil {
			return
		}

		var left, right uint32
		if left, err = DecodeRecordUint32(curData, int(r.recordLength)); err != nil {
			return
		} else if right, err = DecodeRecordUint32(curData[r.recordLength:], int(r.recordLength)); err != nil {
			return
		}

		if left < r.dbSegmentOffset {
			bitMask := make([]byte, len(cur.bitMask))
			copy(bitMask, cur.bitMask)
			bitmap.Clear(bitMask, uint(maxDepth)-(cur.depth+1))
			nodes = append(nodes, &readerTrieNode{
				depth:   cur.depth + 1,
				offset:  int64(left) * recordPairLength,
				bitMask: bitMask,
			})
		} else {
			ip := make([]byte, len(cur.bitMask))
			copy(ip, cur.bitMask)
			recordNet := &net.IPNet{
				IP:   net.IP(ip),
				Mask: net.CIDRMask(int(cur.depth+1), int(maxDepth+1)),
			}

			var record geodbtools.Record
			if record, err = r.readerType.NewRecord(r.source, recordNet, left-r.dbSegmentOffset); err != nil {
				return
			}
			records = append(records, record)
		}

		if right < r.dbSegmentOffset {
			bitMask := make([]byte, len(cur.bitMask))
			copy(bitMask, cur.bitMask)
			bitmap.Set(bitMask, uint(maxDepth)-cur.depth)
			nodes = append(nodes, &readerTrieNode{
				depth:   cur.depth + 1,
				offset:  int64(right) * recordPairLength,
				bitMask: bitMask,
			})
		} else {

			bitMask := cur.bitMask[:]
			bitmap.Set(bitMask, uint(maxDepth)-cur.depth)
			ip := make([]byte, len(cur.bitMask))
			copy(ip, cur.bitMask)
			recordNet := &net.IPNet{
				IP:   net.IP(ip),
				Mask: net.CIDRMask(int(cur.depth+1), int(maxDepth+1)),
			}

			var record geodbtools.Record
			if record, err = r.readerType.NewRecord(r.source, recordNet, right-r.dbSegmentOffset); err != nil {
				return
			}
			records = append(records, record)
		}
	}

	r.recordTree, err = geodbtools.NewRecordTree(uint(maxDepth), records, recordBelongsRight)
	return
}

// RecordTree returns the record tree for the database
func (r *GenericReader) RecordTree(_ geodbtools.IPVersion) (tree *geodbtools.RecordTree, err error) {
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

// NewGenericReader returns a new generic reader instance
func NewGenericReader(source geodbtools.ReaderSource, databaseType Type, dbTypeID DatabaseTypeID, structInfoOffset int64, isIPv6 bool) (*GenericReader, error) {
	recordLength := databaseType.RecordLength(dbTypeID)
	if recordLength == 0 || recordLength > maxRecordLength {
		return nil, geodbtools.ErrDatabaseInvalid
	}

	return &GenericReader{
		source:          source,
		readerType:      databaseType,
		recordLength:    recordLength,
		dbSegmentOffset: databaseType.DatabaseSegmentOffset(source, dbTypeID, structInfoOffset),
		isIPv6:          isIPv6,
	}, nil
}
