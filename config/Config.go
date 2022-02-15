package config

import (
	"encoding/json"
	"hueshelly/logging"
	"io/ioutil"
	"os"
)

var (
	HueShellyConfig Config
)

type Config struct {
	HueBridgeIp               string `json:"hueBridgeIp"`
	HueUser                   string `json:"hueUser"`
	ServerPort                int    `json:"serverPort"`
	RestorePreviousLightState bool   `json:"restorePreviousLightState"`
}

func (config Config) New() {
	logging.Logger.Println("Reading config from config.json")
	jsonFile, err := os.Open("config.json")
	if err != nil {
		handleConfigReadError(err)
	}
	jsonContent, err := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(jsonContent, &HueShellyConfig)
	handleConfigReadError(err)
	if err != nil {
		handleConfigReadError(err)
	}
	err = jsonFile.Close()
	if err != nil {
		handleConfigReadError(err)
	}
}

func handleConfigReadError(err error) {
	if err != nil {
		logging.Logger.Println("Error reading config.json. Please make sure file exists and is valid.")
		logging.Logger.Fatalln(err)
	}
}
