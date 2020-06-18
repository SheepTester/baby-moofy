// Adapted from https://github.com/bwmarrin/discordgo/blob/master/examples/pingpong/main.go

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"strings"
	"regexp"
	"math/rand"
	"time"
	"io/ioutil"
	"encoding/json"

	"github.com/bwmarrin/discordgo"
)

var token string

var order int = 3
var markovAdderChannel chan<- map[string]map[string]int

func markovAdder(markov map[string]map[string]int, channel <-chan map[string]map[string]int, saveChannel chan<- map[string]map[string]int) {
	for miniMarkov := range channel {
		// Merge miniMarkov into big markov
		for context, newFrequencies := range miniMarkov {
			frequencies, ok := markov[context]
			if !ok {
				frequencies = make(map[string]int)
				markov[context] = frequencies
			}
			for possibility, frequency := range newFrequencies {
				frequencies[possibility] += frequency
			}
		}
		saveChannel <- markov
	}
}

var requestLastWords chan<- string
var getLastWords <-chan []string
var saveLastWords chan<- []string

func lastWordsTracker(requestChan <-chan string, getChan chan<- []string, saveChan <-chan []string)  {
	// Maps channel ID to last words in that channel
	channelLastWords := make(map[string][]string)
	for channelID := range requestChan {
		getChan <- channelLastWords[channelID]
		channelLastWords[channelID] = <- saveChan
	}
}

var filterChars = regexp.MustCompile(`[^a-z0-9\s?]`)
var getWord = regexp.MustCompile(`\w+|\?`)

// "Wow 50% ok???sure lol." -> wow 50 ok ? ? ? sure lol
func simplify(message string) []string {
	// regexp.Split: negative number means ALL the elements yes please thank you
	return getWord.FindAllString(filterChars.ReplaceAllString(strings.ToLower(message), ""), -1)
}

// Go doesn't have contains for splices??
// https://stackoverflow.com/a/10485970
func hasWord(words []string, target string) bool {
	for _, word := range words {
		if word == target {
			return true
		}
	}
	return false
}

var path string = "./data/frequencies.json"

// Thanks https://www.youtube.com/watch?v=LvgVSSpwND8 :D
func markovSaver(channel <-chan map[string]map[string]int) {
	for markov := range channel {
		file, err := json.MarshalIndent(markov, "", "\t")
		if err == nil {
			// 0644 is just some weird flags; they are not to be spoken of
			err = ioutil.WriteFile(path, file, 0644)
		}
		if err != nil {
			fmt.Println("Problem saving frequencies to JSON file,", err)
		}
	}
}

func init() {
	flag.StringVar(&token, "t", "", "Bot token")
	flag.Parse()

	if token == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// This maps a sequence of strings (joined by a space) to a map of next words and frequencies
	markov := make(map[string]map[string]int)
	data, err := ioutil.ReadFile(path)
  if err == nil {
		err = json.Unmarshal(data, &markov)
		if err != nil {
			fmt.Println("Problem parsing frequencies JSON,", err)
			return
		}
  }

	requestLastWordsChan := make(chan string, 100)
	defer close(requestLastWordsChan)
	getLastWordsChan := make(chan []string, 100)
	defer close(getLastWordsChan)
	saveLastWordsChan := make(chan []string, 100)
	defer close(saveLastWordsChan)
	go lastWordsTracker(requestLastWordsChan, getLastWordsChan, saveLastWordsChan)
	requestLastWords = requestLastWordsChan
	getLastWords = getLastWordsChan
	saveLastWords = saveLastWordsChan

	saveChannel := make(chan map[string]map[string]int, 100)
	defer close(saveChannel)
	go markovSaver(saveChannel)

	markovAdderBiChannel := make(chan map[string]map[string]int, 100)
	defer close(markovAdderBiChannel)
	go markovAdder(markov, markovAdderBiChannel, saveChannel)
	markovAdderChannel = markovAdderBiChannel

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Problem creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Problem connecting,", err)
		return
	}
	defer dg.Close()

	fmt.Println("The bot RUNS. Press ctrl + C to terminate.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore bots
	if msg.Author.Bot {
		return
	}

	words := simplify(msg.Content)
	// Message cannot be empty
	if len(words) == 0 {
		return
	}

	lastWords := <- getLastWords
	sequence := append(append(lastWords, "/"), words...)
	miniMarkov := make(map[string]map[string]int)
	for i, word := range sequence {
		if i >= order {
			context := strings.Join(sequence[i - order : i], " ")
			frequencies, ok := miniMarkov[context]
			if !ok {
				frequencies = make(map[string]int)
				miniMarkov[context] = frequencies
			}
			frequencies[word]++
		}
	}
	if len(sequence) < order {
		saveLastWords <- sequence
	} else {
		saveLastWords <- sequence[len(sequence) - order :]
	}
	markovAdderChannel <- miniMarkov

	// if rand.Intn(2) == 0 {
	// 	session.ChannelMessageSend(msg.ChannelID, generate())
	// }
}
