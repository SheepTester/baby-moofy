package utils

import (
	"regexp"
	"strings"
)

var filterChars = regexp.MustCompile(`[^a-z0-9\s?]`)
var getWord = regexp.MustCompile(`\w+|\?`)

// "Wow 50% ok???sure lol." -> wow 50 ok ? ? ? sure lol
func Simplify(message string) []string {
	// regexp.Split: negative number means ALL the elements yes please thank you
	return getWord.FindAllString(filterChars.ReplaceAllString(strings.ToLower(message), ""), -1)
}

// Go doesn't have contains for splices??
// https://stackoverflow.com/a/10485970
func HasWord(words []string, target string) bool {
	for _, word := range words {
		if word == target {
			return true
		}
	}
	return false
}
