package geodbtools

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package geodbtools -self_package github.com/anexia-it/geodbtools -destination mock_format_test.go github.com/anexia-it/geodbtools Format

func TestRegisterFormat(t *testing.T) {
	t.Run("IsRegistered", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existingFormat := NewMockFormat(ctrl)
		formatRegistry["test"] = existingFormat

		newFormat := NewMockFormat(ctrl)
		newFormat.EXPECT().FormatName().Return("test")

		err := RegisterFormat(newFormat)
		assert.EqualError(t, err, ErrFormatIsRegistered.Error())

		assert.Len(t, formatRegistry, 1)
		assert.EqualValues(t, existingFormat, formatRegistry["test"])
	})

	t.Run("OK", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		newFormat := NewMockFormat(ctrl)
		newFormat.EXPECT().FormatName().Return("test")

		err := RegisterFormat(newFormat)
		assert.NoError(t, err)

		assert.Len(t, formatRegistry, 1)
		assert.EqualValues(t, newFormat, formatRegistry["test"])
	})
}

func TestMustRegisterFormat(t *testing.T) {
	t.Run("IsRegistered", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		existingFormat := NewMockFormat(ctrl)
		formatRegistry["test"] = existingFormat

		newFormat := NewMockFormat(ctrl)
		newFormat.EXPECT().FormatName().Return("test")

		assert.PanicsWithValue(t, ErrFormatIsRegistered, func() {
			MustRegisterFormat(newFormat)
		})

		assert.Len(t, formatRegistry, 1)
		assert.EqualValues(t, existingFormat, formatRegistry["test"])
	})

	t.Run("OK", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		newFormat := NewMockFormat(ctrl)
		newFormat.EXPECT().FormatName().Return("test")

		assert.NotPanics(t, func() {
			MustRegisterFormat(newFormat)
		})

		assert.Len(t, formatRegistry, 1)
		assert.EqualValues(t, newFormat, formatRegistry["test"])
	})
}

func TestFormatNames(t *testing.T) {
	formatRegistryMu.Lock()
	originalFormatRegistry := formatRegistry
	formatRegistry = make(map[string]Format)
	formatRegistryMu.Unlock()

	defer func() {
		formatRegistryMu.Lock()
		defer formatRegistryMu.Unlock()
		formatRegistry = originalFormatRegistry
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testFormat := NewMockFormat(ctrl)

	formatRegistry = map[string]Format{
		"a": testFormat,
		"b": testFormat,
		"c": testFormat,
	}

	assert.EqualValues(t, []string{"a", "b", "c"}, FormatNames())
}

func TestLookupFormat(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		test1Format := NewMockFormat(ctrl)
		formatRegistry["test1"] = test1Format

		format, err := LookupFormat("test0")
		assert.Nil(t, format)
		assert.EqualError(t, err, ErrFormatNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		test0Format := NewMockFormat(ctrl)
		test0Format.EXPECT().FormatName().Return("test0")
		formatRegistry["test0"] = test0Format

		test1Format := NewMockFormat(ctrl)
		formatRegistry["test1"] = test1Format

		format, err := LookupFormat("test0")
		assert.NoError(t, err)
		if assert.NotNil(t, format) {
			assert.EqualValues(t, test0Format, format)
			assert.EqualValues(t, "test0", format.FormatName())
		}
	})
}

func TestDetectFormat(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewReaderSourceWrapper(bytes.NewReader([]byte{0x0}), 1)

		test0Format := NewMockFormat(ctrl)
		test0Format.EXPECT().DetectFormat(r).Return(false)
		formatRegistry["test0"] = test0Format

		test1Format := NewMockFormat(ctrl)
		test1Format.EXPECT().DetectFormat(r).Return(false)
		formatRegistry["test1"] = test1Format

		format, err := DetectFormat(r)
		assert.Nil(t, format)
		assert.EqualError(t, err, ErrFormatNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		formatRegistryMu.Lock()
		originalFormatRegistry := formatRegistry
		formatRegistry = make(map[string]Format)
		formatRegistryMu.Unlock()

		defer func() {
			formatRegistryMu.Lock()
			defer formatRegistryMu.Unlock()
			formatRegistry = originalFormatRegistry
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		r := NewReaderSourceWrapper(bytes.NewReader([]byte{0x0}), 1)

		test0Format := NewMockFormat(ctrl)
		test0Format.EXPECT().DetectFormat(r).Return(true).AnyTimes()
		test0Format.EXPECT().FormatName().Return("test0")
		formatRegistry["test0"] = test0Format

		test1Format := NewMockFormat(ctrl)
		test1Format.EXPECT().DetectFormat(r).Return(false).AnyTimes()
		formatRegistry["test1"] = test1Format

		format, err := DetectFormat(r)
		assert.NoError(t, err)
		if assert.NotNil(t, format) {
			assert.EqualValues(t, test0Format, format)
			assert.EqualValues(t, "test0", format.FormatName())
		}

	})
}
