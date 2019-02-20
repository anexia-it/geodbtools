package mmdbformat

import (
	"net"

	"github.com/anexia-it/bitmap"
	"github.com/anexia-it/geodbtools"
	"github.com/oschwald/maxminddb-golang"
)

// Record defines the interface used by generic tree handling
type Record interface {
	geodbtools.Record

	// SetNetwork sets the network of the record
	SetNetwork(network *net.IPNet)
}

// RecordFactory defines the function type that returns a new record
type RecordFactory func() Record

// BuildRecordTree builds a record tree
func BuildRecordTree(reader *maxminddb.Reader, ipVersion geodbtools.IPVersion, factory RecordFactory) (tree *geodbtools.RecordTree, err error) {
	var maxDepth uint
	var belongsRightFunc geodbtools.RecordBelongsRightFunc

	switch ipVersion {
	case geodbtools.IPVersion6:
		if reader.Metadata.IPVersion != 6 {
			err = geodbtools.ErrUnsupportedIPVersion
			return
		}

		maxDepth = 127
		belongsRightFunc = geodbtools.RecordBelongsRightIPv6
	case geodbtools.IPVersion4:
		maxDepth = 31
		belongsRightFunc = bitmap.IsSet
	default:
		err = geodbtools.ErrUnsupportedIPVersion
		return

	}

	networks := reader.Networks()
	records := make([]geodbtools.Record, 0, reader.Metadata.NodeCount)

	for networks.Next() {
		record := factory()

		var network *net.IPNet
		if network, err = networks.Network(record); err != nil {
			return
		}
		record.SetNetwork(network)

		if ipVersion == geodbtools.IPVersion4 && network.IP.To4() == nil {
			continue
		}
		records = append(records, record)
	}

	tree, err = geodbtools.NewRecordTree(maxDepth, records, belongsRightFunc)
	return
}
