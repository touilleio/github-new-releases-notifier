package releases

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sqooba/go-common/logging"
	"github.com/stretchr/testify/assert"
	"github.com/touilleio/github-new-releases-notifier/model"
	"github.com/touilleio/github-new-releases-notifier/storage"
)

func TestGoFeed(t *testing.T) {

	fp := gofeed.NewParser()

	projectsFeeds := []string{
		"https://github.com/golang/go/releases.atom",
		"https://github.com/kubernetes/kubernetes/releases.atom",
		"https://github.com/ethereum/go-ethereum/releases.atom",
		"https://github.com/nodejs/node/releases.atom",
		"https://github.com/apache/kafka/releases.atom",
	}

	for _, feed := range projectsFeeds {
		atomFeed, err := fp.ParseURL(feed)
		assert.Nil(t, err)
		fmt.Println(atomFeed.Title)
		for _, item := range atomFeed.Items {
			tag, _, err := parseTag(item.GUID)
			assert.Nil(t, err)
			fmt.Println(tag)
		}
	}
}

func TestNewReleasesHandler(t *testing.T) {

	log := logging.NewLogger()
	logging.SetLogLevel(log, "TRACE")

	dir, err := ioutil.TempDir("", "gh-releases-notifier")
	assert.Nil(t, err)

	storageHandler, err := storage.NewStorageHandler(filepath.Join(dir, "bolt.db"), "notifier", log)
	assert.Nil(t, err)

	notificationChan := make(chan model.TagToNotify)
	counterNewTag := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tag_count",
	})

	project := model.GithubProject{
		ProjectUrl: "https://github.com/kubernetes/kubernetes",
		TagFilter:  "v?\\d+(\\.\\d+){1,2}$",
	}

	projects := []model.GithubProject{project}

	handler, err := NewReleasesHandler(projects, 1*time.Hour, storageHandler, counterNewTag,
		notificationChan, log)
	assert.Nil(t, err)

	err = handler.RunSingle(project, false)
	assert.Nil(t, err)

	tags, err := storageHandler.ListTags()
	assert.Nil(t, err)
	log.Infof("%v", tags)

	err = handler.RunSingle(project, false)
	assert.Nil(t, err)

	tags, err = storageHandler.ListTags()
	assert.Nil(t, err)
	log.Infof("%v", tags)

}
