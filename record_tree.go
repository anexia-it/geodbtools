package geodbtools

import "fmt"

// RecordBelongsRightFunc tests if a given record, given the byte-slice representation of
// its IP address, belongs into the right sub-tree or not.
// This function is used during build of a RecordTree.
type RecordBelongsRightFunc func(b []byte, depth uint) bool

// RecordTree represents the rooted binary tree of records
// Each node inside the tree is either a leaf, or has two children (left and right)
type RecordTree struct {
	records []Record
	left    *RecordTree
	right   *RecordTree
}

// Leaf returns the leaf value of the tree
func (t *RecordTree) Leaf() Record {
	if len(t.records) == 1 {
		return t.records[0]
	}
	return nil
}

// Left returns the left sub-tree
func (t *RecordTree) Left() *RecordTree {
	return t.left
}

// Right returns the right sub-tree
func (t *RecordTree) Right() *RecordTree {
	return t.right
}

// Records returns all records the tree node and its children represent
func (t *RecordTree) Records() []Record {
	return t.records
}

// Build builds the sub-tree starting at the given depth, a slice of records and a RecordBelongsRightFunc
func (t *RecordTree) Build(depth int, records []Record, belongsRightFn RecordBelongsRightFunc) (err error) {
	t.records = records

	if depth < 0 {
		err = fmt.Errorf("depth<0! #records=%d", len(records))
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	recordsLeft := make([]Record, 0, len(records))
	recordsRight := make([]Record, 0, len(records))

	for _, r := range records {
		if belongsRightFn(r.GetNetwork().IP, uint(depth)) {
			recordsRight = append(recordsRight, r)
		} else {
			recordsLeft = append(recordsLeft, r)
		}
	}

	if len(recordsLeft) > 0 {
		t.left = &RecordTree{}
		if len(recordsLeft) > 1 {
			if err = t.left.Build(depth-1, recordsLeft, belongsRightFn); err != nil {
				return
			}
		} else {
			t.left.records = recordsLeft
		}
	}

	if len(recordsRight) > 0 {
		t.right = &RecordTree{}
		if len(recordsRight) > 1 {
			if err = t.right.Build(depth-1, recordsRight, belongsRightFn); err != nil {
				return
			}
		} else {
			t.right.records = recordsRight
		}
	}

	return
}

// NewRecordTree initializes and builds a new RecordTree, given a slice of records
func NewRecordTree(maxDepth uint, records []Record, belongsRightFunc RecordBelongsRightFunc) (t *RecordTree, err error) {
	t = &RecordTree{}

	if err = t.Build(int(maxDepth), records, belongsRightFunc); err != nil {
		t = nil
	}
	return
}
