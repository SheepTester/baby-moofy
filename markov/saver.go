package markov

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Thanks https://www.youtube.com/watch?v=LvgVSSpwND8 :D
func MarkovSaver(path string, channel <-chan Markov) {
	for markov := range channel {
		file, err := json.MarshalIndent(markov, "", "\t")
		if err == nil {
			// 0644 is just some weird flags; they are not to be spoken of
			err = ioutil.WriteFile(path, file, 0644)
			fmt.Println("Saved.")
		} else {
			fmt.Println("Problem saving frequencies to JSON file,", err)
		}
	}
}

func LoadMarkov(path string) (markov Markov, err error) {
	markov = make(Markov)
	data, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(data, &markov)
	}
	return
}
