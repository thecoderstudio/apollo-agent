package net

import "regexp"

const notPartOfHostRegexp = `(.*:\/\/)|(\/|\?).*`

// GetHostFromURLString returns the host in a URL string.
func GetHostFromURLString(uncleanedURL string) string {
	re := regexp.MustCompile(notPartOfHostRegexp)
	return re.ReplaceAllString(uncleanedURL, "")
}
