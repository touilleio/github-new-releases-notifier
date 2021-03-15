package model

type GithubProject struct {
	ProjectUrl string `yaml:"projectUrl"`
	TagFilter  string `yaml:"tagFilter"`
	TitleFilter  string `yaml:"titleFilter"`
}
