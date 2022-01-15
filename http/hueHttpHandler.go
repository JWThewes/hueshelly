package hueHttp

import (
	"encoding/json"
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
	writer.Header().Add("Content-Type", "application/json; Charset=UTF-8'")
	jsonValue, _ := json.Marshal(groups)
	groups = nil
	handler.writeResponse(writer, string(jsonValue))
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
