package mmdatformat

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package mmdatformat -self_package github.com/anexia-it/geodbtools/mmdatformat -destination mock_type_test.go github.com/anexia-it/geodbtools/mmdatformat Type
//go:generate mockgen -package mmdatformat -self_package github.com/anexia-it/geodbtools/mmdatformat -destination mock_reader_source_test.go github.com/anexia-it/geodbtools ReaderSource
//go:generate mockgen -package mmdatformat -self_package github.com/anexia-it/geodbtools/mmdatformat -destination mock_reader_test.go github.com/anexia-it/geodbtools Reader
//go:generate mockgen -package mmdatformat -self_package github.com/anexia-it/geodbtools/mmdatformat -destination mock_record_test.go github.com/anexia-it/geodbtools Record,CountryRecord
//go:generate mockgen -package mmdatformat -self_package github.com/anexia-it/geodbtools/mmdatformat -destination mock_io_test.go io Writer

type testReaderSource struct {
	*bytes.Reader
	size int64
}

func (s *testReaderSource) Size() int64 {
	return s.size
}

func (s *testReaderSource) Close() error {
	return nil
}

func TestDatFormat_FormatName(t *testing.T) {
	assert.EqualValues(t, "mmdat", format{}.FormatName())
}

func TestDatFormat_NewReaderAt(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		readerSource := NewMockReaderSource(ctrl)
		readerSource.EXPECT().Size().Return(int64(0))

		r, meta, err := format{}.NewReaderAt(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, geodbtools.ErrDatabaseInvalid.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reader := NewMockReader(ctrl)

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))
		dbType.EXPECT().NewReader(gomock.Any(), DatabaseTypeIDBase, "TestDB", nil).Return(reader, geodbtools.Metadata{
			Type: "test generated",
		}, nil)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			DatabaseTypeIDBase: dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize)
		testData = append(testData, []byte("TestDB")...)
		testData = append(testData, []byte{0x00, 0xff, 0xff, 0xff, byte(DatabaseTypeIDBase)}...)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		r, meta, err := format{}.NewReaderAt(readerSource)
		assert.NoError(t, err)
		assert.EqualValues(t, "test generated", meta.Type)
		assert.EqualValues(t, reader, r)
	})
}

func TestDatFormat_DetectFormat(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize-1)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		assert.False(t, format{}.DetectFormat(readerSource))
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reader := NewMockReader(ctrl)

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))
		dbType.EXPECT().NewReader(gomock.Any(), DatabaseTypeIDBase, "TestDB", nil).Return(reader, geodbtools.Metadata{
			Type: "test generated",
		}, nil)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			DatabaseTypeIDBase: dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize)
		testData = append(testData, []byte("TestDB")...)
		testData = append(testData, []byte{0x00, 0xff, 0xff, 0xff, byte(DatabaseTypeIDBase)}...)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		assert.True(t, format{}.DetectFormat(readerSource))
	})
}

func TestDatFormat_NewWriter(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		w, err := format{}.NewWriter(ioutil.Discard, "test", geodbtools.IPVersion4)
		assert.Nil(t, w)
		assert.EqualError(t, err, ErrTypeNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().Return(geodbtools.DatabaseType("test"))
		dbType.EXPECT().NewWriter(ioutil.Discard, geodbtools.IPVersion4).Return(nil, testErr)

		typeRegistryMu.Lock()
		originalTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			0: dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = originalTypeRegistry
		}()

		w, err := format{}.NewWriter(ioutil.Discard, "test", geodbtools.IPVersion4)
		assert.Nil(t, w)
		assert.EqualError(t, err, testErr.Error())
	})
}
