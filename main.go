package main

import (
	"hueshelly/config"
	hueHttp "hueshelly/http"
	"hueshelly/logging"
)

func init() ***REMOVED***
	logging.Logger.Println("Starting")
***REMOVED***

func main() ***REMOVED***
	config.HueShellyConfig.New()
	hueHttp.New()
***REMOVED***
