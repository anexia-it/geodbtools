package mmdatformat

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/anexia-it/bitmap"
	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountryType_DatabaseType(t *testing.T) {
	assert.EqualValues(t, geodbtools.DatabaseTypeCountry, countryType{}.DatabaseType())
}

func TestCountryType_NewReader(t *testing.T) {
	t.Run("UnsupportedDBType", func(t *testing.T) {
		reader, meta, err := countryType{}.NewReader(nil, DatabaseTypeIDBase, "test", nil)
		assert.Nil(t, reader)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, geodbtools.ErrUnsupportedDatabaseType.Error())
	})

	t.Run("Country", func(t *testing.T) {
		t.Run("IPv4", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			source := NewMockReaderSource(ctrl)

			buildTime := time.Now()

			reader, meta, err := countryType{}.NewReader(source, DatabaseTypeIDCountryEdition, "test", &buildTime)
			assert.NoError(t, err)
			assert.EqualValues(t, geodbtools.Metadata{
				Type:               geodbtools.DatabaseTypeCountry,
				BuildTime:          buildTime,
				Description:        "test",
				MajorFormatVersion: 1,
				MinorFormatVersion: 0,
				IPVersion:          geodbtools.IPVersion4,
			}, meta)
			if assert.NotNil(t, reader) && assert.IsType(t, &readerCountry{}, reader) {
				r := reader.(*readerCountry)
				assert.EqualValues(t, source, r.source)
				assert.EqualValues(t, DatabaseTypeIDCountryEdition, r.dbType)
			}
		})

		t.Run("IPv6", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			source := NewMockReaderSource(ctrl)

			buildTime := time.Now()

			reader, meta, err := countryType{}.NewReader(source, DatabaseTypeIDCountryEditionV6, "test", &buildTime)
			assert.NoError(t, err)
			assert.EqualValues(t, geodbtools.Metadata{
				Type:               geodbtools.DatabaseTypeCountry,
				BuildTime:          buildTime,
				Description:        "test",
				MajorFormatVersion: 1,
				MinorFormatVersion: 0,
				IPVersion:          geodbtools.IPVersion6,
			}, meta)
			if assert.NotNil(t, reader) && assert.IsType(t, &readerCountry{}, reader) {
				r := reader.(*readerCountry)
				assert.EqualValues(t, source, r.source)
				assert.EqualValues(t, DatabaseTypeIDCountryEditionV6, r.dbType)
			}
		})

		t.Run("IPv4NilBuildTime", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			source := NewMockReaderSource(ctrl)

			reader, meta, err := countryType{}.NewReader(source, DatabaseTypeIDCountryEdition, "test", nil)
			assert.NoError(t, err)
			assert.EqualValues(t, geodbtools.DatabaseTypeCountry, meta.Type)
			assert.EqualValues(t, "test", meta.Description)
			assert.EqualValues(t, 1, meta.MajorFormatVersion)
			assert.EqualValues(t, 0, meta.MinorFormatVersion)
			assert.EqualValues(t, geodbtools.IPVersion4, meta.IPVersion)
			assert.WithinDuration(t, time.Now(), meta.BuildTime, time.Second)

			if assert.NotNil(t, reader) && assert.IsType(t, &readerCountry{}, reader) {
				r := reader.(*readerCountry)
				assert.EqualValues(t, source, r.source)
				assert.EqualValues(t, DatabaseTypeIDCountryEdition, r.dbType)
			}
		})
	})
}

func TestCountryType_NewWriter(t *testing.T) {
	t.Run("UnsupportedIPVersion", func(t *testing.T) {
		writer, err := countryType{}.NewWriter(nil, geodbtools.IPVersionUndefined)
		assert.Nil(t, writer)
		assert.EqualError(t, err, geodbtools.ErrUnsupportedDatabaseType.Error())
	})

	t.Run("IPv4", func(t *testing.T) {
		buf := bytes.NewBufferString("")
		w, err := countryType{}.NewWriter(buf, geodbtools.IPVersion4)
		assert.NoError(t, err)
		if assert.NotNil(t, w) && assert.IsType(t, &writer{}, w) {
			wr := w.(*writer)
			assert.EqualValues(t, DatabaseTypeIDCountryEdition, wr.typeID)
			assert.IsType(t, countryType{}, wr.t)
			assert.EqualValues(t, buf, wr.w)
		}
	})

	t.Run("IPv6", func(t *testing.T) {
		buf := bytes.NewBufferString("")
		w, err := countryType{}.NewWriter(buf, geodbtools.IPVersion6)
		assert.NoError(t, err)
		if assert.NotNil(t, w) && assert.IsType(t, &writer{}, w) {
			wr := w.(*writer)
			assert.EqualValues(t, DatabaseTypeIDCountryEditionV6, wr.typeID)
			assert.IsType(t, countryType{}, wr.t)
			assert.EqualValues(t, buf, wr.w)
		}
	})
}

func TestReaderCountry_RecordTree(t *testing.T) {
	t.Run("ReadAtError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).Return(-1, testErr)

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		tree, err := reader.RecordTree(geodbtools.IPVersion4)
		assert.Nil(t, tree)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("TwoRecords", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 6) {
				copy(buf, []byte{0xff, 0xff, 0xff, 0xfd, 0xff, 0xff})
			}

			n = 6
			return
		})

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		tree, err := reader.RecordTree(geodbtools.IPVersion4)
		assert.NoError(t, err)
		if assert.NotNil(t, tree) {
			records := tree.Records()
			expectedRecords := []geodbtools.Record{
				&countryRecord{
					network: &net.IPNet{
						IP:   net.ParseIP("0.0.0.0"),
						Mask: net.CIDRMask(1, 32),
					},
					countryCode: "O1",
				},
				&countryRecord{
					network: &net.IPNet{
						IP:   net.ParseIP("128.0.0.0"),
						Mask: net.CIDRMask(1, 32),
					},
					countryCode: "BQ",
				},
			}
			assert.EqualValues(t, recordStrings(expectedRecords), recordStrings(records))
		}
	})

	t.Run("TwoLevels", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 6) {
				copy(buf, []byte{0x01, 0x00, 0x00, 0x01, 0x00, 0x00})
			}

			n = 6
			return
		})
		source.EXPECT().ReadAt(gomock.Any(), int64(6)).Times(2).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 6) {
				copy(buf, []byte{0xff, 0xff, 0xff, 0xfd, 0xff, 0xff})
			}

			n = 6
			return
		})

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		tree, err := reader.RecordTree(geodbtools.IPVersion4)
		assert.NoError(t, err)
		if assert.NotNil(t, tree) {
			records := tree.Records()
			expectedRecords := []geodbtools.Record{
				&countryRecord{
					network: &net.IPNet{
						IP:   net.ParseIP("0.0.0.0"),
						Mask: net.CIDRMask(2, 32),
					},
					countryCode: "O1",
				},
				&countryRecord{
					network: &net.IPNet{
						IP:   net.ParseIP("64.0.0.0"),
						Mask: net.CIDRMask(2, 32),
					},
					countryCode: "BQ",
				},
				&countryRecord{
					network: &net.IPNet{
						IP:   net.ParseIP("128.0.0.0"),
						Mask: net.CIDRMask(2, 32),
					},
					countryCode: "O1",
				},
				&countryRecord{
					network: &net.IPNet{
						IP:   net.ParseIP("192.0.0.0"),
						Mask: net.CIDRMask(2, 32),
					},
					countryCode: "BQ",
				},
			}
			assert.EqualValues(t, recordStrings(expectedRecords), recordStrings(records))
		}
	})
}

func TestReaderCountry_LookupIP(t *testing.T) {
	t.Run("LeftReadAtError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).Return(-1, testErr)
		source.EXPECT().Size().Return(int64(6))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("127.0.0.1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("RightReadAtError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(3)).Return(-1, testErr)
		source.EXPECT().Size().Return(int64(6))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("128.0.0.1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("OneLevelOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xff, 0xff, 0xff})
			}

			n = 3
			return
		})
		source.EXPECT().ReadAt(gomock.Any(), int64(3)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xfd, 0xff, 0xff})
			}

			n = 3
			return
		})

		source.EXPECT().Size().Times(2).Return(int64(6))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("127.0.0.1"))
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			assert.EqualValues(t, "0.0.0.0/1: country code O1", record.String())
		}

		record, err = reader.LookupIP(net.ParseIP("128.0.0.1"))
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			assert.EqualValues(t, "128.0.0.0/1: country code BQ", record.String())
		}
	})

	t.Run("OneLevel6Mapped4OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xff, 0xff, 0xff})
			}

			n = 3
			return
		})
		source.EXPECT().ReadAt(gomock.Any(), int64(3)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xfd, 0xff, 0xff})
			}

			n = 3
			return
		})

		source.EXPECT().Size().Times(2).Return(int64(6))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEditionV6,
		}

		record, err := reader.LookupIP(net.ParseIP("127.0.0.1"))
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			assert.EqualValues(t, "::/1: country code O1", record.String())
		}

		record, err = reader.LookupIP(net.ParseIP("ffff::1"))
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			assert.EqualValues(t, "8000::/1: country code BQ", record.String())
		}
	})

	t.Run("TwoLevelsOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0x01, 0x00, 0x00})
			}

			n = 3
			return
		})
		source.EXPECT().ReadAt(gomock.Any(), int64(3)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0x01, 0x00, 0x00})
			}

			n = 3
			return
		})
		source.EXPECT().ReadAt(gomock.Any(), int64(6)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xff, 0xff, 0xff})
			}

			n = 3
			return
		})
		source.EXPECT().ReadAt(gomock.Any(), int64(9)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xfd, 0xff, 0xff})
			}

			n = 3
			return
		})

		source.EXPECT().Size().Times(2).Return(int64(18))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("127.0.0.1"))
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			assert.EqualValues(t, "64.0.0.0/2: country code BQ", record.String())
		}

		record, err = reader.LookupIP(net.ParseIP("128.0.0.1"))
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			assert.EqualValues(t, "128.0.0.0/2: country code O1", record.String())
		}
	})

	t.Run("RecordNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0x00, 0x00, 0x00})
			}

			n = 3
			return
		})

		source.EXPECT().Size().Return(int64(12))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("127.0.0.1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, geodbtools.ErrRecordNotFound.Error())
	})

	t.Run("IPv6LookupInIPv4DB", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("2001:db8::1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, geodbtools.ErrRecordNotFound.Error())
	})

	t.Run("InvalidDB", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0x01, 0x00, 0x00})
			}

			n = 3
			return
		})

		source.EXPECT().Size().Return(int64(6))

		reader := &readerCountry{
			source: source,
			dbType: DatabaseTypeIDCountryEdition,
		}

		record, err := reader.LookupIP(net.ParseIP("127.0.0.1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, geodbtools.ErrDatabaseInvalid.Error())
	})
}

func TestCountryType_EncodeTreeNode(t *testing.T) {
	t.Run("LeftError", func(t *testing.T) {
		t.Run("UnsupportedRecordType", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

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

			var position uint32

			b, additionalNodes, err := countryType{}.EncodeTreeNode(&position, root)
			assert.Nil(t, b)
			assert.Nil(t, additionalNodes)
			assert.EqualError(t, err, ErrUnsupportedRecordType.Error())
			assert.EqualValues(t, 0, position)
		})

		t.Run("InvalidCountryCode", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
			require.NoError(t, err)

			leftRecord := NewMockCountryRecord(ctrl)
			leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
			leftRecord.EXPECT().GetCountryCode().Return("INVALID")
			rightRecord := NewMockRecord(ctrl)
			rightRecord.EXPECT().GetNetwork().Return(rightNetwork)

			root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
			require.NoError(t, err)
			require.NotNil(t, root)

			var position uint32

			b, additionalNodes, err := countryType{}.EncodeTreeNode(&position, root)
			assert.Nil(t, b)
			assert.Nil(t, additionalNodes)
			assert.EqualError(t, err, ErrCountryNotFound.Error())
			assert.EqualValues(t, 0, position)
		})
	})

	t.Run("RightError", func(t *testing.T) {
		t.Run("UnsupportedRecordType", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
			require.NoError(t, err)

			leftRecord := NewMockCountryRecord(ctrl)
			leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
			leftRecord.EXPECT().GetCountryCode().Return("US")
			rightRecord := NewMockRecord(ctrl)
			rightRecord.EXPECT().GetNetwork().Return(rightNetwork)

			root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
			require.NoError(t, err)
			require.NotNil(t, root)

			var position uint32

			b, additionalNodes, err := countryType{}.EncodeTreeNode(&position, root)
			assert.Nil(t, b)
			assert.Nil(t, additionalNodes)
			assert.EqualError(t, err, ErrUnsupportedRecordType.Error())
			assert.EqualValues(t, 0, position)
		})

		t.Run("InvalidCountryCode", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, leftNetwork, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			_, rightNetwork, err := net.ParseCIDR("128.0.0.1/32")
			require.NoError(t, err)

			leftRecord := NewMockCountryRecord(ctrl)
			leftRecord.EXPECT().GetNetwork().Return(leftNetwork)
			leftRecord.EXPECT().GetCountryCode().Return("US")
			rightRecord := NewMockCountryRecord(ctrl)
			rightRecord.EXPECT().GetNetwork().Return(rightNetwork)
			rightRecord.EXPECT().GetCountryCode().Return("INVALID")

			root, err := geodbtools.NewRecordTree(31, []geodbtools.Record{leftRecord, rightRecord}, bitmap.IsSet)
			require.NoError(t, err)
			require.NotNil(t, root)

			var position uint32

			b, additionalNodes, err := countryType{}.EncodeTreeNode(&position, root)
			assert.Nil(t, b)
			assert.Nil(t, additionalNodes)
			assert.EqualError(t, err, ErrCountryNotFound.Error())
			assert.EqualValues(t, 0, position)
		})
	})

	t.Run("SingleLevelOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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

		var position uint32

		b, additionalNodes, err := countryType{}.EncodeTreeNode(&position, root)
		assert.EqualValues(t, []byte{0xe1, 0xff, 0xff, 0x38, 0xff, 0xff}, b)
		assert.Nil(t, additionalNodes)
		assert.NoError(t, err)
		assert.EqualValues(t, 0, position)
	})

	t.Run("TwoLevelsOK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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

		var position uint32

		b, additionalNodes, err := countryType{}.EncodeTreeNode(&position, root)
		assert.EqualValues(t, []byte{0x01, 0x00, 0x00, 0x02, 0x00, 0x00}, b)
		assert.EqualValues(t, []*geodbtools.RecordTree{root.Left(), root.Right()}, additionalNodes)
		assert.NoError(t, err)
		assert.EqualValues(t, 2, position)
	})
}

func recordStrings(records []geodbtools.Record) (s []string) {
	s = make([]string, len(records))
	for i, rec := range records {
		s[i] = rec.String()
	}

	return
}
