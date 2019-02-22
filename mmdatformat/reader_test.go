package mmdatformat

import (
	"bytes"
	"errors"
	"testing"

	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewReader(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))

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

		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize-1)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		r, meta, err := NewReader(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, geodbtools.ErrDatabaseInvalid.Error())
	})

	t.Run("DBInfoReadError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))

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

		testErr := errors.New("test error")
		readerSource := NewMockReaderSource(ctrl)
		readerSource.EXPECT().Size().Return(int64(len(testData)))
		readerSource.EXPECT().ReadAt(gomock.Any(), int64(6)).Return(0, testErr)

		r, meta, err := NewReader(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("NoDBInfoStart", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))

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

		testData := bytes.Repeat([]byte{0xff}, databaseInfoMaxSize)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		r, meta, err := NewReader(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, ErrDatabaseInfoNotFound.Error())
	})

	t.Run("DBInfoEmpty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))

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
		testData = append(testData, 0xff)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		r, meta, err := NewReader(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, ErrDatabaseInfoNotFound.Error())
	})

	t.Run("StructInfoReadError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))

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

		testErr := errors.New("test error")
		readerSource := NewMockReaderSource(ctrl)
		readerSource.EXPECT().Size().Return(int64(len(testData)))
		readerSource.EXPECT().ReadAt(gomock.Any(), int64(6)).DoAndReturn(func(b []byte, offs int64) (n int, err error) {
			copy(b, testData[offs:])
			return
		})
		readerSource.EXPECT().ReadAt(gomock.Any(), int64(86)).Return(-1, testErr)

		r, meta, err := NewReader(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("TypeLookupError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))

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
		testData = append(testData, []byte{0x00, 0xff, 0xff, 0xff, byte(DatabaseTypeIDBase - 1)}...)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		r, meta, err := NewReader(readerSource)
		assert.Nil(t, r)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, ErrTypeNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize)
		testData = append(testData, []byte("TestDB")...)
		testData = append(testData, []byte{0x00, 0xff, 0xff, 0xff, byte(DatabaseTypeIDBase)}...)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))
		dbType.EXPECT().RecordLength(DatabaseTypeIDBase).Return(uint(3))
		dbType.EXPECT().IPVersion(DatabaseTypeIDBase).Return(geodbtools.IPVersion4)
		dbType.EXPECT().DatabaseSegmentOffset(readerSource, DatabaseTypeIDBase, int64(19)).Return(countryBegin)

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

		r, meta, err := NewReader(readerSource)
		assert.NoError(t, err)
		assert.EqualValues(t, "test", meta.Type)
		if !assert.NotNil(t, r) && assert.IsType(t, &GenericReader{}, r) {
			gr := r.(*GenericReader)
			assert.EqualValues(t, readerSource, gr.source)
			assert.EqualValues(t, 3, gr.recordLength)
			assert.EqualValues(t, dbType, gr.readerType)
			assert.EqualValues(t, countryBegin, gr.dbSegmentOffset)
			assert.False(t, gr.isIPv6)
		}
	})

	t.Run("OKDatabaseTypeOffset", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize)
		testData = append(testData, []byte("TestDB")...)
		testData = append(testData, []byte{0x00, 0xff, 0xff, 0xff, byte(DatabaseTypeIDBase - DatabaseTypeIDBase)}...)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))
		dbType.EXPECT().RecordLength(DatabaseTypeIDBase).Return(uint(3))
		dbType.EXPECT().IPVersion(DatabaseTypeIDBase).Return(geodbtools.IPVersion4)
		dbType.EXPECT().DatabaseSegmentOffset(readerSource, DatabaseTypeIDBase, int64(19)).Return(countryBegin)

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

		r, meta, err := NewReader(readerSource)
		assert.NoError(t, err)
		assert.EqualValues(t, "test", meta.Type)
		if !assert.NotNil(t, r) && assert.IsType(t, &GenericReader{}, r) {
			gr := r.(*GenericReader)
			assert.EqualValues(t, readerSource, gr.source)
			assert.EqualValues(t, 3, gr.recordLength)
			assert.EqualValues(t, dbType, gr.readerType)
			assert.EqualValues(t, countryBegin, gr.dbSegmentOffset)
			assert.False(t, gr.isIPv6)
		}
	})

	t.Run("OKWithBuildTime", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testData := bytes.Repeat([]byte{0x00}, databaseInfoMaxSize)
		testData = append(testData, []byte("TestDB 20190215")...)
		testData = append(testData, []byte{0x00, 0xff, 0xff, 0xff, byte(DatabaseTypeIDBase - DatabaseTypeIDBase)}...)

		readerSource := &testReaderSource{
			Reader: bytes.NewReader(testData),
			size:   int64(len(testData)),
		}

		dbType := NewMockType(ctrl)
		dbType.EXPECT().DatabaseType().AnyTimes().Return(geodbtools.DatabaseType("test"))
		dbType.EXPECT().RecordLength(DatabaseTypeIDBase).Return(uint(3))
		dbType.EXPECT().IPVersion(DatabaseTypeIDBase).Return(geodbtools.IPVersion4)
		dbType.EXPECT().DatabaseSegmentOffset(readerSource, DatabaseTypeIDBase, int64(19)).Return(countryBegin)

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

		r, meta, err := NewReader(readerSource)
		assert.NoError(t, err)
		assert.EqualValues(t, "test", meta.Type)
		assert.EqualValues(t, 2019, meta.BuildTime.Year())
		assert.EqualValues(t, 2, meta.BuildTime.Month())
		assert.EqualValues(t, 15, meta.BuildTime.Day())
		if !assert.NotNil(t, r) && assert.IsType(t, &GenericReader{}, r) {
			gr := r.(*GenericReader)
			assert.EqualValues(t, readerSource, gr.source)
			assert.EqualValues(t, 3, gr.recordLength)
			assert.EqualValues(t, dbType, gr.readerType)
			assert.EqualValues(t, countryBegin, gr.dbSegmentOffset)
			assert.False(t, gr.isIPv6)
		}
	})
}
