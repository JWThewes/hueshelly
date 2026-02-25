package huehttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
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

type roomResponse struct {
	Name string `json:"name"`
}

type lightResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Room string `json:"room"`
}

type homePageData struct {
	GeneratedAt string
	Rooms       []roomResponse
	Lights      []lightResponse
}

var homePageTemplate = template.Must(template.New("home").Funcs(template.FuncMap{
	"pathEscape": url.PathEscape,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>hueshelly</title>
  <style>
    body { font-family: "Trebuchet MS", "Segoe UI", sans-serif; margin: 0; background: #f5f7fa; color: #212b36; }
    .container { max-width: 980px; margin: 0 auto; padding: 24px 18px 36px; }
    h1 { margin: 0 0 8px; }
    .meta { color: #4f5b67; margin-bottom: 16px; }
    .panel { background: #ffffff; border-radius: 10px; padding: 18px; box-shadow: 0 1px 6px rgba(15, 23, 42, 0.08); margin-bottom: 16px; }
    table { width: 100%; border-collapse: collapse; font-size: 14px; }
    th, td { text-align: left; padding: 8px 6px; border-bottom: 1px solid #e6eaef; }
    th { font-size: 13px; text-transform: uppercase; color: #526171; }
    a { color: #0b66d0; text-decoration: none; }
    a:hover { text-decoration: underline; }
    code { background: #f1f4f8; padding: 2px 4px; border-radius: 4px; }
  </style>
</head>
<body>
  <div class="container">
    <h1>hueshelly</h1>
    <div class="meta">Generated at {{.GeneratedAt}}</div>
    <div class="panel">
      <h2>Endpoints</h2>
      <p><a href="/groups">/groups</a> full group and light JSON</p>
      <p><a href="/rooms">/rooms</a> room list JSON</p>
      <p><a href="/lights">/lights</a> light list JSON (flattened)</p>
    </div>
    <div class="panel">
      <h2>Rooms</h2>
      <table>
        <thead><tr><th>Room</th><th>Toggle URL</th></tr></thead>
        <tbody>
        {{range .Rooms}}
          <tr>
            <td>{{.Name}}</td>
            <td><code>/toggle/lights/group/{{pathEscape .Name}}</code></td>
          </tr>
        {{else}}
          <tr><td colspan="2">No rooms found.</td></tr>
        {{end}}
        </tbody>
      </table>
    </div>
    <div class="panel">
      <h2>Lights</h2>
      <table>
        <thead><tr><th>ID</th><th>Name</th><th>Room</th><th>Toggle URL</th></tr></thead>
        <tbody>
        {{range .Lights}}
          <tr>
            <td>{{.ID}}</td>
            <td>{{.Name}}</td>
            <td>{{.Room}}</td>
            <td><code>/toggle/light/{{.ID}}</code></td>
          </tr>
        {{else}}
          <tr><td colspan="4">No lights found.</td></tr>
        {{end}}
        </tbody>
      </table>
    </div>
  </div>
</body>
</html>`))

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
	mux.HandleFunc("/rooms", handler.rooms)
	mux.HandleFunc("/lights", handler.lights)
	mux.HandleFunc("/", handler.home)

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

func (handler *Handler) rooms(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		handler.writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	groups, err := handler.hueService.AvailableGroups()
	if err != nil {
		handler.writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	handler.writeJSON(writer, http.StatusOK, collectRooms(groups))
}

func (handler *Handler) lights(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		handler.writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	groups, err := handler.hueService.AvailableGroups()
	if err != nil {
		handler.writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	handler.writeJSON(writer, http.StatusOK, collectLights(groups))
}

func (handler *Handler) home(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		handler.writeError(writer, http.StatusNotFound, "endpoint not found")
		return
	}
	if request.Method != http.MethodGet {
		handler.writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	groups, err := handler.hueService.AvailableGroups()
	if err != nil {
		handler.writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	pageData := homePageData{
		GeneratedAt: time.Now().Format(time.RFC1123),
		Rooms:       collectRooms(groups),
		Lights:      collectLights(groups),
	}

	var page bytes.Buffer
	if err := homePageTemplate.Execute(&page, pageData); err != nil {
		handler.writeError(writer, http.StatusInternalServerError, "failed to render home page")
		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(page.Bytes()); err != nil {
		logging.Logger.Println(fmt.Errorf("write home page: %w", err))
	}
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

func collectRooms(groups []hue.Group) []roomResponse {
	rooms := make([]roomResponse, 0, len(groups))
	for _, group := range groups {
		rooms = append(rooms, roomResponse{Name: group.Name})
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].Name < rooms[j].Name
	})
	return rooms
}

func collectLights(groups []hue.Group) []lightResponse {
	lights := make([]lightResponse, 0)
	for _, group := range groups {
		for _, light := range group.Lights {
			lights = append(lights, lightResponse{
				ID:   light.ID,
				Name: light.Name,
				Room: group.Name,
			})
		}
	}

	sort.Slice(lights, func(i, j int) bool {
		if lights[i].Room != lights[j].Room {
			return lights[i].Room < lights[j].Room
		}
		if lights[i].Name != lights[j].Name {
			return lights[i].Name < lights[j].Name
		}
		return lights[i].ID < lights[j].ID
	})
	return lights
}
