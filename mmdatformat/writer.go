package mmdatformat

import (
	"fmt"
	"io"

	"github.com/anexia-it/geodbtools"
)

var _ geodbtools.Writer = (*writer)(nil)

type writer struct {
	w      io.Writer
	t      Type
	typeID DatabaseTypeID
}

func (w *writer) WriteDatabase(meta geodbtools.Metadata, tree *geodbtools.RecordTree) (err error) {
	nodes := []*geodbtools.RecordTree{
		tree,
	}

	var currentPosition uint32

	for len(nodes) > 0 {
		cur := nodes[0]
		nodes = nodes[1:]

		var pair []byte
		var additionalNodes []*geodbtools.RecordTree
		if pair, additionalNodes, err = w.t.EncodeTreeNode(
			&currentPosition,
			cur,
		); err != nil {
			return
		}

		if _, err = w.w.Write(pair); err != nil {
			return
		}

		if len(additionalNodes) > 0 {
			nodes = append(nodes, additionalNodes...)
		}
	}

	// metadata
	if _, err = w.w.Write([]byte{0x00, 0x00, 0x00}); err != nil {
		return
	}

	metaRecord := fmt.Sprintf("GEO-%d %04d%02d%02d %s",
		w.typeID,
		meta.BuildTime.Year(),
		meta.BuildTime.Month(),
		meta.BuildTime.Day(),
		meta.Description,
	)
	if _, err = w.w.Write([]byte(metaRecord)); err != nil {
		return
	}

	// structure info
	if _, err = w.w.Write([]byte{0xff, 0xff, 0xff,
		byte(w.typeID)}); err != nil {
		return
	}

	return
}

// NewWriter returns a new writer instance
func NewWriter(w io.Writer, t Type, typeID DatabaseTypeID) geodbtools.Writer {
	return &writer{
		w:      w,
		t:      t,
		typeID: typeID,
	}
}
