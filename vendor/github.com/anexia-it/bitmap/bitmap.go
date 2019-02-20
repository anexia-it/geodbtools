// Package bitmap provides functions for working with bitmaps stored as byte slices of arbitrary size
package bitmap

import "fmt"

// IsSet checks if a bit is set inside a given byte slice
func IsSet(b []byte, bit uint) bool {
	bitNumberInByte := bit % 8
	byteIndex := uint(len(b)-1) - (bit-bitNumberInByte)/8
	targetBit := byte(1 << bitNumberInByte)

	return (b[byteIndex] & targetBit) > 0
}

// Set sets a bit inside a given byte slice
func Set(b []byte, bit uint) {
	bitNumberInByte := bit % 8
	byteIndex := uint(len(b)-1) - (bit-bitNumberInByte)/8
	targetBit := byte(1 << bitNumberInByte)

	b[byteIndex] |= targetBit
}

// Clear clears a bit inside a given byte slice
func Clear(b []byte, bit uint) {
	bitNumberInByte := bit % 8
	byteIndex := uint(len(b)-1) - (bit-bitNumberInByte)/8
	targetBit := byte(1 << bitNumberInByte)

	b[byteIndex] &= ^targetBit
}

// Mask applies the given mask to a given byte slice
func Mask(b []byte, mask []byte) (masked []byte, err error) {
	if len(b) != len(mask) {
		err = fmt.Errorf("mismatching bit lengths: %d vs %d", len(b)*8, len(mask)*8)
		return
	}

	masked = make([]byte, len(b))
	for i := uint(0); i < uint(len(b)*8); i++ {
		if IsSet(mask, i) && IsSet(b, i) {
			Set(masked, i)
		}
	}

	return
}
