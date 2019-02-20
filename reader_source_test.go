package geodbtools

import (
	"errors"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package geodbtools -self_package github.com/anexia-it/geodbtools -destination mock_io_test.go io ReaderAt,Closer
//go:generate mockgen -package geodbtools -self_package github.com/anexia-it/geodbtools -destination mock_afero_test.go github.com/spf13/afero Fs,File

type readerAtCloser struct {
	io.ReaderAt
	io.Closer
}

func TestReaderSourceWrapper_Close(t *testing.T) {
	t.Run("NotCloser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewMockReaderAt(ctrl)

		w := &readerSourceWrapper{
			ReaderAt: r,
		}

		assert.NoError(t, w.Close())
	})

	t.Run("Closer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		r := NewMockReaderAt(ctrl)
		c := NewMockCloser(ctrl)
		c.EXPECT().Close().Return(testErr)

		w := &readerSourceWrapper{
			ReaderAt: readerAtCloser{
				ReaderAt: r,
				Closer:   c,
			},
		}

		assert.EqualError(t, w.Close(), testErr.Error())
	})
}

func TestReaderSourceWrapper_Size(t *testing.T) {
	w := &readerSourceWrapper{
		size: 1234567890,
	}

	assert.EqualValues(t, 1234567890, w.Size())
}

func TestNewReaderSourceWrapper(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedReader := NewMockReaderAt(ctrl)
	expectedSize := int64(255)

	s := NewReaderSourceWrapper(expectedReader, expectedSize)
	if assert.NotNil(t, s) && assert.IsType(t, &readerSourceWrapper{}, s) {
		w := s.(*readerSourceWrapper)
		assert.EqualValues(t, expectedReader, w.ReaderAt)
		assert.EqualValues(t, expectedSize, w.size)
	}
}

func TestNewFileReaderSource(t *testing.T) {
	t.Run("OpenError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		mockFS := NewMockFs(ctrl)
		mockFS.EXPECT().Open("/test").Return(nil, testErr)

		origFS := fs
		fs = mockFS
		defer func() {
			fs = origFS
		}()

		s, err := NewFileReaderSource("/test")
		assert.Nil(t, s)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("SeekEndError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		mockFile := NewMockFile(ctrl)
		mockFile.EXPECT().Seek(int64(0), io.SeekEnd).Return(int64(-1), testErr)
		mockFile.EXPECT().Close().Return(nil)

		mockFS := NewMockFs(ctrl)
		mockFS.EXPECT().Open("/test").Return(mockFile, nil)

		origFS := fs
		fs = mockFS
		defer func() {
			fs = origFS
		}()

		s, err := NewFileReaderSource("/test")
		assert.Nil(t, s)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("SeekStartError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		testErr := errors.New("test error")

		mockFile := NewMockFile(ctrl)
		mockFile.EXPECT().Seek(int64(0), io.SeekEnd).Return(int64(255), nil)
		mockFile.EXPECT().Seek(int64(0), io.SeekStart).Return(int64(-1), testErr)
		mockFile.EXPECT().Close().Return(nil)

		mockFS := NewMockFs(ctrl)
		mockFS.EXPECT().Open("/test").Return(mockFile, nil)

		origFS := fs
		fs = mockFS
		defer func() {
			fs = origFS
		}()

		s, err := NewFileReaderSource("/test")
		assert.Nil(t, s)
		assert.EqualError(t, err, testErr.Error())
	})

	t.Run("OK", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := NewMockFile(ctrl)
		mockFile.EXPECT().Seek(int64(0), io.SeekEnd).Return(int64(255), nil)
		mockFile.EXPECT().Seek(int64(0), io.SeekStart).Return(int64(0), nil)

		mockFS := NewMockFs(ctrl)
		mockFS.EXPECT().Open("/test").Return(mockFile, nil)

		origFS := fs
		fs = mockFS
		defer func() {
			fs = origFS
		}()

		s, err := NewFileReaderSource("/test")
		assert.NoError(t, err)
		if assert.NotNil(t, s) && assert.IsType(t, &readerSourceWrapper{}, s) {
			w := s.(*readerSourceWrapper)
			assert.EqualValues(t, mockFile, w.ReaderAt)
			assert.EqualValues(t, 255, w.size)
		}
	})
}
