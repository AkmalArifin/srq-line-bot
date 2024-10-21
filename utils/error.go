package utils

import "regexp"

func IsDuplicateError(message string) bool {
	pattern := `.*1062.*`
	re := regexp.MustCompile(pattern)

	return re.MatchString(message)
}
