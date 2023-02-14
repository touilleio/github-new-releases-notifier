package notification

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	t "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/touilleio/github-new-releases-notifier/model"
	a "golang.org/x/net/html/atom"
)

type Handler struct {
	uri              string
	notificationChan chan model.TagToNotify
	sender           *router.ServiceRouter
	log              *logrus.Logger
}

func NewNotificationHandler(uri string, notificationChan chan model.TagToNotify, log *logrus.Logger) (*Handler, error) {

	sender, err := shoutrrr.CreateSender(uri)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		uri:              uri,
		notificationChan: notificationChan,
		sender:           sender,
		log:              log,
	}

	return h, nil
}

func (h *Handler) Handle() error {

	for ttn := range h.notificationChan {
		h.log.Infof("🎉 Got a tag to notify: %s", a.Link)

		err := h.HandleSingle(ttn)

		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) HandleSingle(tagToNotify model.TagToNotify) error {

	params := &t.Params{}
	h.sender.Send(fmt.Sprintf("A new tag %s has been released for %s, direct url %s", tagToNotify.Tag, tagToNotify.ProjectUrl, tagToNotify.Link), params)

	return nil
}
