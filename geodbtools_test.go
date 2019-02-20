package geodbtools

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionString(t *testing.T) {
	expectedVersionString := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	assert.EqualValues(t, expectedVersionString, VersionString())
}
