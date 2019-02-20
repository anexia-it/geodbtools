package mmdatformat

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountryRecord_GetNetwork(t *testing.T) {
	_, network, err := net.ParseCIDR("127.0.0.0/8")
	require.NoError(t, err)
	rec := &countryRecord{
		network: network,
	}

	assert.EqualValues(t, network, rec.GetNetwork())
}

func TestCountryRecord_GetRecordKey(t *testing.T) {
	_, network, err := net.ParseCIDR("127.1.2.3/32")
	require.NoError(t, err)
	rec := &countryRecord{
		network: network,
	}

	assert.EqualValues(t, network.IP, rec.GetRecordKey())
}

func TestCountryRecord_GetCountryCode(t *testing.T) {
	rec := &countryRecord{
		countryCode: "XX",
	}

	assert.EqualValues(t, "XX", rec.GetCountryCode())
}

func TestCountryRecord_String(t *testing.T) {
	_, network, err := net.ParseCIDR("127.0.0.127/32")
	require.NoError(t, err)
	rec := &countryRecord{
		network:     network,
		countryCode: "XX",
	}

	assert.EqualValues(t, "127.0.0.127/32: country code XX", rec.String())
}
