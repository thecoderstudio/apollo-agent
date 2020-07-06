package net

import "regexp"

// GetHostFromURLString returns the host in a URL string.
func GetHostFromURLString(uncleanedURL string) string {
	re := regexp.MustCompile(`(.*:\/\/)|\/.*`)
	return re.ReplaceAllString(uncleanedURL, "")
}
