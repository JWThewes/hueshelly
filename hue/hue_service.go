package hue

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"hueshelly/config"
	"hueshelly/logging"

	"github.com/openhue/openhue-go"
)

var errServiceNotInitialized = errors.New("hue service is not initialized")

type Service struct {
	home                      *openhue.Home
	restorePreviousLightState bool
}

func New(cfg config.Config) (*Service, error) {
	bridgeIP := strings.TrimSpace(cfg.HueBridgeIP)
	if bridgeIP != "" {
		logging.Logger.Println("Using bridge at", bridgeIP)
	} else {
		logging.Logger.Println("Searching for bridge")
		discoveredBridge, err := openhue.NewBridgeDiscovery().Discover()
		if err != nil {
			return nil, fmt.Errorf("discover bridge: %w", err)
		}
		bridgeIP = discoveredBridge.IpAddress
		if bridgeIP == "" {
			return nil, fmt.Errorf("discovered bridge but received empty IP address")
		}
		logging.Logger.Println("Found hue bridge at", bridgeIP)
	}

	home, err := openhue.NewHome(bridgeIP, cfg.HueUser)
	if err != nil {
		return nil, fmt.Errorf("create hue home: %w", err)
	}
	if _, err := home.GetBridgeHome(); err != nil {
		return nil, fmt.Errorf("communicate with bridge: %w", err)
	}

	logging.Logger.Println("Logged in at hue bridge")
	return &Service{
		home:                      home,
		restorePreviousLightState: cfg.RestorePreviousLightState,
	}, nil
}

func (service *Service) ToggleLight(lightID int) error {
	if err := service.ensureInitialized(); err != nil {
		return err
	}

	light, err := service.findLightByID(lightID)
	if err != nil {
		return err
	}
	if light.Id == nil {
		return errors.New("light has no id")
	}

	if light.IsOn() {
		off := false
		if err := service.home.UpdateLight(*light.Id, openhue.LightPut{On: &openhue.On{On: &off}}); err != nil {
			return err
		}
		logging.Logger.Println("Light found - toggled to off")
		return nil
	}

	on := true
	body := openhue.LightPut{On: &openhue.On{On: &on}}
	if !service.restorePreviousLightState {
		brightness := openhue.Brightness(100)
		body.Dimming = &openhue.Dimming{Brightness: &brightness}
	}
	if err := service.home.UpdateLight(*light.Id, body); err != nil {
		return err
	}
	logging.Logger.Println("Light found - toggled to on")
	return nil
}

func (service *Service) ToggleLightsInRoom(roomName string) error {
	if err := service.ensureInitialized(); err != nil {
		return err
	}

	rooms, err := service.home.GetRooms()
	if err != nil {
		return fmt.Errorf("get rooms: %w", err)
	}

	for _, room := range rooms {
		if nameFromRoom(room) != roomName {
			continue
		}

		groupedLightID, ok := groupedLightIDFromRoom(room)
		if !ok {
			return errors.New("group has no grouped_light service")
		}
		return service.toggleGroupedLightByID(groupedLightID)
	}

	return fmt.Errorf("no room with name %q found", roomName)
}

func (service *Service) AvailableGroups() ([]Group, error) {
	if err := service.ensureInitialized(); err != nil {
		return nil, err
	}

	rooms, err := service.home.GetRooms()
	if err != nil {
		return nil, fmt.Errorf("get rooms: %w", err)
	}
	lights, err := service.home.GetLights()
	if err != nil {
		return nil, fmt.Errorf("get lights: %w", err)
	}

	groupList := make([]Group, 0, len(rooms))
	for _, room := range rooms {
		group := Group{Name: nameFromRoom(room)}
		lightIDs := service.lightIDsFromRoom(room)
		lightList := make([]Light, 0, len(lightIDs))
		for _, lightID := range lightIDs {
			light, exists := lights[lightID]
			if !exists {
				continue
			}
			lightIDInt, err := lightIDV1ToInt(light.IdV1)
			if err != nil {
				continue
			}
			lightList = append(lightList, Light{
				Name: nameFromLight(light),
				ID:   lightIDInt,
			})
		}

		sort.Slice(lightList, func(i, j int) bool {
			if lightList[i].Name == lightList[j].Name {
				return lightList[i].ID < lightList[j].ID
			}
			return lightList[i].Name < lightList[j].Name
		})
		group.Lights = lightList
		groupList = append(groupList, group)
	}

	sort.Slice(groupList, func(i, j int) bool {
		return groupList[i].Name < groupList[j].Name
	})
	return groupList, nil
}

func (service *Service) toggleGroupedLightByID(groupedLightID string) error {
	groupedLight, err := service.home.GetGroupedLightById(groupedLightID)
	if err != nil {
		return err
	}
	if groupedLight.Id == nil {
		return errors.New("grouped light has no id")
	}

	if groupedLight.IsOn() {
		off := false
		if err := service.home.UpdateGroupedLight(*groupedLight.Id, openhue.GroupedLightPut{On: &openhue.On{On: &off}}); err != nil {
			return err
		}
		logging.Logger.Println("Group found - any lights on toggling to off")
		return nil
	}

	on := true
	body := openhue.GroupedLightPut{On: &openhue.On{On: &on}}
	if !service.restorePreviousLightState {
		brightness := openhue.Brightness(100)
		body.Dimming = &openhue.Dimming{Brightness: &brightness}
	}
	if err := service.home.UpdateGroupedLight(*groupedLight.Id, body); err != nil {
		return err
	}
	logging.Logger.Println("Group found - all lights off toggling to on")
	return nil
}

func (service *Service) findLightByID(lightID int) (*openhue.LightGet, error) {
	lights, err := service.home.GetLights()
	if err != nil {
		return nil, err
	}
	for _, light := range lights {
		lightIDInt, err := lightIDV1ToInt(light.IdV1)
		if err != nil {
			continue
		}
		if lightIDInt == lightID {
			lightCopy := light
			return &lightCopy, nil
		}
	}
	return nil, fmt.Errorf("light with id %d not found", lightID)
}

func lightIDV1ToInt(idV1 *string) (int, error) {
	if idV1 == nil {
		return 0, errors.New("light has no id_v1")
	}
	lightID := strings.TrimPrefix(*idV1, "/lights/")
	if lightID == *idV1 {
		return 0, errors.New("light id_v1 is not a v1 light path")
	}
	return strconv.Atoi(lightID)
}

func groupedLightIDFromRoom(room openhue.RoomGet) (string, bool) {
	if room.Services == nil {
		return "", false
	}
	for _, roomService := range *room.Services {
		if roomService.Rid == nil || roomService.Rtype == nil {
			continue
		}
		if *roomService.Rtype == openhue.ResourceIdentifierRtypeGroupedLight {
			return *roomService.Rid, true
		}
	}
	return "", false
}

func nameFromRoom(room openhue.RoomGet) string {
	if room.Metadata == nil || room.Metadata.Name == nil {
		return ""
	}
	return *room.Metadata.Name
}

func nameFromLight(light openhue.LightGet) string {
	if light.Metadata == nil || light.Metadata.Name == nil {
		return ""
	}
	return *light.Metadata.Name
}

func (service *Service) lightIDsFromRoom(room openhue.RoomGet) []string {
	lightMap := map[string]struct{}{}
	if room.Children == nil {
		return []string{}
	}

	for _, child := range *room.Children {
		if child.Rid == nil || child.Rtype == nil {
			continue
		}
		switch *child.Rtype {
		case openhue.ResourceIdentifierRtypeLight:
			lightMap[*child.Rid] = struct{}{}
		case openhue.ResourceIdentifierRtypeDevice:
			device, err := service.home.GetDeviceById(*child.Rid)
			if err != nil || device.Services == nil {
				continue
			}
			for _, serviceRef := range *device.Services {
				if serviceRef.Rid == nil || serviceRef.Rtype == nil {
					continue
				}
				if *serviceRef.Rtype == openhue.ResourceIdentifierRtypeLight {
					lightMap[*serviceRef.Rid] = struct{}{}
				}
			}
		}
	}

	lightIDs := make([]string, 0, len(lightMap))
	for lightID := range lightMap {
		lightIDs = append(lightIDs, lightID)
	}
	sort.Strings(lightIDs)
	return lightIDs
}

func (service *Service) ensureInitialized() error {
	if service == nil || service.home == nil {
		return errServiceNotInitialized
	}
	return nil
}

type Group struct {
	Name   string  `json:"name"`
	Lights []Light `json:"lights"`
}

type Light struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}
