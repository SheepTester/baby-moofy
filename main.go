// Adapted from https://github.com/bwmarrin/discordgo/blob/master/examples/pingpong/main.go

package main

import (
	"flag"
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
	bot.Start(&bot.BotOptions{
		MarkovPath: path,
		Token: token,
		MarkovOrder: order,
		DefaultDelay: 30 * time.Second, // May want to increase in the future
	})
}
