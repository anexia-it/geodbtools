package mmdatformat

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/anexia-it/bitmap"
	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_WriteDatabase(t *testing.T) {
	t.Run("EncodeTreeNodeError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		buf := bytes.NewBufferString("")

		_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		leftRecord := NewMockRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
		rightRecord := NewMockRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(rightNetwork)

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{}, root)
		assert.EqualError(t, err, ErrUnsupportedRecordType.Error())
		assert.Empty(t, buf.Bytes())
	})

	t.Run("RecordPairWriteError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		buf := NewMockWriter(ctrl)
		buf.EXPECT().Write(gomock.Any()).Return(-1, testErr)

		_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		leftRecord := NewMockCountryRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
		leftRecord.EXPECT().GetCountryCode().Return("US")
		rightRecord := NewMockCountryRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(rightNetwork)
		rightRecord.EXPECT().GetCountryCode().Return("DE")

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{}, root)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("MetadataMarkerWriteError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		buf := NewMockWriter(ctrl)
		buf.EXPECT().Write(gomock.Any()).DoAndReturn(func(buf []byte) (n int, err error) {
			assert.EqualValues(t, []byte{0xe1, 0xff, 0xff, 0x38, 0xff, 0xff}, buf)
			return
		})
		buf.EXPECT().Write(gomock.Any()).Return(-1, testErr)

		_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		leftRecord := NewMockCountryRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
		leftRecord.EXPECT().GetCountryCode().Return("US")
		rightRecord := NewMockCountryRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(rightNetwork)
		rightRecord.EXPECT().GetCountryCode().Return("DE")

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{}, root)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("MetadataWriteError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		buf := NewMockWriter(ctrl)
		buf.EXPECT().Write(gomock.Any()).DoAndReturn(func(buf []byte) (n int, err error) {
			assert.EqualValues(t, []byte{0xe1, 0xff, 0xff, 0x38, 0xff, 0xff}, buf)
			return
		})
		buf.EXPECT().Write(gomock.Any()).DoAndReturn(func(buf []byte) (n int, err error) {
			assert.EqualValues(t, []byte{0x00, 0x00, 0x00}, buf)
			return
		})
		buf.EXPECT().Write(gomock.Any()).Return(-1, testErr)

		_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		leftRecord := NewMockCountryRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
		leftRecord.EXPECT().GetCountryCode().Return("US")
		rightRecord := NewMockCountryRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(rightNetwork)
		rightRecord.EXPECT().GetCountryCode().Return("DE")

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{}, root)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("StructureInfoWriteError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		buildTime := time.Now()
		testErr := errors.New("test error")

		buf := NewMockWriter(ctrl)
		buf.EXPECT().Write(gomock.Any()).DoAndReturn(func(buf []byte) (n int, err error) {
			assert.EqualValues(t, []byte{0xe1, 0xff, 0xff, 0x38, 0xff, 0xff}, buf)
			return
		})
		buf.EXPECT().Write(gomock.Any()).DoAndReturn(func(buf []byte) (n int, err error) {
			assert.EqualValues(t, []byte{0x00, 0x00, 0x00}, buf)
			return
		})
		buf.EXPECT().Write(gomock.Any()).DoAndReturn(func(buf []byte) (n int, err error) {
			assert.EqualValues(t, []byte(fmt.Sprintf("GEO-%d %04d%02d%02d test DB", DatabaseTypeIDCountryEdition, buildTime.Year(), buildTime.Month(), buildTime.Day())), buf)
			return
		})
		buf.EXPECT().Write(gomock.Any()).Return(-1, testErr)

		_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		leftRecord := NewMockCountryRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
		leftRecord.EXPECT().GetCountryCode().Return("US")
		rightRecord := NewMockCountryRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(rightNetwork)
		rightRecord.EXPECT().GetCountryCode().Return("DE")

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{
			BuildTime:   buildTime,
			Description: "test DB",
		}, root)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("SimpleOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		buildTime := time.Now()

		buf := bytes.NewBufferString("")

		_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		leftRecord := NewMockCountryRecord(ctrl)
		leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
		leftRecord.EXPECT().GetCountryCode().Return("US")
		rightRecord := NewMockCountryRecord(ctrl)
		rightRecord.EXPECT().GetNetwork().Return(rightNetwork)
		rightRecord.EXPECT().GetCountryCode().Return("DE")

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{
			BuildTime:   buildTime,
			Description: "test DB",
		}, root)
		assert.NoError(t, err)

		expectedContents := []byte{
			0xe1, 0xff, 0xff, 0x38, 0xff, 0xff, // root record pair
			0x00, 0x00, 0x00, // metadata start marker
		}
		expectedContents = append(expectedContents,
			// metadata
			[]byte(fmt.Sprintf("GEO-%d %04d%02d%02d test DB", DatabaseTypeIDCountryEdition, buildTime.Year(), buildTime.Month(), buildTime.Day()))...,
		)
		expectedContents = append(expectedContents,
			[]byte{0xff, 0xff, 0xff, byte(DatabaseTypeIDCountryEdition)}..., // structure info
		)

		assert.EqualValues(t, expectedContents, buf.Bytes())
	})

	t.Run("MultilevelOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		buildTime := time.Now()

		buf := bytes.NewBufferString("")

		_, leftLeftNetwork, err := net.ParseCIDR("0.0.0.1/32")
		require.NoError(t, err)

		_, leftRightNetwork, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)

		_, rightLeftNetwork, err := net.ParseCIDR("128.0.0.1/32")
		require.NoError(t, err)

		_, rightRightNetwork, err := net.ParseCIDR("192.0.0.1/32")
		require.NoError(t, err)

		leftLeftRecord := NewMockCountryRecord(ctrl)
		leftLeftRecord.EXPECT().GetNetwork().AnyTimes().Return(leftLeftNetwork)
		leftLeftRecord.EXPECT().GetCountryCode().AnyTimes().Return("US")

		leftRightRecord := NewMockCountryRecord(ctrl)
		leftRightRecord.EXPECT().GetNetwork().AnyTimes().Return(leftRightNetwork)
		leftRightRecord.EXPECT().GetCountryCode().AnyTimes().Return("AT")

		rightLeftRecord := NewMockCountryRecord(ctrl)
		rightLeftRecord.EXPECT().GetNetwork().AnyTimes().Return(rightLeftNetwork)
		rightLeftRecord.EXPECT().GetCountryCode().AnyTimes().Return("DE")

		rightRightRecord := NewMockCountryRecord(ctrl)
		rightRightRecord.EXPECT().GetNetwork().AnyTimes().Return(rightRightNetwork)
		rightRightRecord.EXPECT().GetCountryCode().AnyTimes().Return("SI")

		root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftLeftRecord, leftRightRecord, rightLeftRecord, rightRightRecord}, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, root)

		w := &writer{
			w:      buf,
			t:      countryType{},
			typeID: DatabaseTypeIDCountryEdition,
		}

		err = w.WriteDatabase(geodbtools.Metadata{
			BuildTime:   buildTime,
			Description: "test DB",
		}, root)
		assert.NoError(t, err)

		expectedContents := []byte{
			0x01, 0x00, 0x00, 0x02, 0x00, 0x00, // root record pair
			0xe1, 0xff, 0xff, 0x0f, 0xff, 0xff, // left record pair
			0x38, 0xff, 0xff, 0xc2, 0xff, 0xff, // right record pair
			0x00, 0x00, 0x00, // metadata start marker
		}
		expectedContents = append(expectedContents,
			// metadata
			[]byte(fmt.Sprintf("GEO-%d %04d%02d%02d test DB", DatabaseTypeIDCountryEdition, buildTime.Year(), buildTime.Month(), buildTime.Day()))...,
		)
		expectedContents = append(expectedContents,
			[]byte{0xff, 0xff, 0xff, byte(DatabaseTypeIDCountryEdition)}..., // structure info
		)

		assert.EqualValues(t, expectedContents, buf.Bytes())
	})
}
