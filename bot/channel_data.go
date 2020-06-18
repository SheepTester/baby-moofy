package bot

type ChannelComm struct {
	request chan<- string
	data     <-chan interface{}
	set    chan<- *setRequest
}

type dataManagerComm struct {
	request <-chan string
	data     chan<- interface{}
	set    <-chan *setRequest
}

type setRequest struct {
	channelID string
	data interface{}
}

func (comm *ChannelComm) Close() {
	close(comm.request)
	close(comm.set)
}

func (comm *ChannelComm) Get(channelID string) interface{} {
	comm.request <- channelID
	return <- comm.data
}

func (comm *ChannelComm) Set(channelID string, data interface{}) {
	comm.set <- &setRequest{channelID, data}
}

func channelDataManager(defaultValue interface{}, comm *dataManagerComm) {
	// Maps channel ID to last words in that channel
	data := make(map[string]interface{})
	for comm.request != nil || comm.set != nil {
		select {
		case channelID, open := <-comm.request:
			channelData, ok := data[channelID]
			if ok {
				comm.data <- channelData
			} else {
				comm.data <- defaultValue
			}
			if !open {
				comm.request = nil
			}
		case setRequest, open := <-comm.set:
			data[setRequest.channelID] = setRequest.data
			if !open {
				comm.set = nil
			}
		}
	}
	close(comm.data)
}

func NewChannelData(defaultValue interface{}) *ChannelComm {
	requestChan := make(chan string, 100)
	dataChan := make(chan interface{}, 100)
	setChan := make(chan *setRequest, 100)
	go channelDataManager(defaultValue, &dataManagerComm{requestChan, dataChan, setChan})
	return &ChannelComm{requestChan, dataChan, setChan}
}
