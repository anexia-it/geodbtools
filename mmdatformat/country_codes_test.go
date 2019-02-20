package mmdatformat

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetISO2CountryCodeIndex(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		idx, err := GetISO2CountryCodeIndex("")
		assert.EqualValues(t, 0, idx)
		assert.NoError(t, err)
	})

	t.Run("TooShort", func(t *testing.T) {
		idx, err := GetISO2CountryCodeIndex("A")
		assert.EqualValues(t, -1, idx)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("TooLong", func(t *testing.T) {
		idx, err := GetISO2CountryCodeIndex("AAA")
		assert.EqualValues(t, -1, idx)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("Unknown", func(t *testing.T) {
		idx, err := GetISO2CountryCodeIndex("??")
		assert.EqualValues(t, -1, idx)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		expectedIndexMap := make(map[string]int, len(countryCodesISO2))
		for expectedIndex, countryCode := range countryCodesISO2 {
			if _, exists := expectedIndexMap[countryCode]; !exists {
				expectedIndexMap[countryCode] = expectedIndex
			}
		}

		for countryCode, expectedIndex := range expectedIndexMap {
			t.Run(countryCode, func(t *testing.T) {
				idx, err := GetISO2CountryCodeIndex(strings.ToLower(countryCode))
				assert.EqualValues(t, expectedIndex, idx)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("Mappings", func(t *testing.T) {
		t.Parallel()

		for sourceCode, mappedCode := range countryCodesISO2Mappings {
			t.Run(fmt.Sprintf("%s->%s", sourceCode, mappedCode), func(t *testing.T) {
				expectedIndex, err := GetISO2CountryCodeIndex(strings.ToLower(mappedCode))
				require.NoError(t, err)
				idx, err := GetISO2CountryCodeIndex(strings.ToLower(sourceCode))
				assert.EqualValues(t, expectedIndex, idx)
				assert.NoError(t, err)
			})
		}
	})
}

func TestGetISO3CountryCodeIndex(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		idx, err := GetISO3CountryCodeIndex("")
		assert.EqualValues(t, 0, idx)
		assert.NoError(t, err)
	})

	t.Run("TooShort", func(t *testing.T) {
		t.Parallel()

		idx, err := GetISO3CountryCodeIndex("AA")
		assert.EqualValues(t, -1, idx)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("TooLong", func(t *testing.T) {
		t.Parallel()

		idx, err := GetISO3CountryCodeIndex("AAAA")
		assert.EqualValues(t, -1, idx)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("Unknown", func(t *testing.T) {
		t.Parallel()

		idx, err := GetISO3CountryCodeIndex("???")
		assert.EqualValues(t, -1, idx)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		expectedIndexMap := make(map[string]int, len(countryCodesISO3))
		for expectedIndex, countryCode := range countryCodesISO3 {
			if _, exists := expectedIndexMap[countryCode]; !exists {
				expectedIndexMap[countryCode] = expectedIndex
			}
		}

		for countryCode, expectedIndex := range expectedIndexMap {
			t.Run(countryCode, func(t *testing.T) {
				t.Parallel()

				idx, err := GetISO3CountryCodeIndex(strings.ToLower(countryCode))
				assert.EqualValues(t, expectedIndex, idx)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("Mappings", func(t *testing.T) {
		t.Parallel()

		for sourceCode, mappedCode := range countryCodesISO3Mappings {
			t.Run(fmt.Sprintf("%s->%s", sourceCode, mappedCode), func(t *testing.T) {
				expectedIndex, err := GetISO3CountryCodeIndex(strings.ToLower(mappedCode))
				require.NoError(t, err)
				idx, err := GetISO3CountryCodeIndex(strings.ToLower(sourceCode))
				assert.EqualValues(t, expectedIndex, idx)
				assert.NoError(t, err)
			})
		}
	})
}

func TestGetISO2CountryCodeString(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		t.Parallel()

		countryCode, err := GetISO2CountryCodeString(-1)
		assert.EqualValues(t, "", countryCode)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("TooLarge", func(t *testing.T) {
		t.Parallel()

		countryCode, err := GetISO2CountryCodeString(len(countryCodesISO2))
		assert.EqualValues(t, "", countryCode)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		for idx, expectedCountryCode := range countryCodesISO2 {
			t.Run(fmt.Sprintf("%d-%s", idx, expectedCountryCode), func(t *testing.T) {
				t.Parallel()

				countryCode, err := GetISO2CountryCodeString(idx)
				assert.EqualValues(t, expectedCountryCode, countryCode)
				assert.NoError(t, err)
			})
		}
	})
}

func TestGetISO3CountryCodeString(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		t.Parallel()

		countryCode, err := GetISO3CountryCodeString(-1)
		assert.EqualValues(t, "", countryCode)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("TooLarge", func(t *testing.T) {
		t.Parallel()

		countryCode, err := GetISO3CountryCodeString(len(countryCodesISO3))
		assert.EqualValues(t, "", countryCode)
		assert.EqualError(t, err, ErrCountryNotFound.Error())
	})

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		for idx, expectedCountryCode := range countryCodesISO3 {
			t.Run(fmt.Sprintf("%d-%s", idx, expectedCountryCode), func(t *testing.T) {
				t.Parallel()

				countryCode, err := GetISO3CountryCodeString(idx)
				assert.EqualValues(t, expectedCountryCode, countryCode)
				assert.NoError(t, err)
			})
		}
	})
}
