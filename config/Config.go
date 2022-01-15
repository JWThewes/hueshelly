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

type Config struct ***REMOVED***
	HueBridgeIp string `json:"hueBridgeIp"`
	HueUser     string `json:"hueUser"`
	ServerPort  int    `json:"serverPort"`
***REMOVED***

func (config Config) New() ***REMOVED***
	logging.Logger.Println("Reading config from config.json")
	jsonFile, err := os.Open("config.json")
	if err != nil ***REMOVED***
		handleConfigReadError(err)
	***REMOVED***
	jsonContent, err := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(jsonContent, &HueShellyConfig)
	handleConfigReadError(err)
	if err != nil ***REMOVED***
		handleConfigReadError(err)
	***REMOVED***
	err = jsonFile.Close()
	if err != nil ***REMOVED***
		handleConfigReadError(err)
	***REMOVED***
***REMOVED***

func handleConfigReadError(err error) ***REMOVED***
	if err != nil ***REMOVED***
		logging.Logger.Println("Error reading config.json. Please make sure file exists and is valid.")
		logging.Logger.Fatalln(err)
	***REMOVED***
***REMOVED***
