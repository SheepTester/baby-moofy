package markov

type MarkovComm struct {
	Add chan<- Markov
	Context chan<- string
	Generated <-chan string
}

func markovManager(markov Markov, addChan <-chan Markov, contextChan <-chan string, generatedChan chan<- string, savePath string) {
	var saveChan chan Markov
	if savePath != "" {
		saveChan = make(chan Markov, 100)
		defer close(saveChan)
		go MarkovSaver(savePath, saveChan)
	}

	for addChan != nil || contextChan != nil {
		select {
		case miniMarkov, open := <- addChan:
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
			if saveChan != nil {
				saveChan <- markov
			}
			if !open {
				addChan = nil
			}
		case context, open := <- contextChan:
			generatedChan <- Generate(markov, context)
			if !open {
				contextChan = nil
			}
		}
	}
	close(generatedChan)
}

func NewMarkovManager(markov Markov, savePath string) MarkovComm {
	addChan := make(chan Markov, 100)
	contextChan := make(chan string, 100)
	generatedChan := make(chan string, 100)
	go markovManager(markov, addChan, contextChan, generatedChan, savePath)
	return MarkovComm{addChan, contextChan, generatedChan}
}

func NewMarkovManagerFromFile(path string) (comm MarkovComm, err error) {
	markov, err := LoadMarkov(path)
	if err == nil {
		comm = NewMarkovManager(markov, path)
	}
	return
}

func CloseMarkovManager(comm MarkovComm)  {
	close(comm.Add)
	close(comm.Context)
}
