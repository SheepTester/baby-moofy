package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"math/rand"
	"time"

	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/SheepTester/baby-moofy/markov"
	"github.com/SheepTester/baby-moofy/utils"
)

var channelLastWords *LastWordsComm
var markovManager *markov.MarkovComm

var order int

type BotOptions struct {
	MarkovPath string
	Token string
	MarkovOrder int
}

func Start(options *BotOptions) {
	rand.Seed(time.Now().UnixNano())

	order = options.MarkovOrder

	var err error

	channelLastWords = NewLastWordsTracker()
	defer channelLastWords.Close()

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

func MessageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore bots
	if msg.Author.Bot {
		return
	}

	words := utils.Simplify(msg.Content)
	// Message cannot be empty
	if len(words) == 0 {
		return
	}

	lastWords := channelLastWords.Get(msg.ChannelID)
	sequence := append(append(lastWords, words...), "/")
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
	var contextWords []string
	if len(sequence) < order {
		contextWords = sequence
	} else {
		contextWords = sequence[len(sequence)-order:]
	}
	channelLastWords.Save(msg.ChannelID, contextWords)
	markovManager.Add(miniMarkov)

	context := strings.Join(contextWords[0:order], " ")
	if gen := markovManager.Generate(context); gen != "" {
		mentionsMe := HasUser(msg.Mentions, session.State.User)
		if mentionsMe {
			session.ChannelMessageSend(msg.ChannelID, gen)
		}
	}
}
