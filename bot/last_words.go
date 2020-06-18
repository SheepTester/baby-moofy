package bot

type LastWordsComm struct {
	Request chan<- string
	Get     <-chan []string
	Save    chan<- []string
}

func lastWordsTracker(requestChan <-chan string, getChan chan<- []string, saveChan <-chan []string) {
	// Maps channel ID to last words in that channel
	channelLastWords := make(map[string][]string)
	for channelID := range requestChan {
		getChan <- channelLastWords[channelID]
		channelLastWords[channelID] = <-saveChan
	}
	close(getChan)
}

func NewLastWordsTracker() LastWordsComm {
	requestChan := make(chan string, 100)
	getChan := make(chan []string, 100)
	saveChan := make(chan []string, 100)
	go lastWordsTracker(requestChan, getChan, saveChan)
	return LastWordsComm{requestChan, getChan, saveChan}
}

func CloseLastWords(comm LastWordsComm) {
	close(comm.Request)
	close(comm.Save)
}
