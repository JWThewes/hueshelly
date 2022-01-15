package hueHttp

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
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
	validate   *validator.Validate
}

func New() Handler {
	httpHandler := Handler{hueService: hue.New()}
	httpHandler.init()
	return httpHandler
}

func (handler Handler) init() {
	handler.initValidator()
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

func (handler *Handler) initValidator() {
	handler.validate = validator.New()
	err := handler.validate.RegisterValidation("room", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) > 0 && fl.Field().Len() <= 32
	})
	if err != nil {
		logging.Logger.Println("Error registering room validator")
	}
}

func (handler Handler) toggleLightsRoom(writer http.ResponseWriter, request *http.Request) {
	room := strings.TrimPrefix(request.URL.Path, "/toggle/lights/group/")
	room = strings.Replace(room, "\n", "", -1)
	room = strings.Replace(room, "\r", "", -1)
	err := handler.validate.Var(room, "room")
	if err != nil {
		handler.handleError(writer, errors.New("given group name is not valid"))
		return
	}
	err = handler.hueService.ToggleLightsInRoom(room)
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
