package main

import (
	"hueshelly/config"
	hueHttp "hueshelly/http"
	"hueshelly/logging"
)

func init() {
	logging.Logger.Println("Starting")
}

func main() {
	config.HueShellyConfig.New()
	hueHttp.New()
}
