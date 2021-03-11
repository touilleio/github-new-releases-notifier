package releases

import "regexp"

// https://golang.org/pkg/regexp/syntax/

func filterTag(tag string, tagFilter string) (bool, error) {

	if tagFilter == "" {
		return true, nil
	}

	r, err := regexp.Compile(tagFilter)
	if err != nil {
		return false, err
	}

	matches := r.MatchString(tag)
	return matches, nil
}
