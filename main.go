// Adapted from https://github.com/bwmarrin/discordgo/blob/master/examples/pingpong/main.go

package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/SheepTester/baby-moofy/bot"
)

var token string

var path string = "./data/frequencies.json"

var order int = 3

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

	bot.Start(path, token, order)
}
