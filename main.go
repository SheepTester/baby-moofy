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
var delay time.Duration

func init() {
	flag.StringVar(&token, "t", "", "Bot token")
	flag.StringVar(&path, "p", "", "Markov chain frequencies JSON file path")
	flag.IntVar(&order, "o", 3, "Markov chain order")
	flag.DurationVar(&delay, "d", 30 * time.Second, "Default delay per channel between contributions")
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
		DefaultDelay: delay,
	})
}
