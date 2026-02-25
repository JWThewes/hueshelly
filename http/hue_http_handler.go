package huehttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hueshelly/hue"
	"hueshelly/logging"
)

var errNilHueService = errors.New("hue service is nil")

type Handler struct {
	hueService *hue.Service
}

type errorResponse struct {
	Error string `json:"error"`
}

func New(hueService *hue.Service) (*Handler, error) {
	if hueService == nil {
		return nil, errNilHueService
	}
	return &Handler{hueService: hueService}, nil
}

func (handler *Handler) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/toggle/lights/group/", handler.toggleLightsRoom)
	mux.HandleFunc("/toggle/light/", handler.toggleLight)
	mux.HandleFunc("/groups", handler.groups)

	server := http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logging.Logger.Println("Starting server on", addr)
	return server.ListenAndServe()
}

func (handler *Handler) toggleLightsRoom(writer http.ResponseWriter, request *http.Request) {
	if !isToggleMethod(request.Method) {
		handler.writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	room, err := parseRoomName(request.URL.Path)
	if err != nil {
		handler.writeError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := handler.hueService.ToggleLightsInRoom(room); err != nil {
		handler.writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) toggleLight(writer http.ResponseWriter, request *http.Request) {
	if !isToggleMethod(request.Method) {
		handler.writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	lightID, err := parseLightID(request.URL.Path)
	if err != nil {
		handler.writeError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := handler.hueService.ToggleLight(lightID); err != nil {
		handler.writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) groups(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		handler.writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	groups, err := handler.hueService.AvailableGroups()
	if err != nil {
		handler.writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	handler.writeJSON(writer, http.StatusOK, groups)
}

func (handler *Handler) writeError(writer http.ResponseWriter, statusCode int, message string) {
	logging.Logger.Println(message)
	handler.writeJSON(writer, statusCode, errorResponse{Error: message})
}

func (handler *Handler) writeJSON(writer http.ResponseWriter, statusCode int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		logging.Logger.Println(fmt.Errorf("encode JSON response: %w", err))
	}
}

func parseRoomName(path string) (string, error) {
	const prefix = "/toggle/lights/group/"
	if !strings.HasPrefix(path, prefix) {
		return "", errors.New("invalid group path")
	}

	room := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	switch {
	case room == "":
		return "", errors.New("given group name is not valid")
	case len(room) > 32:
		return "", errors.New("given group name is not valid")
	case strings.Contains(room, "/"):
		return "", errors.New("given group name is not valid")
	}
	return room, nil
}

func parseLightID(path string) (int, error) {
	const prefix = "/toggle/light/"
	if !strings.HasPrefix(path, prefix) {
		return 0, errors.New("invalid light path")
	}

	rawID := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if rawID == "" || strings.Contains(rawID, "/") {
		return 0, errors.New("given light id is not valid")
	}

	lightID, err := strconv.Atoi(rawID)
	if err != nil || lightID <= 0 {
		return 0, errors.New("given light id is not valid")
	}
	return lightID, nil
}

func isToggleMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodPost
}
