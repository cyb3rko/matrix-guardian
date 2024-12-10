package validation

import (
	"net/url"
	"regexp"
)

const usernameRegex = "^[0-9a-z-.=_/+]+$"

func IsValidUrl(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme == "https" || u.Scheme == "http"
}

func IsValidUsername(username string) bool {
	return regexp.MustCompile(usernameRegex).MatchString(username)
}
