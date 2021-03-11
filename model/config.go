package model

import "time"

type ReleaseNotifierConfig struct {
	Projects      []GithubProject `yaml:"projects"`
	PollFrequency time.Duration   `yaml:"pollFrequency"`
	Notification  Notification    `yaml:"notification"`
}
