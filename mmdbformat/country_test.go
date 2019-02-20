package mmdbformat

import (
	"bytes"
	"net"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anexia-it/geodbtools"
	"github.com/oschwald/maxminddb-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountryType_DatabaseType(t *testing.T) {
	assert.EqualValues(t, geodbtools.DatabaseTypeCountry, countryType{}.DatabaseType())
}

func TestCountryType_NewWriter(t *testing.T) {
	buf := bytes.NewBufferString("")
	w, err := countryType{}.NewWriter(buf, geodbtools.IPVersion4)
	assert.Nil(t, w)
	assert.EqualError(t, err, "not implemented")
}

func TestCountryType_NewReader(t *testing.T) {
	_, testFilename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

	maxmindDB, err := maxminddb.Open(testPath)
	require.NoError(t, err)
	defer maxmindDB.Close()

	reader, err := countryType{}.NewReader(maxmindDB)
	assert.NoError(t, err)
	assert.EqualValues(t, &countryReader{
		r: maxmindDB,
	}, reader)
}

func TestCountryReader_RecordTree(t *testing.T) {
	t.Run("IPVersionMismatch", func(t *testing.T) {
		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		maxmindDB, err := maxminddb.Open(testPath)
		require.NoError(t, err)
		defer maxmindDB.Close()

		reader := &countryReader{
			r: maxmindDB,
		}

		tree, err := reader.RecordTree(geodbtools.IPVersion6)
		assert.Nil(t, tree)
		assert.EqualError(t, err, geodbtools.ErrUnsupportedIPVersion.Error())
	})

	t.Run("UnsupporyedIPVersion", func(t *testing.T) {
		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		maxmindDB, err := maxminddb.Open(testPath)
		require.NoError(t, err)
		defer maxmindDB.Close()

		reader := &countryReader{
			r: maxmindDB,
		}

		tree, err := reader.RecordTree(geodbtools.IPVersionUndefined)
		assert.Nil(t, tree)
		assert.EqualError(t, err, geodbtools.ErrUnsupportedIPVersion.Error())
	})

	t.Run("OK", func(t *testing.T) {
		t.Run("IPv4", func(t *testing.T) {
			_, testFilename, _, ok := runtime.Caller(0)
			require.True(t, ok)

			testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

			maxmindDB, err := maxminddb.Open(testPath)
			require.NoError(t, err)
			defer maxmindDB.Close()

			reader := &countryReader{
				r: maxmindDB,
			}

			var expectedRecords []geodbtools.Record
			expectedCIDRs := []string{
				"1.1.1.32/32",
				"1.1.1.16/28",
				"1.1.1.8/29",
				"1.1.1.4/30",
				"1.1.1.1/32",
				"1.1.1.2/31",
			}

			for _, cidr := range expectedCIDRs {
				_, network, err := net.ParseCIDR(cidr)
				require.NoError(t, err)
				expectedRecords = append(expectedRecords, &countryRecord{
					network: network,
				})
			}

			tree, err := reader.RecordTree(geodbtools.IPVersion4)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				treeRecords := tree.Records()
				assert.Len(t, treeRecords, len(expectedRecords))
				for _, expectedRecord := range expectedRecords {
					assert.Contains(t, treeRecords, expectedRecord)
				}
			}
		})

		t.Run("IPv6", func(t *testing.T) {
			_, testFilename, _, ok := runtime.Caller(0)
			require.True(t, ok)

			testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv6-24.mmdb")

			maxmindDB, err := maxminddb.Open(testPath)
			require.NoError(t, err)
			defer maxmindDB.Close()

			reader := &countryReader{
				r: maxmindDB,
			}

			var expectedRecords []geodbtools.Record
			expectedCIDRs := []string{
				"::1:ffff:ffff/128",
				"::2:0:0/122",
				"::2:0:40/124",
				"::2:0:50/125",
				"::2:0:58/127",
			}

			for _, cidr := range expectedCIDRs {
				_, network, err := net.ParseCIDR(cidr)
				require.NoError(t, err)
				expectedRecords = append(expectedRecords, &countryRecord{
					network: network,
				})
			}

			tree, err := reader.RecordTree(geodbtools.IPVersion6)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				assert.EqualValues(t, expectedRecords, tree.Records())
			}
		})

		t.Run("4in6", func(t *testing.T) {
			_, testFilename, _, ok := runtime.Caller(0)
			require.True(t, ok)

			testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv6-24.mmdb")

			maxmindDB, err := maxminddb.Open(testPath)
			require.NoError(t, err)
			defer maxmindDB.Close()

			reader := &countryReader{
				r: maxmindDB,
			}

			tree, err := reader.RecordTree(geodbtools.IPVersion4)
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				assert.Len(t, tree.Records(), 0)
			}
		})
	})
}

func TestCountryReader_LookupIP(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		maxmindDB, err := maxminddb.Open(testPath)
		require.NoError(t, err)
		defer maxmindDB.Close()

		reader := &countryReader{
			r: maxmindDB,
		}

		_, expectedNetwork, err := net.ParseCIDR("1.1.1.32/32")
		require.NoError(t, err)

		expectedRecord := &countryRecord{
			network: expectedNetwork,
		}

		record, err := reader.LookupIP(expectedNetwork.IP)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedRecord, record)
	})

	t.Run("LookupFailure", func(t *testing.T) {
		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		maxmindDB, err := maxminddb.Open(testPath)
		require.NoError(t, err)
		defer maxmindDB.Close()

		reader := &countryReader{
			r: maxmindDB,
		}

		record, err := reader.LookupIP(net.ParseIP("::1"))
		assert.Nil(t, record)
		assert.EqualError(t, err, "error looking up '::1': you attempted to look up an IPv6 address in an IPv4-only database")
	})
}
