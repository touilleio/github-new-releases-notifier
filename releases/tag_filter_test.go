package releases

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTagFiltering(t *testing.T) {
	matches, err := filterTag("v1.20.5-rc.0", "")
	assert.Nil(t, err)
	assert.True(t, matches)

	matches, err = filterTag("v1.20.5-rc.0", ".*")
	assert.Nil(t, err)
	assert.True(t, matches)

	matches, err = filterTag("v1.20.5-rc.0", "v?\\d+(\\.\\d+){1,2}$")
	assert.Nil(t, err)
	assert.False(t, matches)

	matches, err = filterTag("v1.20.5", "v?\\d+(\\.\\d+){1,2}$")
	assert.Nil(t, err)
	assert.True(t, matches)

	matches, err = filterTag("go1.16", "go\\d+(\\.\\d+){1,2}$")
	assert.Nil(t, err)
	assert.True(t, matches)

}
