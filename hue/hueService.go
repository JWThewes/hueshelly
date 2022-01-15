package hue

import (
	"errors"
	"github.com/amimof/huego"
	"hueshelly/config"
	"hueshelly/logging"
	"strconv"
)

type Service struct {
	bridge *huego.Bridge
}

func New() Service {
	service := Service{}
	service.init()
	return service
}

func (service *Service) init() {
	var hueBridge *huego.Bridge
	if len(config.HueShellyConfig.HueBridgeIp) > 0 {
		logging.Logger.Println("Using bridge at " + config.HueShellyConfig.HueBridgeIp)
		hueBridge = huego.New(config.HueShellyConfig.HueBridgeIp, config.HueShellyConfig.HueUser)
	} else {
		logging.Logger.Println("Searching for bridge")
		discoveredBridge, err := huego.Discover()
		if err != nil {
			logging.Logger.Println("Error discovering bridge")
			logging.Logger.Fatal(err)
		}
		logging.Logger.Println("Found hue bridge at " + discoveredBridge.Host)
		hueBridge = discoveredBridge.Login(config.HueShellyConfig.HueUser)
	}
	_, err := hueBridge.GetUsers()
	if err != nil {
		logging.Logger.Println("Error communicating with bridge")
		logging.Logger.Fatal(err)
	}
	service.bridge = hueBridge
	logging.Logger.Println("Logged in at hue bridge")
}

func (service Service) ToggleLight(lightId int) error {
	light, err := service.bridge.GetLight(lightId)
	if err != nil {
		return err
	}
	if light.IsOn() {
		err := light.SetState(huego.State{On: false, TransitionTime: 0})
		if err != nil {
			return err
		}
		logging.Logger.Println("Light found - toggled to off")
		light = nil
		return err
	} else {
		err := light.SetState(huego.State{On: true, Bri: 254, TransitionTime: 0})
		if err != nil {
			return err
		}
		logging.Logger.Println("Light found - toggled to on")
		light = nil
		return nil
	}
}

func (service *Service) ToggleLightsInRoom(room string) error {
	groups, err := service.bridge.GetGroups()
	if err != nil {
		logging.Logger.Println("Error getting groups")
		return err
	}
	for i := range groups {
		group := groups[i]
		if group.Name == room {
			if group.GroupState.AnyOn {
				err := group.SetState(huego.State{On: false, TransitionTime: 0})
				if err != nil {
					return err
				}
				logging.Logger.Println("Group found - any lights on toggling to off")
				return nil
			} else {
				err := group.SetState(huego.State{On: true, Bri: 254, TransitionTime: 0})
				if err != nil {
					return err
				}
				logging.Logger.Println("Group found - all lights off toggling to on")
				return nil
			}
		}
	}
	groups = nil
	return errors.New("no room with name '" + room + "' found")
}

func (service *Service) AvailableGroups() []Group {
	hueGroups, _ := service.bridge.GetGroups()
	var groupList []Group
	for _, group := range hueGroups {
		groupStruct := Group{Name: group.Name}
		lightIds := group.Lights
		var lightList []Light
		for _, lightId := range lightIds {
			lightIdInt, _ := strconv.Atoi(lightId)
			light, _ := service.bridge.GetLight(lightIdInt)
			lightList = append(lightList, Light{
				Name: light.Name,
				Id:   light.ID,
			})
		}
		groupStruct.Lights = lightList
		groupList = append(groupList, groupStruct)
	}
	hueGroups = nil
	return groupList
}

type Group struct {
	Name   string  `json:"name"`
	Lights []Light `json:"lights"`
}

type Light struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}
