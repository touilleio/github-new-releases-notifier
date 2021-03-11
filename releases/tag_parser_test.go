package releases

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTagParsing(t *testing.T) {
	tag, repoId, err := parseTag("tag:github.com,2008:Repository/20580498/v1.20.5-rc.0")
	assert.Nil(t, err)
	assert.Equal(t, "v1.20.5-rc.0", tag)
	assert.Equal(t, "20580498", repoId)
}
