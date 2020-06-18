package markov

import (
	"strings"
	"math/rand"
)

var order int = 3

// Maps a sequence of strings (joined by a space) to a map of next words and frequencies
type Markov = map[string]map[string]int

func Generate(markov Markov, context string) string {
	var builder strings.Builder
	for {
		frequencies, ok := markov[context]
		if !ok {
			break
		}
		total := 0
		for _, frequency := range frequencies {
			total += frequency
		}
		value := rand.Intn(total)
		var next string
		for nextWord, frequency := range frequencies {
			value -= frequency
			if value < 0 {
				next = nextWord
				break
			}
		}
		// A slash means a message break, so the generation is finished
		if next == "/" {
			break
		}
		next = " " + next
		builder.WriteString(next)
		context = strings.SplitN(context, " ", 2)[1] + next
	}
	return builder.String()
}
