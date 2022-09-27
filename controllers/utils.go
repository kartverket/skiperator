package controllers

import (
	"regexp"
)

var internalPattern = regexp.MustCompile(`[^.]\.skip\.statkart\.no`)

func isInternal(hostname string) bool {
	return internalPattern.MatchString(hostname)
}
