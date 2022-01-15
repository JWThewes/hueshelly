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

type Handler struct ***REMOVED***
	hueService hue.Service
***REMOVED***

func New() Handler ***REMOVED***
	httpHandler := Handler***REMOVED***hueService: hue.New()***REMOVED***
	httpHandler.init()
	return httpHandler
***REMOVED***

func (handler Handler) init() ***REMOVED***
	http.HandleFunc("/toggle/lights/group/", handler.toggleLightsRoom)
	http.HandleFunc("/toggle/light/", handler.toggleLight)
	http.HandleFunc("/groups", handler.groups)
	serverPort := strconv.Itoa(config.HueShellyConfig.ServerPort)
	logging.Logger.Println("Starting server on port " + serverPort)
	err := http.ListenAndServe(":"+serverPort, nil)
	if err != nil ***REMOVED***
		logging.Logger.Fatal(err)
	***REMOVED***
***REMOVED***

func (handler Handler) toggleLightsRoom(writer http.ResponseWriter, request *http.Request) ***REMOVED***
	room := strings.TrimPrefix(request.URL.Path, "/toggle/lights/group/")
	err := handler.hueService.ToggleLightsInRoom(room)
	if err != nil ***REMOVED***
		handler.handleError(writer, err)
	***REMOVED***
***REMOVED***

func (handler *Handler) toggleLight(writer http.ResponseWriter, request *http.Request) ***REMOVED***
	lightId, _ := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/toggle/light/"))
	err := handler.hueService.ToggleLight(lightId)
	if err != nil ***REMOVED***
		handler.handleError(writer, err)
	***REMOVED***
***REMOVED***

func (handler Handler) groups(writer http.ResponseWriter, _ *http.Request) ***REMOVED***
	groups := handler.hueService.AvailableGroups()
	writer.Header().Add("Content-Type", "application/json; Charset=UTF-8'")
	jsonValue, _ := json.Marshal(groups)
	groups = nil
	handler.writeResponse(writer, string(jsonValue))
***REMOVED***

func (handler *Handler) handleError(writer http.ResponseWriter, err error) ***REMOVED***
	logging.Logger.Println(err)
	writer.WriteHeader(500)
***REMOVED***

func (handler *Handler) writeResponse(writer http.ResponseWriter, response string) ***REMOVED***
	_, err := writer.Write([]byte(response))
	if err != nil ***REMOVED***
		log.Println(err)
	***REMOVED***
***REMOVED***
