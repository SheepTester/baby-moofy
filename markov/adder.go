package markov

import (
	"time"
	"github.com/SheepTester/baby-moofy/utils"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type MarkovComm struct {
	add       chan<- Markov
	context   chan<- string
	generated <-chan string
}

func (comm *MarkovComm) Close() {
	close(comm.add)
	close(comm.context)
}

func (comm *MarkovComm) Add(miniMarkov Markov) {
	comm.add <- miniMarkov
}

func (comm *MarkovComm) Generate(context string) string {
	comm.context <- context
	return <- comm.generated
}

type markovManagerComm struct {
	add       <-chan Markov
	context   <-chan string
	generated chan<- string
}

type SaveOptions struct {
	Path string
	Delay time.Duration
}

func markovManager(markov Markov, comm *markovManagerComm, saveOpts *SaveOptions) {
	var saveChan chan interface{}
	var timer *time.Timer
	var timerChan <-chan time.Time
	if saveOpts != nil {
		saveChan = make(chan interface{}, 100)
		defer close(saveChan)
		go utils.Saver(saveOpts.Path, saveChan)

		timer = time.NewTimer(saveOpts.Delay)
		timerChan = timer.C
	}

	needSaving := false
	for comm.add != nil || comm.context != nil {
		select {
		case miniMarkov, open := <-comm.add:
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
			if !needSaving {
				// Drain channel just in case
				timer.Stop()
				select {
				case <-timerChan:
				default:
				}
				timer.Reset(saveOpts.Delay)
				needSaving = true
			}
			if !open {
				comm.add = nil
			}
		case context, open := <-comm.context:
			comm.generated <- Generate(markov, context)
			if !open {
				comm.context = nil
			}
		case <- timerChan:
			if needSaving {
				saveChan <- markov.Clone()
				needSaving = false
			}
		}
	}
	// Save one last time
	if saveOpts != nil {
		saveChan <- markov.Clone()
	}
	close(comm.generated)
}

func NewMarkovManager(markov Markov, saveOpts *SaveOptions) *MarkovComm {
	addChan := make(chan Markov, 100)
	contextChan := make(chan string, 100)
	generatedChan := make(chan string, 100)
	go markovManager(markov, &markovManagerComm{addChan, contextChan, generatedChan}, saveOpts)
	return &MarkovComm{addChan, contextChan, generatedChan}
}

func LoadMarkov(path string) (markov Markov, err error) {
	markov = make(Markov)
	data, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(data, &markov)
	}
	return
}

func NewMarkovManagerFromFile(saveOpts *SaveOptions) (*MarkovComm, error) {
	markov, err := LoadMarkov(saveOpts.Path)
	if err != nil {
		fmt.Println("Something went wrong with getting the markov frequencies file.")
		return nil, err
	} else {
		return NewMarkovManager(markov, saveOpts), nil
	}
}
