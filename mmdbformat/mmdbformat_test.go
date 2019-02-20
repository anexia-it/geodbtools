package mmdbformat

import (
	"bytes"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/anexia-it/geodbtools"
	"github.com/golang/mock/gomock"
	"github.com/oschwald/maxminddb-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -package mmdbformat -self_package github.com/anexia-it/geodbtools/mmdbformat -destination mock_type_test.go github.com/anexia-it/geodbtools/mmdbformat Type
//go:generate mockgen -package mmdbformat -self_package github.com/anexia-it/geodbtools/mmdbformat -destination mock_reader_source_test.go github.com/anexia-it/geodbtools ReaderSource
//go:generate mockgen -package mmdbformat -self_package github.com/anexia-it/geodbtools/mmdbformat -destination mock_reader_test.go github.com/anexia-it/geodbtools Reader
//go:generate mockgen -package mmdbformat -self_package github.com/anexia-it/geodbtools/mmdbformat -destination mock_writer_test.go github.com/anexia-it/geodbtools Writer

type bufferSource struct {
	*bytes.Reader
}

func (s *bufferSource) Size() int64 {
	return int64(s.Len())
}

func (s *bufferSource) Close() error {
	return nil
}

func TestFormat_FormatName(t *testing.T) {
	assert.EqualValues(t, "mmdb", format{}.FormatName())
}

func TestFormat_NewReaderAt(t *testing.T) {
	t.Run("ReadError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		src := NewMockReaderSource(ctrl)
		src.EXPECT().Size().Return(int64(8))
		src.EXPECT().ReadAt(gomock.Any(), gomock.Any()).Do(func(buf []byte, offset int64) {
			assert.EqualValues(t, 0, offset)
			assert.Len(t, buf, 8)
			return
		}).Return(0, testErr)

		reader, meta, err := format{}.NewReaderAt(src)
		assert.Nil(t, reader)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("MaxmindDBError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		src := NewMockReaderSource(ctrl)
		src.EXPECT().Size().Return(int64(8))
		src.EXPECT().ReadAt(gomock.Any(), gomock.Any()).Do(func(buf []byte, offset int64) {
			assert.EqualValues(t, 0, offset)
			assert.Len(t, buf, 8)
			return
		}).Return(0, nil)

		reader, meta, err := format{}.NewReaderAt(src)
		assert.Nil(t, reader)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, "error opening database: invalid MaxMind DB file")
	})

	t.Run("TypeLookupError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = make(map[DatabaseTypeID]Type)
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		src, err := geodbtools.NewFileReaderSource(testPath)
		require.NoError(t, err)

		reader, meta, err := format{}.NewReaderAt(src)
		assert.Nil(t, reader)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, ErrTypeNotFound.Error())
	})

	t.Run("NewReaderError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		dbType := NewMockType(ctrl)
		dbType.EXPECT().NewReader(gomock.Any()).Return(nil, testErr)

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			"Test": dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		src, err := geodbtools.NewFileReaderSource(testPath)
		require.NoError(t, err)

		reader, meta, err := format{}.NewReaderAt(src)
		assert.Nil(t, reader)
		assert.EqualValues(t, geodbtools.Metadata{}, meta)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedReader := NewMockReader(ctrl)

		dbType := NewMockType(ctrl)
		dbType.EXPECT().NewReader(gomock.Any()).Return(expectedReader, nil)
		dbType.EXPECT().DatabaseType().Return(geodbtools.DatabaseType("test"))

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			"Test": dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		src, err := geodbtools.NewFileReaderSource(testPath)
		require.NoError(t, err)

		mmdbReader, err := maxminddb.Open(testPath)
		require.NoError(t, err)
		defer mmdbReader.Close()

		reader, meta, err := format{}.NewReaderAt(src)
		assert.EqualValues(t, expectedReader, reader)
		assert.EqualValues(t, geodbtools.Metadata{
			Type:               "test",
			BuildTime:          time.Unix(int64(mmdbReader.Metadata.BuildEpoch), 0),
			Description:        mmdbReader.Metadata.Description["en"],
			MajorFormatVersion: 2,
			MinorFormatVersion: 0,
			IPVersion:          geodbtools.IPVersion4,
		}, meta)
		assert.NoError(t, err)
	})
}

func TestFormat_NewWriter(t *testing.T) {
	t.Run("TypeNotFoundError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = make(map[DatabaseTypeID]Type)
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		w, err := format{}.NewWriter(nil, "test", geodbtools.IPVersion4)
		assert.Nil(t, w)
		assert.EqualError(t, err, ErrTypeNotFound.Error())
	})

	t.Run("NewWriterError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		dbType := NewMockType(ctrl)
		dbType.EXPECT().NewWriter(gomock.Any(), gomock.Any()).Return(nil, testErr)
		dbType.EXPECT().DatabaseType().Return(geodbtools.DatabaseType("test"))

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			"Test": dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		w, err := format{}.NewWriter(nil, "test", geodbtools.IPVersion4)
		assert.Nil(t, w)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedWriter := NewMockWriter(ctrl)

		dbType := NewMockType(ctrl)
		dbType.EXPECT().NewWriter(gomock.Any(), gomock.Any()).Return(expectedWriter, nil)
		dbType.EXPECT().DatabaseType().Return(geodbtools.DatabaseType("test"))

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			"Test": dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		w, err := format{}.NewWriter(nil, "test", geodbtools.IPVersion4)
		assert.EqualValues(t, expectedWriter, w)
		assert.NoError(t, err)

	})
}

func TestFormat_DetectFormat(t *testing.T) {
	t.Run("NoMatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		src := NewMockReaderSource(ctrl)
		src.EXPECT().Size().Return(int64(8))
		src.EXPECT().ReadAt(gomock.Any(), gomock.Any()).Do(func(buf []byte, offset int64) {
			assert.EqualValues(t, 0, offset)
			assert.Len(t, buf, 8)
			return
		}).Return(0, nil)

		isFormat := format{}.DetectFormat(src)
		assert.False(t, isFormat)
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedReader := NewMockReader(ctrl)

		dbType := NewMockType(ctrl)
		dbType.EXPECT().NewReader(gomock.Any()).Return(expectedReader, nil)
		dbType.EXPECT().DatabaseType().Return(geodbtools.DatabaseType("test"))

		typeRegistryMu.Lock()
		origTypeRegistry := typeRegistry
		typeRegistry = map[DatabaseTypeID]Type{
			"Test": dbType,
		}
		typeRegistryMu.Unlock()

		defer func() {
			typeRegistryMu.Lock()
			defer typeRegistryMu.Unlock()
			typeRegistry = origTypeRegistry
		}()

		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		src, err := geodbtools.NewFileReaderSource(testPath)
		require.NoError(t, err)

		isFormat := format{}.DetectFormat(src)
		assert.True(t, isFormat)
	})
}
