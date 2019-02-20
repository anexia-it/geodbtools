package geodbtools

import (
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
)

var (
	// ErrFormatIsRegistered indicates that a format with the given name has already been registered
	ErrFormatIsRegistered = errors.New("format already registered")

	// ErrFormatNotFound indicates that the format is not known
	ErrFormatNotFound = errors.New("format not found")
)

// Format represents a database format
type Format interface {
	// FormatName returns the format's name
	FormatName() string

	// NewReader returns a new reader instance from the given io.ReaderAt
	// This method will return not only a reader instance, but also the database metadata
	// retrieved from the ReaderAt instance.
	NewReaderAt(r ReaderSource) (reader Reader, meta Metadata, err error)

	// NewWriter returns a new writer instance writing to the given io.Writer
	NewWriter(w io.Writer, dbType DatabaseType, ipVersion IPVersion) (writer Writer, err error)

	// DetectFormat checks if the data represented by the passed io.ReaderAt are of the given format
	DetectFormat(r ReaderSource) (isFormat bool)
}

var formatRegistryMu sync.RWMutex
var formatRegistry = make(map[string]Format)

// RegisterFormat registers a database format
func RegisterFormat(f Format) (err error) {
	name := strings.ToLower(f.FormatName())

	formatRegistryMu.Lock()
	defer formatRegistryMu.Unlock()
	if _, exists := formatRegistry[name]; exists {
		err = ErrFormatIsRegistered
		return
	}

	formatRegistry[name] = f
	return
}

// MustRegisterFormat registers a database format and panics if the registration fails
func MustRegisterFormat(f Format) {
	if err := RegisterFormat(f); err != nil {
		panic(err)
	}
}

// FormatNames returns the names of all registered formats
func FormatNames() []string {
	formatRegistryMu.RLock()
	defer formatRegistryMu.RUnlock()

	names := make([]string, 0, len(formatRegistry))
	for name := range formatRegistry {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// LookupFormat retrieves a registered format by name
func LookupFormat(name string) (f Format, err error) {
	var exists bool

	name = strings.ToLower(name)

	formatRegistryMu.RLock()
	defer formatRegistryMu.RUnlock()
	if f, exists = formatRegistry[name]; !exists {
		err = ErrFormatNotFound
		return
	}

	return
}

// DetectFormat takes an io.ReaderAt and tries to detect the database format
func DetectFormat(r ReaderSource) (f Format, err error) {
	formatRegistryMu.RLock()
	defer formatRegistryMu.RUnlock()

	for _, format := range formatRegistry {
		if isFormat := format.DetectFormat(r); isFormat {
			f = format
			return
		}
	}

	err = ErrFormatNotFound
	return
}
