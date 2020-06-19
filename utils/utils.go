package utils

import (
	"regexp"
	"strings"
)

var filterChars = regexp.MustCompile(`[^a-z0-9\s?]`)
var getWord = regexp.MustCompile(`\w+|\?`)

// "Wow 50% ok???sure lol." -> wow 50 ok ? ? ? sure lol
func Simplify(message string) ([]string, bool) {
	// regexp.Split: negative number means ALL the elements yes please thank you
	words := getWord.FindAllString(filterChars.ReplaceAllString(strings.ToLower(message), ""), -1)
	return words, strings.HasSuffix(message, "...")
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

func LastN(words []string, n int) []string {
	if len(words) < n {
		return words
	} else {
		return words[len(words)-n:]
	}
}
