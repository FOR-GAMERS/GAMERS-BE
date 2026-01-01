package utils

import "regexp"

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	tagRegex      = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

func IsMatchingEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsMatchingUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func IsMatchingTag(tag string) bool {
	return tagRegex.MatchString(tag)
}
