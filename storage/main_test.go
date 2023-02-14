package storage

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/sqooba/go-common/logging"
	"github.com/stretchr/testify/assert"
)

func TestPackageHandler(t *testing.T) {

	log := logging.NewLogger()
	logging.SetLogLevel(log, "DEBUG")

	dir, err := ioutil.TempDir("", "state")
	assert.Nil(t, err)

	stateHandler, err := NewStateHandler(filepath.Join(dir, "state.boltdb"), "test", log)
	assert.Nil(t, err)

	err = stateHandler.StoreFile("a")
	assert.Nil(t, err)

	exists, err := stateHandler.FileExists("a")
	assert.Nil(t, err)
	assert.True(t, exists)
	exists, err = stateHandler.FileExists("non-existing-file")
	assert.Nil(t, err)
	assert.False(t, exists)

	files, err := stateHandler.ListFiles()
	assert.Nil(t, err)
	assert.Contains(t, files, "a")

	err = stateHandler.RemoveFile("a")
	assert.Nil(t, err)

	exists, err = stateHandler.FileExists("a")
	assert.Nil(t, err)
	assert.False(t, exists)

	files, err = stateHandler.ListFiles()
	assert.Nil(t, err)
	assert.Empty(t, files)

}
