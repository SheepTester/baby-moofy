package main

import (
	"fmt"
	"bufio"
	"strings"
	"os"

	"github.com/SheepTester/baby-moofy/markov"
	"github.com/SheepTester/baby-moofy/utils"
)

var path string = "./data/frequencies.json"
var order int = 3

func main() {
	fmt.Println("Hello! Let me remember how to speak...")

	markovFreqs, err := markov.LoadMarkov(path)
	if err != nil {
		fmt.Println("Couldn't load frequencies JSON file:", err)
		return
	}

	fmt.Println("Okay, I'm ready.")

	lastWords := []string{"/"}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		words := utils.Simplify(scanner.Text())
		if len(words) == 0 {
			continue
		}
		total := append(append(lastWords, words...), "/")
		if len(total) < order {
			lastWords = total
			fmt.Println("Insufficient context:", strings.Join(lastWords, " "))
		} else {
			lastWords = total[len(total) - order:]
			gen := markov.GenerateLoud(markovFreqs, strings.Join(lastWords, " "))
			fmt.Println(gen)
		}
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Problem reading the epic STIN:", err)
	}
}
