package bot

type ChannelComm struct {
	get chan<- *getRequest
	data     <-chan interface{}
	set    chan<- *setRequest
}

type dataManagerComm struct {
	get <-chan *getRequest
	data     chan<- interface{}
	set    <-chan *setRequest
}

type getRequest struct {
	channelID string
	willSet bool
}

type setRequest struct {
	channelID string
	data interface{}
}

func (comm *ChannelComm) Close() {
	close(comm.get)
	close(comm.set)
}

func (comm *ChannelComm) Get(channelID string) interface{} {
	comm.get <- &getRequest{channelID, false}
	return <- comm.data
}

func (comm *ChannelComm) GetWillSet(channelID string) interface{} {
	comm.get <- &getRequest{channelID, true}
	return <- comm.data
}

func (comm *ChannelComm) Set(channelID string, data interface{}) {
	comm.set <- &setRequest{channelID, data}
}

func channelDataManager(defaultValue interface{}, comm *dataManagerComm) {
	// Maps channel ID to last words in that channel
	data := make(map[string]interface{})
	doSetRequest := func (request *setRequest, open bool) {
		data[request.channelID] = request.data
		if !open {
			comm.set = nil
		}
	}
	for comm.get != nil || comm.set != nil {
		select {
		case request, open := <-comm.get:
			channelData, ok := data[request.channelID]
			if ok {
				comm.data <- channelData
			} else {
				comm.data <- defaultValue
			}
			if !open {
				comm.get = nil
			}
			// Avoid race conditions by waiting for a set before continuing
			if request.willSet {
				request, open := <-comm.set
				doSetRequest(request, open)
			}
		case request, open := <-comm.set:
			doSetRequest(request, open)
		}
	}
	close(comm.data)
}

func NewChannelData(defaultValue interface{}) *ChannelComm {
	getChan := make(chan *getRequest, 100)
	dataChan := make(chan interface{}, 100)
	setChan := make(chan *setRequest, 100)
	go channelDataManager(defaultValue, &dataManagerComm{getChan, dataChan, setChan})
	return &ChannelComm{getChan, dataChan, setChan}
}
