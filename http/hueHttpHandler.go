package hueHttp

import (
	"encoding/json"
	"errors"
	"hueshelly/config"
	"hueshelly/hue"
	"hueshelly/logging"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	hueService hue.Service
}

func New() Handler {
	httpHandler := Handler{hueService: hue.New()}
	httpHandler.init()
	return httpHandler
}

func (handler Handler) init() {
	http.HandleFunc("/toggle/lights/group/", handler.toggleLightsRoom)
	http.HandleFunc("/toggle/light/", handler.toggleLight)
	http.HandleFunc("/groups", handler.groups)
	serverPort := strconv.Itoa(config.HueShellyConfig.ServerPort)
	logging.Logger.Println("Starting server on port " + serverPort)
	err := http.ListenAndServe(":"+serverPort, nil)
	if err != nil {
		logging.Logger.Fatal(err)
	}
}

func (handler Handler) toggleLightsRoom(writer http.ResponseWriter, request *http.Request) {
	room := strings.TrimPrefix(request.URL.Path, "/toggle/lights/group/")
	room = strings.Replace(room, "\n", "", -1)
	room = strings.Replace(room, "\r", "", -1)
	if len(room) == 0 || len(room) > 32 {
		handler.handleError(writer, errors.New("given group name is not valid"))
		return
	}
	err := handler.hueService.ToggleLightsInRoom(room)
	if err != nil {
		handler.handleError(writer, err)
	}
}

func (handler *Handler) toggleLight(writer http.ResponseWriter, request *http.Request) {
	lightId, _ := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/toggle/light/"))
	err := handler.hueService.ToggleLight(lightId)
	if err != nil {
		handler.handleError(writer, err)
	}
}

func (handler Handler) groups(writer http.ResponseWriter, _ *http.Request) {
	groups := handler.hueService.AvailableGroups()
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(groups); err != nil {
		log.Println(err)
	}
}

func (handler *Handler) handleError(writer http.ResponseWriter, err error) {
	logging.Logger.Println(err)
	writer.WriteHeader(500)
}

func (handler *Handler) writeResponse(writer http.ResponseWriter, response string) {
	_, err := writer.Write([]byte(response))
	if err != nil {
		log.Println(err)
	}
}
