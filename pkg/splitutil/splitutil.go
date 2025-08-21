package splitutil

import (
	"regexp"
)

var trimExpr = regexp.MustCompile(`/([^/]+)`)

func SplitURL(url string, vars []string, skip ...int) (map[string]string, error) {
	trimmedMatches := trimExpr.FindAllStringSubmatch(url, -1)

	startSlice := 0

	if len(skip) > 0 {
		startSlice += skip[0]
	}

	var trimmed []string
	for _, match := range trimmedMatches {
		if len(match) > 1 {
			trimmed = append(trimmed, match[1])
		}
	}

	if startSlice < len(trimmed) {
		trimmed = trimmed[startSlice:]
	} else {
		trimmed = []string{}
	}

	matched := make(map[string]string, len(vars))

	for i, v := range vars {
		if i < len(trimmed) {
			matched[v] = trimmed[i]
		}
	}

	return matched, nil
}
