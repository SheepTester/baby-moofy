package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"math/rand"
	"time"
	"strings"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/SheepTester/baby-moofy/markov"
	"github.com/SheepTester/baby-moofy/utils"
)

var channelLastWords *ChannelComm
var nextContribution *ChannelComm
var channelCooldowns *ChannelComm
var markovManager *markov.MarkovComm

var order int
var defaultDelay time.Duration
var prefix string

type BotOptions struct {
	MarkovPath string
	Token string
	MarkovOrder int
	DefaultDelay time.Duration
	Prefix string
}

func Start(options *BotOptions) {
	rand.Seed(time.Now().UnixNano())

	order = options.MarkovOrder
	defaultDelay = options.DefaultDelay
	prefix = options.Prefix

	var err error

	channelLastWords = NewChannelData([]string{"/"})
	defer channelLastWords.Close()

	nextContribution = NewChannelData(time.Now())
	defer nextContribution.Close()

	channelCooldowns = NewChannelData(defaultDelay)
	defer channelCooldowns.Close()

	markovManager, err = markov.NewMarkovManagerFromFile(&markov.SaveOptions{
		Path: options.MarkovPath,
		Delay: 10 * time.Second, // Save every 10 seconds
	})
	if err != nil {
		fmt.Println("Problem loading frequencies JSON file:", err)
		return
	}
	defer markovManager.Close()

	session, err := discordgo.New("Bot " + options.Token)
	if err != nil {
		fmt.Println("Problem creating Discord session,", err)
		return
	}

	session.AddHandler(MessageCreate)

	err = session.Open()
	if err != nil {
		fmt.Println("Problem connecting,", err)
		return
	}
	defer session.Close()

	fmt.Println("The bot RUNS. Press ctrl + C to terminate.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func trackWords(channelID string, words []string, merge bool) ([]string, []string) {
	data := channelLastWords.GetWillSet(channelID)
	lastWords, ok := data.([]string)
	if !ok {
		fmt.Println("Didn't get a string splice??", data)
		return nil, nil
	}

	sequence := append(lastWords, words...)
	if !merge {
		sequence = append(sequence, "/")
	}
	contextWords := utils.LastN(sequence, order)
	channelLastWords.Set(channelID, contextWords)

	return sequence, contextWords
}

func respond(session *discordgo.Session, msg *discordgo.MessageCreate, text string) {
	session.ChannelMessageSend(msg.ChannelID, text)
	data := channelCooldowns.Get(msg.ChannelID)
	delay, ok := data.(time.Duration)
	if !ok {
		fmt.Println("Didn't get a Duration??", data)
		delay = defaultDelay
	}
	nextContribution.Set(msg.ChannelID, time.Now().Add(delay))

	words, _ := utils.Simplify(text)
	if len(words) > 0 {
		// Update last words sent in channel
		trackWords(msg.ChannelID, words, false)
	}
}

var channelCooldownSetterParser = regexp.MustCompile(`set channel cooldown to (\d+)s`)

func considerCommand(session *discordgo.Session, msg *discordgo.MessageCreate, command string) bool {
	if command == "help" {
		respond(session, msg, `i have and so on up to also is the prefix btw
then after prefix maybe put
set channel cooldown to 3s
for example and i will wait 3 seconds before yelling`)
		return true
	}
	var match []string
	match = channelCooldownSetterParser.FindStringSubmatch(command)
	if match != nil {
		seconds, err := strconv.Atoi(match[1])
		if err != nil {
			fmt.Println("Converting to num is oof", err)
			return false
		}
		channelCooldowns.Set(msg.ChannelID, time.Duration(seconds) * time.Second)
		respond(session, msg, "i will adjust my conceptualization speed accordingly")
		return true
	}
	return false
}

func MessageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore self
	if msg.Author.ID == session.State.User.ID {
		return
	}

	if strings.HasPrefix(msg.Content, prefix) {
		trimmed := strings.TrimSpace(strings.TrimPrefix(msg.Content, prefix))
		if considerCommand(session, msg, trimmed) {
			return
		} else {
			session.MessageReactionAdd(msg.ChannelID, msg.ID, "❓")
		}
	}

	words, trailing := utils.Simplify(msg.Content)
	// Message cannot be empty
	if len(words) == 0 {
		return
	}

	// Stores last words sent in channel
	sequence, contextWords := trackWords(msg.ChannelID, words, trailing)

	// Do not learn from bots
	if !msg.Author.Bot {
		miniMarkov := make(markov.Markov)
		for i, word := range sequence {
			if i >= order {
				context := strings.Join(sequence[i-order:i], " ")
				frequencies, ok := miniMarkov[context]
				if !ok {
					frequencies = make(map[string]int)
					miniMarkov[context] = frequencies
				}
				frequencies[word]++
			}
		}
		markovManager.Add(miniMarkov)
	}

	context := strings.Join(contextWords, " ")
	if gen := markovManager.Generate(context); gen != "" {
		if trailing {
			gen = "..." + gen
		}
		mentionsMe := HasUser(msg.Mentions, session.State.User)
		if !mentionsMe {
			data := nextContribution.GetWillSet(msg.ChannelID)
			nextTime, ok := data.(time.Time)
			if !ok {
				fmt.Println("Didn't get a time??", data)
			}
			// Neither mentioned nor has there been ample time since the previous
			// contribution, so do not speak
			if time.Now().Before(nextTime) {
				// We promised a set earlier though, so we must fulfill it to unlock the
				// channel data manager
				nextContribution.Set(msg.ChannelID, nextTime)
				return
			}
		}
		respond(session, msg, gen)
	}
}
