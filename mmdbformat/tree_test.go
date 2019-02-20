package mmdbformat

import (
	"net"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anexia-it/geodbtools"
	"github.com/oschwald/maxminddb-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRecordTree(t *testing.T) {
	t.Run("IPVersionMismatch", func(t *testing.T) {
		_, testFilename, _, ok := runtime.Caller(0)
		require.True(t, ok)

		testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv4-24.mmdb")

		maxmindDB, err := maxminddb.Open(testPath)
		require.NoError(t, err)
		defer maxmindDB.Close()

		tree, err := BuildRecordTree(maxmindDB, geodbtools.IPVersion6, func() Record {
			return &countryRecord{}
		})
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

		tree, err := BuildRecordTree(maxmindDB, geodbtools.IPVersionUndefined, func() Record {
			return &countryRecord{}
		})
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

			tree, err := BuildRecordTree(maxmindDB, geodbtools.IPVersion4, func() Record {
				return &countryRecord{}
			})
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

			tree, err := BuildRecordTree(maxmindDB, geodbtools.IPVersion6, func() Record {
				return &countryRecord{}
			})
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				treeRecords := tree.Records()

				assert.Len(t, treeRecords, len(expectedRecords))
				for _, expectedRecord := range expectedRecords {
					assert.Contains(t, treeRecords, expectedRecord)
				}
			}
		})

		t.Run("4in6", func(t *testing.T) {
			_, testFilename, _, ok := runtime.Caller(0)
			require.True(t, ok)

			testPath := filepath.Join(filepath.Dir(testFilename), "test-data", "test-data", "MaxMind-DB-test-ipv6-24.mmdb")

			maxmindDB, err := maxminddb.Open(testPath)
			require.NoError(t, err)
			defer maxmindDB.Close()

			tree, err := BuildRecordTree(maxmindDB, geodbtools.IPVersion4, func() Record {
				return &countryRecord{}
			})
			assert.NoError(t, err)
			if assert.NotNil(t, tree) {
				treeRecords := tree.Records()

				assert.Len(t, treeRecords, 0)
			}
		})
	})
}
