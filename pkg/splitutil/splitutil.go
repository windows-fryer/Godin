package splitutil

import (
	"errors"
	"regexp"
	"strings"
)

var trimExpr = regexp.MustCompile(`/v\d*/(.*)`)

func SplitURL(url string, vars []string, skip ...int) (map[string]string, error) {
	trimmedMatches := trimExpr.FindStringSubmatch(url)

	if len(trimmedMatches) != 2 {
		return nil, errors.New("invalid url")
	}

	startSlice := 0

	if len(skip) > 0 {
		startSlice = skip[0]
	}

	trimmed := strings.Split(trimmedMatches[1], "/")[startSlice:]
	matched := make(map[string]string, len(vars))

	for i, v := range vars {
		matched[v] = trimmed[i]
	}

	return matched, nil
}
