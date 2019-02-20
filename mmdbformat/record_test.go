package mmdbformat

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountryRecord_GetRecordKey(t *testing.T) {
	t.Run("NetworkOK", func(t *testing.T) {
		_, network, err := net.ParseCIDR("127.0.0.1/32")
		require.NoError(t, err)
		assert.NotNil(t, network)

		rec := &countryRecord{
			network: network,
		}

		assert.EqualValues(t, network.IP, rec.GetRecordKey())
	})

	t.Run("NilNetwork", func(t *testing.T) {
		rec := &countryRecord{
			network: nil,
		}

		assert.EqualValues(t, net.IPv6zero, rec.GetRecordKey())
	})
}

func TestCountryRecord_GetCountryCode(t *testing.T) {
	rec := &countryRecord{}
	rec.Country.ISOCode = "TEST"
	assert.EqualValues(t, "TEST", rec.GetCountryCode())
}

func TestCountryRecord_GetNetwork(t *testing.T) {
	_, network, err := net.ParseCIDR("127.0.0.0/8")
	require.NoError(t, err)
	assert.NotNil(t, network)

	rec := &countryRecord{
		network: network,
	}

	assert.EqualValues(t, network, rec.GetNetwork())
}

func TestCountryRecord_String(t *testing.T) {
	_, network, err := net.ParseCIDR("127.0.0.127/32")
	require.NoError(t, err)
	rec := &countryRecord{
		network: network,
	}
	rec.Country.ISOCode = "XX"

	assert.EqualValues(t, "127.0.0.127/32: country code XX", rec.String())
}

func TestCountryRecord_SetNetwork(t *testing.T) {
	_, network, err := net.ParseCIDR("127.0.0.127/32")
	require.NoError(t, err)
	rec := &countryRecord{}

	rec.SetNetwork(network)
	assert.EqualValues(t, network, rec.network)
}
