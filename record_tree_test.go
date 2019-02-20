package geodbtools

import (
	"net"
	"testing"

	"github.com/anexia-it/bitmap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package geodbtools -self_package github.com/anexia-it/geodbtools -destination mock_record_test.go github.com/anexia-it/geodbtools Record,CountryRecord,CityRecord

func TestRecordTree_Leaf(t *testing.T) {
	t.Run("IsLeaf", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewMockRecord(ctrl)

		tree := &RecordTree{
			records: []Record{r},
		}

		assert.EqualValues(t, r, tree.Leaf())
	})

	t.Run("NotLeaf", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r0 := NewMockRecord(ctrl)
		r1 := NewMockRecord(ctrl)

		tree := &RecordTree{
			records: []Record{r0, r1},
		}

		assert.Nil(t, tree.Leaf())
	})

}

func TestRecordTree_Left(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := NewMockRecord(ctrl)

	leftTree := &RecordTree{
		records: []Record{r},
	}

	tree := &RecordTree{
		left: leftTree,
	}

	assert.EqualValues(t, leftTree, tree.Left())
}

func TestRecordTree_Right(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := NewMockRecord(ctrl)

	rightTree := &RecordTree{
		records: []Record{r},
	}

	tree := &RecordTree{
		right: rightTree,
	}

	assert.EqualValues(t, rightTree, tree.Right())
}

func TestRecordTree_Records(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r0 := NewMockRecord(ctrl)
	r1 := NewMockRecord(ctrl)

	expectedRecords := []Record{
		r0,
		r1,
	}

	tree := &RecordTree{
		records: expectedRecords,
	}

	assert.EqualValues(t, expectedRecords, tree.Records())
}

func TestRecordTree_Build(t *testing.T) {
	t.Run("NegativeDepth", func(t *testing.T) {
		tree := &RecordTree{}
		err := tree.Build(-1, nil, nil)
		assert.EqualError(t, err, "depth<0! #records=0")
	})

	t.Run("Leaf", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewMockRecord(ctrl)
		r.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00, 0x00, 0x00, 0x00},
		})

		tree := &RecordTree{}

		assert.NoError(t, tree.Build(31, []Record{r}, bitmap.IsSet))
		if assert.NotNil(t, tree.left) {
			assert.EqualValues(t, []Record{r}, tree.left.records)
		}
		assert.Nil(t, tree.right)
		assert.EqualValues(t, []Record{r}, tree.records)
	})

	t.Run("LeafsBelowRoot", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftRecord := NewMockRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00, 0x00, 0x00, 0x00},
		})

		rightRecord := NewMockRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})

		records := []Record{
			rightRecord,
			leftRecord,
		}

		tree := &RecordTree{}
		assert.NoError(t, tree.Build(31, records, bitmap.IsSet))

		if assert.NotNil(t, tree.left) && assert.NotNil(t, tree.left.records) {
			assert.EqualValues(t, []Record{leftRecord}, tree.left.records)
		}

		if assert.NotNil(t, tree.right) && assert.NotNil(t, tree.right.records) {
			assert.EqualValues(t, []Record{rightRecord}, tree.right.records)
		}
		assert.EqualValues(t, records, tree.records)
	})

	t.Run("TopLevelPanic", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftRecord := NewMockRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00},
		})
		rightRecord := NewMockRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})
		records := []Record{
			rightRecord,
			leftRecord,
		}

		tree := &RecordTree{}
		err := tree.Build(31, records, bitmap.IsSet)
		assert.EqualError(t, err, "recovered from panic: runtime error: index out of range")
	})

	t.Run("LeftPanic", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftLeftRecord := NewMockRecord(ctrl)
		leftLeftRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x00, 0x00, 0x00, 0x00},
		})

		leftRightRecord := NewMockRecord(ctrl)
		leftRightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x40, 0x00, 0x00, 0x00},
		})
		leftRightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00},
		})

		rightLeftRecord := NewMockRecord(ctrl)
		rightLeftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})

		rightRightRecord := NewMockRecord(ctrl)
		rightRightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80 | 0x40, 0x00, 0x00, 0x00},
		})

		records := []Record{
			rightRightRecord,
			leftLeftRecord,
			leftRightRecord,
			rightLeftRecord,
		}

		tree := &RecordTree{}
		err := tree.Build(31, records, bitmap.IsSet)
		assert.EqualError(t, err, "recovered from panic: runtime error: index out of range")
	})

	t.Run("RightPanic", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftLeftRecord := NewMockRecord(ctrl)
		leftLeftRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x00, 0x00, 0x00, 0x00},
		})

		leftRightRecord := NewMockRecord(ctrl)
		leftRightRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x40, 0x00, 0x00, 0x00},
		})

		rightLeftRecord := NewMockRecord(ctrl)
		rightLeftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})
		rightLeftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00},
		})

		rightRightRecord := NewMockRecord(ctrl)
		rightRightRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x80 | 0x40, 0x00, 0x00, 0x00},
		})

		records := []Record{
			rightRightRecord,
			leftLeftRecord,
			leftRightRecord,
			rightLeftRecord,
		}

		tree := &RecordTree{}
		err := tree.Build(31, records, bitmap.IsSet)
		assert.EqualError(t, err, "recovered from panic: runtime error: index out of range")
	})

	t.Run("TwoLevels", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftLeftRecord := NewMockRecord(ctrl)
		leftLeftRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x00, 0x00, 0x00, 0x00},
		})
		leftLeftRecord.EXPECT().String().AnyTimes().Return("leftLeft")

		leftRightRecord := NewMockRecord(ctrl)
		leftRightRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x40, 0x00, 0x00, 0x00},
		})
		leftRightRecord.EXPECT().String().AnyTimes().Return("leftRight")

		rightLeftRecord := NewMockRecord(ctrl)
		rightLeftRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})
		rightLeftRecord.EXPECT().String().AnyTimes().Return("rightLeft")

		rightRightRecord := NewMockRecord(ctrl)
		rightRightRecord.EXPECT().GetNetwork().Times(2).Return(&net.IPNet{
			IP: net.IP{0x80 | 0x40, 0x00, 0x00, 0x00},
		})
		rightRightRecord.EXPECT().String().AnyTimes().Return("rightRight")

		records := []Record{
			rightRightRecord,
			leftLeftRecord,
			leftRightRecord,
			rightLeftRecord,
		}

		expectedTree := &RecordTree{
			records: records,
			left: &RecordTree{
				records: []Record{
					leftLeftRecord,
					leftRightRecord,
				},
				left: &RecordTree{
					records: []Record{leftLeftRecord},
				},
				right: &RecordTree{
					records: []Record{leftRightRecord},
				},
			},
			right: &RecordTree{
				records: []Record{
					rightRightRecord,
					rightLeftRecord,
				},
				left: &RecordTree{
					records: []Record{rightLeftRecord},
				},
				right: &RecordTree{
					records: []Record{rightRightRecord},
				},
			},
		}

		tree := &RecordTree{}
		assert.NoError(t, tree.Build(31, records, bitmap.IsSet))

		assert.EqualValues(t, expectedTree, tree)
	})
}

func TestNewRecordTree(t *testing.T) {
	t.Run("EmptyRecords", func(t *testing.T) {
		tree, err := NewRecordTree(31, nil, nil)
		assert.NoError(t, err)
		if assert.NotNil(t, tree) {
			assert.Nil(t, tree.records)
			assert.Nil(t, tree.left)
			assert.Nil(t, tree.right)
		}
	})

	t.Run("BuildError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftRecord := NewMockRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00},
		})

		rightRecord := NewMockRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})

		records := []Record{
			rightRecord,
			leftRecord,
		}

		tree, err := NewRecordTree(31, records, bitmap.IsSet)
		assert.Nil(t, tree)
		assert.EqualError(t, err, "recovered from panic: runtime error: index out of range")
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		leftRecord := NewMockRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x00, 0x00, 0x00, 0x00},
		})

		rightRecord := NewMockRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(&net.IPNet{
			IP: net.IP{0x80, 0x00, 0x00, 0x00},
		})

		records := []Record{
			rightRecord,
			leftRecord,
		}

		tree, err := NewRecordTree(31, records, bitmap.IsSet)
		assert.NoError(t, err)
		if assert.NotNil(t, tree) {

			assert.EqualValues(t, records, tree.records)
			if assert.NotNil(t, tree.left) && assert.NotNil(t, tree.left.records) {
				assert.EqualValues(t, []Record{leftRecord}, tree.left.records)
			}

			if assert.NotNil(t, tree.right) && assert.NotNil(t, tree.right.records) {
				assert.EqualValues(t, []Record{rightRecord}, tree.right.records)
			}
		}
	})
}
