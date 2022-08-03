package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ParseJsonConfig(configPath string, v interface{}) interface{} {

	jsonConfig, err := os.Open(configPath)

	if err != nil {
		panic(err)
	}

	defer jsonConfig.Close()

	data, _ := ioutil.ReadAll(jsonConfig)

	err = json.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
	return v
}
