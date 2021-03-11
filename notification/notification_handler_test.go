package notification

import (
	"github.com/sqooba/go-common/logging"
	"github.com/stretchr/testify/assert"
	"github.com/touilleio/github-new-releases-notifier/model"
	"testing"
)

func TestNotification(t *testing.T) {
	log := logging.NewLogger()
	logging.SetLogLevel(log, "TRACE")

	c := make(chan model.TagToNotify)
	n, err := NewNotificationHandler("slack://T016H2M724T/B01PNPGHU58/TSPxamkp5DCm46b3kegiMjuE", c, log)
	assert.Nil(t, err)

	ttn := model.TagToNotify{
		ProjectUrl: "https://github.com/touille/me",
		Tag:        "v1.2.3",
		Link:       "https://github.com/touille/me/release/v1.2.3",
		Author:     "killerwhile",
		Date:       "2021.03.05 12:12:12Z",
		RepoId:     "1234",
	}
	err = n.HandleSingle(ttn)
	assert.Nil(t, err)

}
