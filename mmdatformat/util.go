package mmdatformat

import (
	"bytes"
	"encoding/binary"
	"unicode"
)

// EncodeRecord encodes the given value as a record of the given length
func EncodeRecord(v interface{}, length int) (record []byte, err error) {
	buf := bytes.NewBufferString("")
	if err = binary.Write(buf, binary.BigEndian, v); err != nil {
		return
	}

	buffer := buf.Bytes()

	record = make([]byte, length)
	for i := 0; i < length; i++ {
		idx := len(buffer) - (i + 1)
		if idx >= 0 {
			record[i] = buffer[idx]
		}
	}
	return
}

// DecodeRecordUint32 decodes a record to an uint32 value
func DecodeRecordUint32(b []byte, length int) (record uint32, err error) {
	buf := reverseBytes(b[:length])

	if len(buf) < 4 {
		buf = append(bytes.Repeat([]byte{0x00}, 4-len(buf)), buf...)
	}

	err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &record)

	return
}

func reverseBytes(b []byte) (reversed []byte) {
	reversed = make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		reversed[len(b)-(i+1)] = b[i]
	}

	return
}

// ContainsOnlyNumericCharacters checks if a given string contains
// only numeric characters (0-9)
func ContainsOnlyNumericCharacters(s string) bool {
	for _, c := range s {
		if !unicode.IsNumber(c) {
			return false
		}
	}

	return true
}
