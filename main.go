package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"hueshelly/config"
	huehttp "hueshelly/http"
	"hueshelly/hue"
	"hueshelly/logging"
)

func main() {
	if err := logging.Init(""); err != nil {
		log.Printf("failed to initialize file logging: %v", err)
	}
	defer func() {
		if err := logging.Close(); err != nil {
			logging.Logger.Println(err)
		}
	}()

	cfg, err := config.Load("config.json")
	if err != nil {
		logging.Logger.Fatal(err)
	}

	hueService, err := hue.New(cfg)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	handler, err := huehttp.New(hueService)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	address := fmt.Sprintf(":%d", cfg.ServerPort)
	if err := handler.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Logger.Fatal(err)
	}
}
