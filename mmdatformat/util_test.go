package mmdatformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeRecord(t *testing.T) {
	data := uint32(0x80402010)

	t.Run("Length1", func(t *testing.T) {
		record, err := EncodeRecord(data, 1)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{0x10}, record)
	})

	t.Run("Length2", func(t *testing.T) {
		record, err := EncodeRecord(data, 2)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{0x10, 0x20}, record)
	})

	t.Run("Length3", func(t *testing.T) {
		record, err := EncodeRecord(data, 3)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{0x10, 0x20, 0x40}, record)
	})

	t.Run("Length4", func(t *testing.T) {
		record, err := EncodeRecord(data, 4)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{0x10, 0x20, 0x40, 0x80}, record)
	})

	t.Run("Length5", func(t *testing.T) {
		record, err := EncodeRecord(data, 5)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{0x10, 0x20, 0x40, 0x80, 0x00}, record)
	})

	t.Run("Uint32Length3", func(t *testing.T) {
		record, err := EncodeRecord(uint32(8), 3)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{0x08, 0x00, 0x00}, record)
	})
}

func TestDecodeRecordUint32(t *testing.T) {
	t.Run("Length3", func(t *testing.T) {
		rec := []byte{0x10, 0x20, 0x40, 0x80}
		v, err := DecodeRecordUint32(rec, 3)
		assert.NoError(t, err)
		assert.EqualValues(t, 0x402010, v)
	})
}

func TestEncodeDecode(t *testing.T) {
	data := uint32(0x80402010)

	t.Run("Length1", func(t *testing.T) {
		expectedValue := uint32(0x10)
		record, err := EncodeRecord(data, 1)
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			val, err := DecodeRecordUint32(record, 1)
			assert.NoError(t, err)
			assert.EqualValues(t, expectedValue, val)
		}
	})

	t.Run("Length2", func(t *testing.T) {
		expectedValue := uint32(0x2010)
		record, err := EncodeRecord(data, 2)
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			val, err := DecodeRecordUint32(record, 2)
			assert.NoError(t, err)
			assert.EqualValues(t, expectedValue, val)
		}
	})

	t.Run("Length3", func(t *testing.T) {
		expectedValue := uint32(0x402010)
		record, err := EncodeRecord(data, 3)
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			val, err := DecodeRecordUint32(record, 3)
			assert.NoError(t, err)
			assert.EqualValues(t, expectedValue, val)
		}
	})

	t.Run("Length4", func(t *testing.T) {
		expectedValue := uint32(0x80402010)
		record, err := EncodeRecord(data, 4)
		assert.NoError(t, err)
		if assert.NotNil(t, record) {
			val, err := DecodeRecordUint32(record, 4)
			assert.NoError(t, err)
			assert.EqualValues(t, expectedValue, val)
		}
	})
}

func TestContainsOnlyNumericCharacters(t *testing.T) {
	t.Run("Numeric", func(t *testing.T) {
		assert.True(t, ContainsOnlyNumericCharacters("0123456789"))
	})

	t.Run("Alphanumeric", func(t *testing.T) {
		assert.False(t, ContainsOnlyNumericCharacters("01234a56789"))
	})

}
