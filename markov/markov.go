package markov

import (
	"math/rand"
	"strings"
	"fmt"
)

var order int = 3

// Maps a sequence of strings (joined by a space) to a map of next words and frequencies
type Markov = map[string]map[string]int

func generate(markov Markov, context string, loud bool) string {
	var builder strings.Builder
	if loud {
		fmt.Println("[begin]")
	}
	for {
		frequencies, ok := markov[context]
		if !ok {
			if loud {
				fmt.Println("[couldn't find frequencies for]", context)
			}
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
				if loud {
					fmt.Printf("[found] %s -> %s (%d/%d chance)\n", context, next, frequency, total)
				}
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
	if loud {
		fmt.Println("[end]")
	}
	return builder.String()
}

func Generate(markov Markov, context string) string {
	return generate(markov, context, false)
}

func GenerateLoud(markov Markov, context string) string {
	return generate(markov, context, true)
}