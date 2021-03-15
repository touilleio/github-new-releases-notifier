package releases

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/touilleio/github-new-releases-notifier/model"
	"github.com/touilleio/github-new-releases-notifier/storage"
	"regexp"
	"time"
)

type Handler struct {
	projects         []model.GithubProject
	pollFrequency    time.Duration
	storageHandler   *storage.Handler
	newTagCounter    prometheus.Counter
	feedParser       *gofeed.Parser
	notificationChan chan model.TagToNotify
	notifyAllTags    bool
	log              *logrus.Logger
}

func NewReleasesHandler(projects []model.GithubProject, pollFrequency time.Duration,
	notifyAllTag bool,
	storageHandler *storage.Handler, counterNewTag prometheus.Counter,
	notificationChan chan model.TagToNotify,
	log *logrus.Logger) (*Handler, error) {

	// validate configuration
	projectUrlRe := regexp.MustCompile("https://github.com/.+/.+")
	for _, p := range projects {
		if !projectUrlRe.MatchString(p.ProjectUrl) {
			return nil, fmt.Errorf("project URI %s do not match https://github.com/.+/.+ pattern", p.ProjectUrl)
		}
		_, err := regexp.Compile(p.TagFilter)
		if err != nil {
			return nil, err
		}

		_, err = regexp.Compile(p.TitleFilter)
		if err != nil {
			return nil, err
		}
	}

	h := &Handler{
		projects:         projects,
		pollFrequency:    pollFrequency,
		notifyAllTags:    notifyAllTag,
		storageHandler:   storageHandler,
		newTagCounter:    counterNewTag,
		feedParser:       gofeed.NewParser(),
		notificationChan: notificationChan,
		log:              log,
	}

	return h, nil
}

func (h *Handler) Handle() error {

	err := h.Run(true)
	if err != nil {
		return err
	}
	for range time.Tick(h.pollFrequency) {
		err = h.Run(false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) Run(firstRun bool) error {

	h.log.Infof("Polling all the project URLs...")
	for _, project := range h.projects {
		projectExist := true
		var err error
		if firstRun {
			projectExist, err = h.storageHandler.TagExists(project.ProjectUrl, "")
			if err != nil {
				h.log.Errorf("Got an error while checking if the project %s was already seen, %v", project.ProjectUrl, err)
				return err
			}
			if !projectExist {
				h.log.Infof("New project %s", project.ProjectUrl)
				err = h.storageHandler.StoreTag(project.ProjectUrl, "")
				if err != nil {
					h.log.Errorf("Got an error while marking the project %s was already seen, %v", project.ProjectUrl, err)
					return err
				}
			}
		}
		notify := projectExist || h.notifyAllTags
		err = h.RunSingle(project, notify)
		if err != nil {
			h.log.Errorf("Got an error while handling %s, %v", project.ProjectUrl, err)
			return err
		}
	}
	h.log.Infof("Polling of all the projects completed.")
	return nil
}

func (h *Handler) RunSingle(project model.GithubProject, notify bool) error {

	feedUrl := h.getAtomReleaseFeed(project.ProjectUrl)
	atomFeed, err := h.feedParser.ParseURL(feedUrl)
	if err != nil {
		return err
	}

	for _, item := range atomFeed.Items {
		// parse tag
		tag, repoId, err := parseTag(item.GUID)
		if err != nil {
			return err
		}
		h.log.Debugf("Got a tag %s", tag)

		// filter tag
		tagMatches, err := filterTag(tag, project.TagFilter)
		if err != nil {
			return err
		}
		if !tagMatches {
			continue
		}
		h.log.Debugf("Tag %s matches filter %s", tag, project.TagFilter)

		// filter title
		titleMatches, err := filterTag(item.Title, project.TitleFilter)
		if err != nil {
			return err
		}
		if !titleMatches {
			continue
		}
		h.log.Debugf("Title %s matches filter %s", item.Title, project.TitleFilter)

		// check if it exists
		exists, err := h.storageHandler.TagExists(project.ProjectUrl, tag)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		h.log.Debugf("Tag %s does not exist", tag)

		// if it does not exist already, notify!
		if notify {
			tagToNotify := model.TagToNotify{
				ProjectUrl: project.ProjectUrl,
				Tag:        tag,
				RepoId:     repoId,
				Link:       item.Link,
				Date:       item.Published,
				Author:     item.Author.Name,
			}

			h.log.Debugf("Notify tag %s", tag)
			h.notificationChan <- tagToNotify
		}

		// Mark tag as handled
		h.storageHandler.StoreTag(project.ProjectUrl, tag)

		h.log.Debugf("Done with tag %s", tag)
	}
	return nil
}

func (h *Handler) getAtomReleaseFeed(projectUrl string) string {
	return projectUrl + "/releases.atom"
}
