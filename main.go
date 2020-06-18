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
// This maps a sequence of strings (joined by a space) to a map of next words and frequencies
var markov map[string]map[string]int

// Maps channel ID to last words in that channel
var channelLastWords map[string][]string

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

func saveMarkov() error {
	file, err := json.MarshalIndent(markov, "", " ")
	if err == nil {
		// 0644 is just some weird flags; they are not to be spoken of
		err = ioutil.WriteFile("./data/frequencies.json", file, 0644)
	}
	return err
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

	channelLastWords = make(map[string][]string)

	data, err := ioutil.ReadFile("./data/frequencies.json")
  if err == nil {
		err = json.Unmarshal(data, &markov)
  }
	if err != nil {
		markov = make(map[string]map[string]int)
	}

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
	if len(words) > 0 {
		lastWords := channelLastWords[msg.ChannelID]
		sequence := append(append(lastWords, "/"), words...)
		for i, word := range sequence {
			if i >= order {
				context := strings.Join(sequence[i - order : i], " ")
				frequencies, ok := markov[context]
				if !ok {
					frequencies = make(map[string]int)
					markov[context] = frequencies
				}
				frequencies[word]++
			}
		}
		if len(sequence) < order {
			channelLastWords[msg.ChannelID] = sequence
		} else {
			channelLastWords[msg.ChannelID] = sequence[len(sequence) - order :]
		}
		err := saveMarkov()
		if err != nil {
			fmt.Println("Problem saving frequencies:", err)
		}
	}

	// rand.Intn(2) == 0
	// if hasWord(words, "moofy") {
	// 	session.ChannelMessageSend(msg.ChannelID, strings.Join(message, " "))
	// }
}
