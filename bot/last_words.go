package bot

type LastWordsComm struct {
	request chan<- string
	get     <-chan []string
	save    chan<- *saveLastWords
}

type saveLastWords struct {
	channelID string
	words []string
}

func (comm *LastWordsComm) Close() {
	close(comm.request)
	close(comm.save)
}

func (comm *LastWordsComm) Get(channelID string) []string {
	comm.request <- channelID
	return <- comm.get
}

func (comm *LastWordsComm) Save(channelID string, words []string) {
	comm.save <- &saveLastWords{channelID, words}
}

func lastWordsTracker(requestChan <-chan string, getChan chan<- []string, saveChan <-chan *saveLastWords) {
	// Maps channel ID to last words in that channel
	channelLastWords := make(map[string][]string)
	for requestChan != nil || saveChan != nil {
		select {
		case channelID, open := <-requestChan:
			lastWords, ok := channelLastWords[channelID]
			if ok {
				getChan <- lastWords
			} else {
				getChan <- []string{"/"}
			}
			if !open {
				requestChan = nil
			}
		case saveRequest, open := <-saveChan:
			channelLastWords[saveRequest.channelID] = saveRequest.words
			if !open {
				saveChan = nil
			}
		}
	}
	close(getChan)
}

func NewLastWordsTracker() *LastWordsComm {
	requestChan := make(chan string, 100)
	getChan := make(chan []string, 100)
	saveChan := make(chan *saveLastWords, 100)
	go lastWordsTracker(requestChan, getChan, saveChan)
	return &LastWordsComm{requestChan, getChan, saveChan}
}
