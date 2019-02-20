package geodbtools

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordBelongsRightIPv6(t *testing.T) {
	t.Run("ShortSlice", func(t *testing.T) {
		b := bytes.Repeat([]byte{0xff}, 15)
		assert.False(t, RecordBelongsRightIPv6(b, 127))
	})

	t.Run("Depth0", func(t *testing.T) {
		t.Run("Left", func(t *testing.T) {
			testIP := net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fff0")
			require.NotNil(t, testIP)

			assert.False(t, RecordBelongsRightIPv6(testIP, 0))
		})

		t.Run("Right", func(t *testing.T) {
			testIP := net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0001")
			require.NotNil(t, testIP)

			assert.True(t, RecordBelongsRightIPv6(testIP, 0))
		})
	})

	t.Run("Depth127", func(t *testing.T) {
		t.Run("Left", func(t *testing.T) {
			testIP := net.ParseIP("7fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
			require.NotNil(t, testIP)

			assert.False(t, RecordBelongsRightIPv6(testIP, 127))
		})

		t.Run("Right", func(t *testing.T) {
			testIP := net.ParseIP("8000:0000:0000:0000:0000:0000:0000:0000")
			require.NotNil(t, testIP)

			assert.True(t, RecordBelongsRightIPv6(testIP, 127))
		})
	})
}
