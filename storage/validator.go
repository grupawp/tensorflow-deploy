package storage

import (
	"regexp"
)

// isDirectoryLayoutValid compares directory struct with regexp
func isDirectoryLayoutValid(directoryLayout, directoryRegexp []string) bool {
	for _, regexpToCheck := range directoryRegexp {
		re := regexp.MustCompile(regexpToCheck)
		match := false
		for _, directoryItem := range directoryLayout {
			if re.MatchString(directoryItem) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}
