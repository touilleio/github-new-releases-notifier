package model

type Notification struct {
	Uri string `yaml:"uri"`
}

type TagToNotify struct {
	ProjectUrl string
	Tag        string
	RepoId     string
	Link       string
	Date       string
	Author     string
}
