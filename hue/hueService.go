package hue

import (
	"errors"
	"hueshelly/config"
	"hueshelly/logging"
	"strconv"
	"strings"

	"github.com/openhue/openhue-go"
)

type Service struct {
	home *openhue.Home
}

func New() Service {
	service := Service{}
	service.init()
	return service
}

func (service *Service) init() {
	bridgeIP := strings.TrimSpace(config.HueShellyConfig.HueBridgeIp)
	if len(bridgeIP) > 0 {
		logging.Logger.Println("Using bridge at " + bridgeIP)
	} else {
		logging.Logger.Println("Searching for bridge")
		discoveredBridge, err := openhue.NewBridgeDiscovery().Discover()
		if err != nil {
			logging.Logger.Println("Error discovering bridge")
			logging.Logger.Fatal(err)
		}
		bridgeIP = discoveredBridge.IpAddress
		logging.Logger.Println("Found hue bridge at " + bridgeIP)
	}

	home, err := openhue.NewHome(bridgeIP, config.HueShellyConfig.HueUser)
	if err != nil {
		logging.Logger.Println("Error creating openhue home")
		logging.Logger.Fatal(err)
	}
	_, err = home.GetBridgeHome()
	if err != nil {
		logging.Logger.Println("Error communicating with bridge")
		logging.Logger.Fatal(err)
	}
	service.home = home
	logging.Logger.Println("Logged in at hue bridge")
}

func (service Service) ToggleLight(lightID int) error {
	light, err := service.findLightByID(lightID)
	if err != nil {
		return err
	}
	if light.Id == nil {
		return errors.New("light has no id")
	}
	if light.IsOn() {
		off := false
		err := service.home.UpdateLight(*light.Id, openhue.LightPut{On: &openhue.On{On: &off}})
		if err != nil {
			return err
		}
		logging.Logger.Println("Light found - toggled to off")
		return nil
	}

	on := true
	body := openhue.LightPut{On: &openhue.On{On: &on}}
	if !config.HueShellyConfig.RestorePreviousLightState {
		brightness := openhue.Brightness(100)
		body.Dimming = &openhue.Dimming{Brightness: &brightness}
	}
	err = service.home.UpdateLight(*light.Id, body)
	if err != nil {
		return err
	}
	logging.Logger.Println("Light found - toggled to on")
	return nil
}

func (service *Service) ToggleLightsInRoom(roomName string) error {
	rooms, err := service.home.GetRooms()
	if err != nil {
		logging.Logger.Println("Error getting rooms")
		return err
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

	return errors.New("no room with name '" + roomName + "' found")
}

func (service *Service) AvailableGroups() []Group {
	rooms, err := service.home.GetRooms()
	if err != nil {
		return []Group{}
	}
	lights, err := service.home.GetLights()
	if err != nil {
		return []Group{}
	}

	var groupList []Group
	for _, room := range rooms {
		groupStruct := Group{Name: nameFromRoom(room)}
		lightIDs := service.lightIDsFromRoom(room)
		var lightList []Light
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
				Id:   lightIDInt,
			})
		}
		groupStruct.Lights = lightList
		groupList = append(groupList, groupStruct)
	}

	return groupList
}

func (service Service) toggleGroupedLightByID(groupedLightID string) error {
	groupedLight, err := service.home.GetGroupedLightById(groupedLightID)
	if err != nil {
		return err
	}
	if groupedLight.Id == nil {
		return errors.New("grouped light has no id")
	}
	if groupedLight.IsOn() {
		off := false
		err := service.home.UpdateGroupedLight(*groupedLight.Id, openhue.GroupedLightPut{On: &openhue.On{On: &off}})
		if err != nil {
			return err
		}
		logging.Logger.Println("Group found - any lights on toggling to off")
		return nil
	}

	on := true
	body := openhue.GroupedLightPut{On: &openhue.On{On: &on}}
	if !config.HueShellyConfig.RestorePreviousLightState {
		brightness := openhue.Brightness(100)
		body.Dimming = &openhue.Dimming{Brightness: &brightness}
	}
	err = service.home.UpdateGroupedLight(*groupedLight.Id, body)
	if err != nil {
		return err
	}
	logging.Logger.Println("Group found - all lights off toggling to on")
	return nil
}

func (service Service) findLightByID(lightID int) (*openhue.LightGet, error) {
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
	return nil, errors.New("light with id '" + strconv.Itoa(lightID) + "' not found")
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
	for _, service := range *room.Services {
		if service.Rid == nil || service.Rtype == nil {
			continue
		}
		if *service.Rtype == openhue.ResourceIdentifierRtypeGroupedLight {
			return *service.Rid, true
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

func (service Service) lightIDsFromRoom(room openhue.RoomGet) []string {
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

	var lightIDs []string
	for lightID := range lightMap {
		lightIDs = append(lightIDs, lightID)
	}
	return lightIDs
}

type Group struct {
	Name   string  `json:"name"`
	Lights []Light `json:"lights"`
}

type Light struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}
