package mmdatformat

import (
	"errors"
	"net"
	"testing"

	"github.com/anexia-it/bitmap"
	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenericReader(t *testing.T) {
	t.Run("IPv6", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		dbType := NewMockType(ctrl)
		dbType.EXPECT().RecordLength(DatabaseTypeIDCountryEdition).Return(uint(3))
		dbType.EXPECT().DatabaseSegmentOffset(source, DatabaseTypeIDCountryEdition, int64(-1)).Return(uint32(countryBegin))

		gr, err := NewGenericReader(source, dbType, DatabaseTypeIDCountryEdition, -1, true)
		assert.NoError(t, err)
		if assert.NotNil(t, gr) {
			assert.EqualValues(t, source, gr.source)
			assert.EqualValues(t, 3, gr.recordLength)
			assert.EqualValues(t, countryBegin, gr.dbSegmentOffset)
			assert.True(t, gr.isIPv6)
		}
	})

	t.Run("IPv4", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		dbType := NewMockType(ctrl)
		dbType.EXPECT().RecordLength(DatabaseTypeIDCountryEdition).Return(uint(4))
		dbType.EXPECT().DatabaseSegmentOffset(source, DatabaseTypeIDCountryEdition, int64(-1)).Return(uint32(0xdeadbeef))

		gr, err := NewGenericReader(source, dbType, DatabaseTypeIDCountryEdition, -1, false)
		assert.NoError(t, err)
		if assert.NotNil(t, gr) {
			assert.EqualValues(t, source, gr.source)
			assert.EqualValues(t, 4, gr.recordLength)
			assert.EqualValues(t, 0xdeadbeef, gr.dbSegmentOffset)
			assert.False(t, gr.isIPv6)
		}
	})

	t.Run("ZeroRecordLength", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		dbType := NewMockType(ctrl)
		dbType.EXPECT().RecordLength(DatabaseTypeIDCountryEdition).Return(uint(0))

		gr, err := NewGenericReader(source, dbType, DatabaseTypeIDCountryEdition, -1, false)
		assert.Nil(t, gr)
		assert.EqualError(t, err, geodbtools.ErrDatabaseInvalid.Error())
	})

	t.Run("TooLargeRecordLength", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		dbType := NewMockType(ctrl)
		dbType.EXPECT().RecordLength(DatabaseTypeIDCountryEdition).Return(uint(maxRecordLength + 1))

		gr, err := NewGenericReader(source, dbType, DatabaseTypeIDCountryEdition, -1, false)
		assert.Nil(t, gr)
		assert.EqualError(t, err, geodbtools.ErrDatabaseInvalid.Error())
	})
}

func TestGenericReader_FindRecordValue(t *testing.T) {
	t.Run("IPv6InIPv4TreeError", func(t *testing.T) {
		r := &GenericReader{
			isIPv6: false,
		}

		value, network, err := r.FindRecordValue(net.ParseIP("::1"))
		assert.EqualValues(t, 0, value)
		assert.Nil(t, network)
		assert.EqualError(t, err, geodbtools.ErrRecordNotFound.Error())
	})

	t.Run("LeftReadError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")
		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).Return(0, testErr)
		source.EXPECT().Size().Return(int64(6))

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
		}

		value, network, err := r.FindRecordValue(net.ParseIP("127.0.0.1"))
		assert.EqualValues(t, 0, value)
		assert.Nil(t, network)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("RightReadError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")
		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(3)).Return(0, testErr)
		source.EXPECT().Size().Return(int64(6))

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
		}

		value, network, err := r.FindRecordValue(net.ParseIP("128.0.0.1"))
		assert.EqualValues(t, 0, value)
		assert.Nil(t, network)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("ShortDBError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		source := NewMockReaderSource(ctrl)
		source.EXPECT().ReadAt(gomock.Any(), int64(0)).DoAndReturn(func(buf []byte, offs int64) (n int, err error) {
			if assert.Len(t, buf, 3) {
				copy(buf, []byte{0xff, 0xff, 0x00})
			}

			n = 3
			return
		})
		source.EXPECT().Size().Return(int64(6))

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
		}

		value, network, err := r.FindRecordValue(net.ParseIP("127.0.0.1"))
		assert.EqualValues(t, 0, value)
		assert.Nil(t, network)
		assert.EqualError(t, err, geodbtools.ErrDatabaseInvalid.Error())
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

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
		}

		value, network, err := r.FindRecordValue(net.ParseIP("127.0.0.1"))
		assert.EqualValues(t, 0, value)
		assert.Nil(t, network)
		assert.EqualError(t, err, geodbtools.ErrRecordNotFound.Error())
	})

	t.Run("OneLevel", func(t *testing.T) {
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

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
		}

		_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/1")
		require.NoError(t, err)
		_, expectedNetwork1, err := net.ParseCIDR("128.0.0.0/1")
		require.NoError(t, err)

		value, network, err := r.FindRecordValue(net.ParseIP("127.0.0.1"))
		assert.NoError(t, err)
		assert.EqualValues(t, uint32(0xff), value)
		assert.EqualValues(t, expectedNetwork0, network)

		value, network, err = r.FindRecordValue(net.ParseIP("128.0.0.1"))
		assert.NoError(t, err)
		assert.EqualValues(t, uint32(0xfd), value)
		assert.EqualValues(t, expectedNetwork1, network)
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

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
		}

		_, expectedNetwork0, err := net.ParseCIDR("64.0.0.0/2")
		require.NoError(t, err)
		_, expectedNetwork1, err := net.ParseCIDR("128.0.0.0/2")
		require.NoError(t, err)

		value, network, err := r.FindRecordValue(net.ParseIP("127.0.0.1"))
		assert.NoError(t, err)
		assert.EqualValues(t, uint32(0xfd), value)
		assert.EqualValues(t, expectedNetwork0, network)

		value, network, err = r.FindRecordValue(net.ParseIP("128.0.0.1"))
		assert.NoError(t, err)
		assert.EqualValues(t, uint32(0xff), value)
		assert.EqualValues(t, expectedNetwork1, network)
	})
}

func TestGenericReader_LookupIP(t *testing.T) {
	t.Run("FindRecordValueError", func(t *testing.T) {
		r := &GenericReader{
			isIPv6: false,
		}

		record, err := r.LookupIP(net.ParseIP("::1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, geodbtools.ErrRecordNotFound.Error())
	})

	t.Run("NewRecordError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		readerType := NewMockType(ctrl)

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

		_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/1")
		require.NoError(t, err)
		_, expectedNetwork1, err := net.ParseCIDR("128.0.0.0/1")
		require.NoError(t, err)

		testErr0 := errors.New("test error #0")
		testErr1 := errors.New("test error #1")
		readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(nil, testErr0)
		readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(nil, testErr1)

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
			readerType:      readerType,
		}

		record, err := r.LookupIP(net.ParseIP("127.0.0.1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, testErr0.Error())

		record, err = r.LookupIP(net.ParseIP("128.0.0.1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, testErr1.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		readerType := NewMockType(ctrl)

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

		_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/1")
		require.NoError(t, err)
		_, expectedNetwork1, err := net.ParseCIDR("128.0.0.0/1")
		require.NoError(t, err)

		expectedRecord0 := &countryRecord{
			countryCode: "00",
		}
		expectedRecord1 := &countryRecord{
			countryCode: "01",
		}
		readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(expectedRecord0, nil)
		readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(expectedRecord1, nil)

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          false,
			readerType:      readerType,
		}

		record, err := r.LookupIP(net.ParseIP("127.0.0.1"))
		assert.NoError(t, err)
		assert.EqualValues(t, expectedRecord0, record)

		record, err = r.LookupIP(net.ParseIP("128.0.0.1"))
		assert.NoError(t, err)
		assert.EqualValues(t, expectedRecord1, record)
	})

	t.Run("IPv6OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		readerType := NewMockType(ctrl)

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

		_, expectedNetwork0, err := net.ParseCIDR("::/1")
		require.NoError(t, err)
		_, expectedNetwork1, err := net.ParseCIDR("8000::/1")
		require.NoError(t, err)

		expectedRecord0 := &countryRecord{
			countryCode: "00",
		}
		expectedRecord1 := &countryRecord{
			countryCode: "01",
		}
		readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(expectedRecord0, nil)
		readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(expectedRecord1, nil)

		r := &GenericReader{
			source:          source,
			recordLength:    3,
			dbSegmentOffset: countryBegin,
			isIPv6:          true,
			readerType:      readerType,
		}

		record, err := r.LookupIP(net.ParseIP("::1"))
		assert.NoError(t, err)
		assert.EqualValues(t, expectedRecord0, record)

		record, err = r.LookupIP(net.ParseIP("8000::1"))
		assert.NoError(t, err)
		assert.EqualValues(t, expectedRecord1, record)
	})
}

func TestGenericReader_RecordTree(t *testing.T) {
	t.Run("Cached", func(t *testing.T) {
		expectedTree, err := geodbtools.NewRecordTree(31, nil, bitmap.IsSet)
		require.NoError(t, err)
		require.NotNil(t, expectedTree)

		r := &GenericReader{
			recordTree: expectedTree,
		}

		tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedTree, tree)
	})

	t.Run("IPv4", func(t *testing.T) {
		t.Run("ReadError", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			testErr := errors.New("test error")
			source := NewMockReaderSource(ctrl)
			source.EXPECT().ReadAt(gomock.Any(), int64(0)).Return(0, testErr)

			r := &GenericReader{
				source:          source,
				recordLength:    3,
				dbSegmentOffset: countryBegin,
				isIPv6:          false,
			}

			tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
			assert.Nil(t, tree)
			assert.EqualError(t, err, testErr.Error())
		})

		t.Run("OneLevel", func(t *testing.T) {
			t.Run("LeftNewRecordError", func(t *testing.T) {
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

				_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/1")
				require.NoError(t, err)

				testErr := errors.New("test error")
				readerType := NewMockType(ctrl)
				readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(nil, testErr)

				r := &GenericReader{
					source:          source,
					recordLength:    3,
					dbSegmentOffset: countryBegin,
					isIPv6:          false,
					readerType:      readerType,
				}

				tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
				assert.Nil(t, tree)
				assert.EqualError(t, err, testErr.Error())
			})

			t.Run("RightNewRecordError", func(t *testing.T) {
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

				_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/1")
				require.NoError(t, err)
				_, expectedNetwork1, err := net.ParseCIDR("128.0.0.0/1")
				require.NoError(t, err)

				testErr := errors.New("test error")
				readerType := NewMockType(ctrl)
				readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(&countryRecord{
					network:     expectedNetwork0,
					countryCode: "US",
				}, nil)
				readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(nil, testErr)

				r := &GenericReader{
					source:          source,
					recordLength:    3,
					dbSegmentOffset: countryBegin,
					isIPv6:          false,
					readerType:      readerType,
				}

				tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
				assert.Nil(t, tree)
				assert.EqualError(t, err, testErr.Error())
			})
		})

		t.Run("OK", func(t *testing.T) {
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

			_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/1")
			require.NoError(t, err)
			_, expectedNetwork1, err := net.ParseCIDR("128.0.0.0/1")
			require.NoError(t, err)

			expectedRecord0 := &countryRecord{
				network:     expectedNetwork0,
				countryCode: "00",
			}
			expectedRecord1 := &countryRecord{
				network:     expectedNetwork1,
				countryCode: "01",
			}

			readerType := NewMockType(ctrl)
			readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(expectedRecord0, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(expectedRecord1, nil)

			r := &GenericReader{
				source:          source,
				recordLength:    3,
				dbSegmentOffset: countryBegin,
				isIPv6:          false,
				readerType:      readerType,
			}

			tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				assert.EqualValues(t, []geodbtools.Record{expectedRecord0, expectedRecord1}, tree.Records())
			}
		})

		t.Run("TwoLevelsOK", func(t *testing.T) {
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

			_, expectedNetwork0, err := net.ParseCIDR("0.0.0.0/2")
			require.NoError(t, err)
			_, expectedNetwork1, err := net.ParseCIDR("64.0.0.0/2")
			require.NoError(t, err)
			_, expectedNetwork2, err := net.ParseCIDR("128.0.0.0/2")
			require.NoError(t, err)
			_, expectedNetwork3, err := net.ParseCIDR("192.0.0.0/2")
			require.NoError(t, err)

			expectedRecord0 := &countryRecord{
				network:     expectedNetwork0,
				countryCode: "00",
			}
			expectedRecord1 := &countryRecord{
				network:     expectedNetwork1,
				countryCode: "01",
			}
			expectedRecord2 := &countryRecord{
				network:     expectedNetwork2,
				countryCode: "02",
			}
			expectedRecord3 := &countryRecord{
				network:     expectedNetwork3,
				countryCode: "03",
			}

			readerType := NewMockType(ctrl)
			readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(expectedRecord0, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(expectedRecord1, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork2, uint32(0xff)).Return(expectedRecord2, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork3, uint32(0xfd)).Return(expectedRecord3, nil)

			r := &GenericReader{
				source:          source,
				recordLength:    3,
				dbSegmentOffset: countryBegin,
				isIPv6:          false,
				readerType:      readerType,
			}

			tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				assert.EqualValues(t, []geodbtools.Record{expectedRecord0, expectedRecord1, expectedRecord2, expectedRecord3}, tree.Records())
			}
		})
	})

	t.Run("IPv6", func(t *testing.T) {
		t.Run("ReadError", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			testErr := errors.New("test error")
			source := NewMockReaderSource(ctrl)
			source.EXPECT().ReadAt(gomock.Any(), int64(0)).Return(0, testErr)

			r := &GenericReader{
				source:          source,
				recordLength:    3,
				dbSegmentOffset: countryBegin,
				isIPv6:          true,
			}

			tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
			assert.Nil(t, tree)
			assert.EqualError(t, err, testErr.Error())
		})

		t.Run("OneLevel", func(t *testing.T) {
			t.Run("LeftNewRecordError", func(t *testing.T) {
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

				_, expectedNetwork0, err := net.ParseCIDR("::/1")
				require.NoError(t, err)

				testErr := errors.New("test error")
				readerType := NewMockType(ctrl)
				readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(nil, testErr)

				r := &GenericReader{
					source:          source,
					recordLength:    3,
					dbSegmentOffset: countryBegin,
					isIPv6:          true,
					readerType:      readerType,
				}

				tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
				assert.Nil(t, tree)
				assert.EqualError(t, err, testErr.Error())
			})

			t.Run("RightNewRecordError", func(t *testing.T) {
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

				_, expectedNetwork0, err := net.ParseCIDR("::/1")
				require.NoError(t, err)
				_, expectedNetwork1, err := net.ParseCIDR("8000::/1")
				require.NoError(t, err)

				testErr := errors.New("test error")
				readerType := NewMockType(ctrl)
				readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(&countryRecord{
					network:     expectedNetwork0,
					countryCode: "US",
				}, nil)
				readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(nil, testErr)

				r := &GenericReader{
					source:          source,
					recordLength:    3,
					dbSegmentOffset: countryBegin,
					isIPv6:          true,
					readerType:      readerType,
				}

				tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
				assert.Nil(t, tree)
				assert.EqualError(t, err, testErr.Error())
			})
		})

		t.Run("OK", func(t *testing.T) {
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

			_, expectedNetwork0, err := net.ParseCIDR("::/1")
			require.NoError(t, err)
			_, expectedNetwork1, err := net.ParseCIDR("8000::/1")
			require.NoError(t, err)

			expectedRecord0 := &countryRecord{
				network:     expectedNetwork0,
				countryCode: "00",
			}
			expectedRecord1 := &countryRecord{
				network:     expectedNetwork1,
				countryCode: "01",
			}

			readerType := NewMockType(ctrl)
			readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(expectedRecord0, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(expectedRecord1, nil)

			r := &GenericReader{
				source:          source,
				recordLength:    3,
				dbSegmentOffset: countryBegin,
				isIPv6:          true,
				readerType:      readerType,
			}

			tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				assert.EqualValues(t, []geodbtools.Record{expectedRecord0, expectedRecord1}, tree.Records())
			}
		})

		t.Run("TwoLevelsOK", func(t *testing.T) {
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

			_, expectedNetwork0, err := net.ParseCIDR("::/2")
			require.NoError(t, err)
			_, expectedNetwork1, err := net.ParseCIDR("4000::/2")
			require.NoError(t, err)
			_, expectedNetwork2, err := net.ParseCIDR("8000::/2")
			require.NoError(t, err)
			_, expectedNetwork3, err := net.ParseCIDR("c000::/2")
			require.NoError(t, err)

			expectedRecord0 := &countryRecord{
				network:     expectedNetwork0,
				countryCode: "00",
			}
			expectedRecord1 := &countryRecord{
				network:     expectedNetwork1,
				countryCode: "01",
			}
			expectedRecord2 := &countryRecord{
				network:     expectedNetwork2,
				countryCode: "02",
			}
			expectedRecord3 := &countryRecord{
				network:     expectedNetwork3,
				countryCode: "03",
			}

			readerType := NewMockType(ctrl)
			readerType.EXPECT().NewRecord(source, expectedNetwork0, uint32(0xff)).Return(expectedRecord0, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork1, uint32(0xfd)).Return(expectedRecord1, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork2, uint32(0xff)).Return(expectedRecord2, nil)
			readerType.EXPECT().NewRecord(source, expectedNetwork3, uint32(0xfd)).Return(expectedRecord3, nil)

			r := &GenericReader{
				source:          source,
				recordLength:    3,
				dbSegmentOffset: countryBegin,
				isIPv6:          true,
				readerType:      readerType,
			}

			tree, err := r.RecordTree(geodbtools.IPVersionUndefined)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				assert.EqualValues(t, []geodbtools.Record{expectedRecord0, expectedRecord1, expectedRecord2, expectedRecord3}, tree.Records())
			}
		})
	})
}
