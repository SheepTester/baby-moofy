package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func Saver(path string, channel <-chan interface{}) {
	for data := range channel {
		file, err := json.MarshalIndent(data, "", "\t")
		if err == nil {
			// 0644 is just some weird flags; they are not to be spoken of
			err = ioutil.WriteFile(path, file, 0644)
		} else {
			fmt.Println("Problem saving data to JSON file,", path, err)
		}
	}
}
