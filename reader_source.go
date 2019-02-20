package geodbtools

import (
	"io"

	"github.com/spf13/afero"
)

type readerSourceWrapper struct {
	io.ReaderAt
	size int64
}

func (w *readerSourceWrapper) Size() int64 {
	return w.size
}

func (w *readerSourceWrapper) Close() error {
	if closer, ok := w.ReaderAt.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// NewReaderSourceWrapper returns a new ReaderSource that wraps a given io.Reader
func NewReaderSourceWrapper(r io.ReaderAt, size int64) ReaderSource {
	return &readerSourceWrapper{
		ReaderAt: r,
		size:     size,
	}
}

var fs = afero.NewOsFs()

// NewFileReaderSource returns a new ReaderSource that is backed by a file
func NewFileReaderSource(path string) (s ReaderSource, err error) {
	var f afero.File
	if f, err = fs.Open(path); err != nil {
		return
	}

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	var size int64
	if size, err = f.Seek(0, io.SeekEnd); err != nil {
		return
	} else if _, err = f.Seek(0, io.SeekStart); err != nil {
		return
	}

	s = NewReaderSourceWrapper(f, size)
	return
}
