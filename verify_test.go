package geodbtools

import (
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

//go:generate mockgen -package geodbtools -self_package github.com/anexia-it/geodbtools -destination mock_reader_test.go github.com/anexia-it/geodbtools Reader

func TestRegisterEquivalentCountryCode(t *testing.T) {
	equivalentCountryCodeMapMu.Lock()
	originalCountryCodeMap := equivalentCountryCodeMap
	equivalentCountryCodeMap = map[string]string{}
	equivalentCountryCodeMapMu.Unlock()

	defer func() {
		equivalentCountryCodeMapMu.Lock()
		defer equivalentCountryCodeMapMu.Unlock()
		equivalentCountryCodeMap = originalCountryCodeMap
	}()

	RegisterEquivalentCountryCode("00", "01")
	assert.EqualValues(t, "01", equivalentCountryCodeMap["00"])
}

func TestAreCountryCodesEqual(t *testing.T) {
	t.Run("SameValues", func(t *testing.T) {
		equivalentCountryCodeMapMu.Lock()
		originalCountryCodeMap := equivalentCountryCodeMap
		equivalentCountryCodeMap = map[string]string{}
		equivalentCountryCodeMapMu.Unlock()

		defer func() {
			equivalentCountryCodeMapMu.Lock()
			defer equivalentCountryCodeMapMu.Unlock()
			equivalentCountryCodeMap = originalCountryCodeMap
		}()

		assert.True(t, AreCountryCodesEqual("00", "00"))
	})

	t.Run("Mapped", func(t *testing.T) {
		equivalentCountryCodeMapMu.Lock()
		originalCountryCodeMap := equivalentCountryCodeMap
		equivalentCountryCodeMap = map[string]string{
			"00": "01",
		}
		equivalentCountryCodeMapMu.Unlock()

		defer func() {
			equivalentCountryCodeMapMu.Lock()
			defer equivalentCountryCodeMapMu.Unlock()
			equivalentCountryCodeMap = originalCountryCodeMap
		}()

		assert.True(t, AreCountryCodesEqual("00", "01"))
		assert.True(t, AreCountryCodesEqual("01", "00"))
	})
}

func TestVerificationError_Error(t *testing.T) {
	t.Run("LookupError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedRecord := NewMockRecord(ctrl)
		expectedRecord.EXPECT().String().Return("expectedRecord")
		lookupErr := errors.New("test error")

		err := &VerificationError{
			ExpectedRecord: expectedRecord,
			LookupError:    lookupErr,
		}

		assert.EqualValues(t, "expected record expectedRecord, received error test error", err.Error())
	})

	t.Run("MismatchError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedRecord := NewMockRecord(ctrl)
		expectedRecord.EXPECT().String().Return("expectedRecord")

		record := NewMockRecord(ctrl)
		record.EXPECT().String().Return("testRecord")

		err := &VerificationError{
			ExpectedRecord: expectedRecord,
			Record:         record,
		}

		assert.EqualValues(t, "expected record expectedRecord, received record testRecord", err.Error())
	})
}

func TestVerify(t *testing.T) {
	t.Run("EmptyTree", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reader := NewMockReader(ctrl)
		root := &RecordTree{}

		assert.NoError(t, Verify(reader, root, nil))
	})

	t.Run("SingleRecord", func(t *testing.T) {
		t.Run("NilNetwork", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			record := NewMockRecord(ctrl)
			record.EXPECT().GetNetwork().Return(nil)

			reader := NewMockReader(ctrl)
			root := &RecordTree{
				records: []Record{
					record,
				},
			}

			assert.NoError(t, Verify(reader, root, nil))
		})

		t.Run("LookupError", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			record := NewMockRecord(ctrl)
			record.EXPECT().GetNetwork().Return(network)
			record.EXPECT().String().AnyTimes().Return("mockRecord")

			testErr := errors.New("test error")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network.IP).Return(nil, testErr)
			root := &RecordTree{
				records: []Record{
					record,
				},
			}

			assert.EqualError(t, Verify(reader, root, nil), "expected record mockRecord, received error test error")
		})

		t.Run("RecordNotEqual", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			record := NewMockRecord(ctrl)
			record.EXPECT().GetNetwork().Return(network)
			record.EXPECT().String().AnyTimes().Return("mockRecord")

			expectedRecord := NewMockRecord(ctrl)
			expectedRecord.EXPECT().String().AnyTimes().Return("expectedRecord")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network.IP).Return(expectedRecord, nil)
			root := &RecordTree{
				records: []Record{
					record,
				},
			}

			errs := multierr.Errors(Verify(reader, root, nil))
			if assert.Len(t, errs, 1) {
				assert.EqualError(t, errs[0], "expected record mockRecord, received record expectedRecord")
			}
		})

		t.Run("OK", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			record := NewMockCountryRecord(ctrl)
			record.EXPECT().GetNetwork().Return(network)
			record.EXPECT().String().AnyTimes().Return("mockRecord")
			record.EXPECT().GetCountryCode().AnyTimes().Return("00")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network.IP).Return(record, nil)
			root := &RecordTree{
				records: []Record{
					record,
				},
			}

			assert.NoError(t, Verify(reader, root, nil))
		})
	})

	t.Run("MultipleRecords", func(t *testing.T) {
		t.Run("NilNetwork", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			record0 := NewMockRecord(ctrl)
			record0.EXPECT().GetNetwork().Return(nil)
			record1 := NewMockCountryRecord(ctrl)
			record1.EXPECT().GetNetwork().Return(network)
			record1.EXPECT().GetCountryCode().AnyTimes().Return("00")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network.IP).Return(record1, nil)
			root := &RecordTree{
				records: []Record{
					record0,
					record1,
				},
			}

			assert.NoError(t, Verify(reader, root, nil))
		})

		t.Run("LookupError", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network0, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			_, network1, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			record0 := NewMockRecord(ctrl)
			record0.EXPECT().GetNetwork().Return(network0)
			record0.EXPECT().String().AnyTimes().Return("mockRecord")

			record1 := NewMockCountryRecord(ctrl)
			record1.EXPECT().GetNetwork().Return(network1)
			record1.EXPECT().GetCountryCode().AnyTimes().Return("00")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network0.IP).Return(nil, errors.New("test error"))
			reader.EXPECT().LookupIP(network1.IP).Return(record1, nil)
			root := &RecordTree{
				records: []Record{
					record0,
					record1,
				},
			}

			errs := multierr.Errors(Verify(reader, root, nil))
			if assert.Len(t, errs, 1) {
				assert.EqualError(t, errs[0], "expected record mockRecord, received error test error")
			}
		})

		t.Run("RecordNotEqual", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network0, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			_, network1, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			record0 := NewMockRecord(ctrl)
			record0.EXPECT().GetNetwork().Return(network0)
			record0.EXPECT().String().AnyTimes().Return("mockRecord")

			record1 := NewMockCountryRecord(ctrl)
			record1.EXPECT().GetNetwork().Return(network1)
			record1.EXPECT().GetCountryCode().AnyTimes().Return("00")

			expectedRecord0 := NewMockRecord(ctrl)
			expectedRecord0.EXPECT().String().AnyTimes().Return("expectedRecord0")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network0.IP).Return(expectedRecord0, nil)
			reader.EXPECT().LookupIP(network1.IP).Return(record1, nil)
			root := &RecordTree{
				records: []Record{
					record0,
					record1,
				},
			}

			errs := multierr.Errors(Verify(reader, root, nil))
			if assert.Len(t, errs, 1) {
				assert.EqualError(t, errs[0], "expected record mockRecord, received record expectedRecord0")
			}
		})

		t.Run("OK", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network0, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			_, network1, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			record0 := NewMockCountryRecord(ctrl)
			record0.EXPECT().GetNetwork().Return(network0)
			record0.EXPECT().GetCountryCode().AnyTimes().Return("00")

			record1 := NewMockCountryRecord(ctrl)
			record1.EXPECT().GetNetwork().Return(network1)
			record1.EXPECT().GetCountryCode().AnyTimes().Return("00")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network0.IP).Return(record0, nil)
			reader.EXPECT().LookupIP(network1.IP).Return(record1, nil)
			root := &RecordTree{
				records: []Record{
					record0,
					record1,
				},
			}

			assert.NoError(t, Verify(reader, root, nil))
		})
	})

	t.Run("ProgressReport", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, network0, err := net.ParseCIDR("127.0.0.0/8")
			require.NoError(t, err)

			_, network1, err := net.ParseCIDR("127.0.0.1/32")
			require.NoError(t, err)

			record0 := NewMockCountryRecord(ctrl)
			record0.EXPECT().GetNetwork().Return(network0)
			record0.EXPECT().GetCountryCode().AnyTimes().Return("00")

			record1 := NewMockCountryRecord(ctrl)
			record1.EXPECT().GetNetwork().Return(network1)
			record1.EXPECT().GetCountryCode().AnyTimes().Return("00")

			reader := NewMockReader(ctrl)
			reader.EXPECT().LookupIP(network0.IP).Return(record0, nil)
			reader.EXPECT().LookupIP(network1.IP).Return(record1, nil)
			root := &RecordTree{
				records: []Record{
					record0,
					record1,
				},
			}

			progress := make(chan *VerificationProgress, 8)
			defer close(progress)
			assert.NoError(t, Verify(reader, root, progress))

			for i := 0; i <= len(root.records); i++ {
				select {
				case report := <-progress:
					if assert.NotNil(t, report) {
						assert.EqualValuesf(t, i, report.CheckedRecords, "CheckedRecords incorrect #%d", i)
						assert.EqualValuesf(t, 2, report.TotalRecords, "TotalRecords incorrect #%d", i)
					}
				default:
					require.FailNowf(t, "progress report missing", "#%d missing", i)

				}
			}
		})

		t.Run("EmptyTree", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			reader := NewMockReader(ctrl)
			root := &RecordTree{
				records: []Record{},
			}

			progress := make(chan *VerificationProgress, 8)
			defer close(progress)
			assert.NoError(t, Verify(reader, root, progress))

			select {
			case report := <-progress:
				if assert.NotNil(t, report) {
					assert.EqualValues(t, 0, report.CheckedRecords)
					assert.EqualValues(t, 0, report.TotalRecords)
				}
			default:
				require.FailNow(t, "progress report missing")

			}
		})
	})
}

func TestCountryRecordsEqual(t *testing.T) {
	t.Run("BNotCountryRecord", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCountryRecord(ctrl)
		b := NewMockRecord(ctrl)

		assert.False(t, CountryRecordsEqual(a, b))
	})

	t.Run("EqualCountryCodes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCountryRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		b := NewMockCountryRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("00")

		assert.True(t, CountryRecordsEqual(a, b))
	})

	t.Run("Equal", func(t *testing.T) {
		equivalentCountryCodeMapMu.Lock()
		originalCountryCodeMap := equivalentCountryCodeMap
		equivalentCountryCodeMap = map[string]string{
			"00": "01",
		}
		equivalentCountryCodeMapMu.Unlock()

		defer func() {
			equivalentCountryCodeMapMu.Lock()
			defer equivalentCountryCodeMapMu.Unlock()
			equivalentCountryCodeMap = originalCountryCodeMap
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCountryRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		b := NewMockCountryRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("01")

		assert.True(t, CountryRecordsEqual(a, b))
	})

	t.Run("NotEqual", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCountryRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		b := NewMockCountryRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("01")

		assert.False(t, CountryRecordsEqual(a, b))
	})
}

func TestCityRecordsEqual(t *testing.T) {
	t.Run("BNotCityRecord", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCityRecord(ctrl)
		b := NewMockRecord(ctrl)

		assert.False(t, CityRecordsEqual(a, b))
	})

	t.Run("Equal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCityRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		a.EXPECT().GetCityName().Return("test city")
		b := NewMockCityRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("00")
		b.EXPECT().GetCityName().Return("test city")

		assert.True(t, CityRecordsEqual(a, b))
	})

	t.Run("CountryCodeMismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCityRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		b := NewMockCityRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("01")

		assert.False(t, CityRecordsEqual(a, b))
	})

	t.Run("EqualMapped", func(t *testing.T) {
		equivalentCountryCodeMapMu.Lock()
		originalCountryCodeMap := equivalentCountryCodeMap
		equivalentCountryCodeMap = map[string]string{
			"00": "01",
		}
		equivalentCountryCodeMapMu.Unlock()

		defer func() {
			equivalentCountryCodeMapMu.Lock()
			defer equivalentCountryCodeMapMu.Unlock()
			equivalentCountryCodeMap = originalCountryCodeMap
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCityRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		a.EXPECT().GetCityName().Return("test city")
		b := NewMockCityRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("01")
		b.EXPECT().GetCityName().Return("test city")

		assert.True(t, CityRecordsEqual(a, b))
	})

	t.Run("NotEqual", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockCityRecord(ctrl)
		a.EXPECT().GetCountryCode().Return("00")
		a.EXPECT().GetCityName().Return("test city")
		b := NewMockCityRecord(ctrl)
		b.EXPECT().GetCountryCode().Return("00")
		b.EXPECT().GetCityName().Return("test city 2")

		assert.False(t, CityRecordsEqual(a, b))
	})
}

func TestRecordsEqual(t *testing.T) {
	t.Run("CountryRecord", func(t *testing.T) {
		t.Run("Equal", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := NewMockCountryRecord(ctrl)
			a.EXPECT().GetCountryCode().Return("00")
			b := NewMockCountryRecord(ctrl)
			b.EXPECT().GetCountryCode().Return("00")

			assert.True(t, RecordsEqual(a, b))
		})

		t.Run("NotEqual", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := NewMockCountryRecord(ctrl)
			a.EXPECT().GetCountryCode().Return("00")
			b := NewMockCountryRecord(ctrl)
			b.EXPECT().GetCountryCode().Return("01")

			assert.False(t, RecordsEqual(a, b))
		})
	})

	t.Run("CityRecord", func(t *testing.T) {
		t.Run("Equal", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := NewMockCityRecord(ctrl)
			a.EXPECT().GetCountryCode().Return("00")
			a.EXPECT().GetCityName().Return("test city")
			b := NewMockCityRecord(ctrl)
			b.EXPECT().GetCountryCode().Return("00")
			b.EXPECT().GetCityName().Return("test city")

			assert.True(t, RecordsEqual(a, b))
		})

		t.Run("NotEqual", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			a := NewMockCityRecord(ctrl)
			a.EXPECT().GetCountryCode().Return("00")
			a.EXPECT().GetCityName().Return("test city")
			b := NewMockCityRecord(ctrl)
			b.EXPECT().GetCountryCode().Return("00")
			b.EXPECT().GetCityName().Return("test city 2")

			assert.False(t, RecordsEqual(a, b))
		})
	})

	t.Run("OtherRecord", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		a := NewMockRecord(ctrl)
		b := NewMockRecord(ctrl)

		assert.False(t, RecordsEqual(a, b))
	})
}
