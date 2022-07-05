package controllers

import (
	"regexp"
)

var externalPattern = regexp.MustCompile(`[^.]\.kartverket\.no`)

func isExternal(hostname string) bool {
	return externalPattern.MatchString(hostname)
}
