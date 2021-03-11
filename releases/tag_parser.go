package releases

import "regexp"

var githubReleaseTagRe = regexp.MustCompile(`^tag:github.com,2008:Repository/(?P<repoId>\d+)/(?P<tag>.+)$`)

// https://golang.org/pkg/regexp/syntax/

func parseTag(id string) (string, string, error) {
	m := githubReleaseTagRe.FindStringSubmatch(id)
	return m[2], m[1], nil
}
