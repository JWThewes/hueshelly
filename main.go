package main

import (
	"errors"
	"fmt"
	"net/http"

	"hueshelly/config"
	huehttp "hueshelly/http"
	"hueshelly/hue"
	"hueshelly/logging"
)

func main() {
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
