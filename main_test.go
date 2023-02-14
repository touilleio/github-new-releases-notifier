package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfigFile(t *testing.T) {

	config, err := parseConfigFile("./test-data/test-config.yml")
	assert.Nil(t, err)

	assert.Equal(t, 30*time.Minute, config.PollFrequency)
	assert.Equal(t, 3, len(config.Projects))
	assert.Equal(t, "slack://slack-uri", config.Notification.Uri)

}
