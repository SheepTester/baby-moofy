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

	"github.com/bwmarrin/discordgo"
)

var token string

var order int = 3
// Cannot do [order]string https://stackoverflow.com/q/38362631
var markov map[[3]string]string

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

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bots
	if m.Author.Bot {
		return
	}

	// Learn
	message := simplify(m.Content)

	// rand.Intn(2) == 0
	if hasWord(message, "moofy") {
		s.ChannelMessageSend(m.ChannelID, strings.Join(message, " "))
	}
}
