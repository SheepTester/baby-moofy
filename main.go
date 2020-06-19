// Adapted from https://github.com/bwmarrin/discordgo/blob/master/examples/pingpong/main.go

package main

import (
	"flag"
	"os"
	"time"

	"github.com/SheepTester/baby-moofy/bot"
)

var token string
var path string
var order int
var delay int64

func init() {
	flag.StringVar(&token, "t", "", "Bot token")
	flag.StringVar(&path, "p", "", "Markov chain frequencies JSON file path")
	flag.IntVar(&order, "o", 3, "Markov chain order")
	flag.Int64Var(&delay, "d", 30, "Default delay in seconds per channel between contributions")
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
		DefaultDelay: time.Duration(delay) * time.Second, // May want to increase in the future
	})
}
