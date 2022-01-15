package hue

import (
	"errors"
	"github.com/amimof/huego"
	"hueshelly/config"
	"hueshelly/logging"
	"strconv"
)

type Service struct ***REMOVED***
	bridge *huego.Bridge
***REMOVED***

func New() Service ***REMOVED***
	service := Service***REMOVED******REMOVED***
	service.init()
	return service
***REMOVED***

func (service *Service) init() ***REMOVED***
	var hueBridge *huego.Bridge
	if len(config.HueShellyConfig.HueBridgeIp) > 0 ***REMOVED***
		logging.Logger.Println("Using bridge at " + config.HueShellyConfig.HueBridgeIp)
		hueBridge = huego.New(config.HueShellyConfig.HueBridgeIp, config.HueShellyConfig.HueUser)
	***REMOVED*** else ***REMOVED***
		logging.Logger.Println("Searching for bridge")
		discoveredBridge, err := huego.Discover()
		if err != nil ***REMOVED***
			logging.Logger.Println("Error discovering bridge")
			logging.Logger.Fatal(err)
		***REMOVED***
		logging.Logger.Println("Found hue bridge at " + discoveredBridge.Host)
		hueBridge = discoveredBridge.Login(config.HueShellyConfig.HueUser)
	***REMOVED***
	_, err := hueBridge.GetUsers()
	if err != nil ***REMOVED***
		logging.Logger.Println("Error communicating with bridge")
		logging.Logger.Fatal(err)
	***REMOVED***
	service.bridge = hueBridge
	logging.Logger.Println("Logged in at hue bridge")
***REMOVED***

func (service Service) ToggleLight(lightId int) error ***REMOVED***
	light, err := service.bridge.GetLight(lightId)
	if err != nil ***REMOVED***
		return err
	***REMOVED***
	if light.IsOn() ***REMOVED***
		err := light.SetState(huego.State***REMOVED***On: false, TransitionTime: 0***REMOVED***)
		if err != nil ***REMOVED***
			return err
		***REMOVED***
		logging.Logger.Println("Light found - toggled to off")
		light = nil
		return err
	***REMOVED*** else ***REMOVED***
		err := light.SetState(huego.State***REMOVED***On: true, Bri: 254, TransitionTime: 0***REMOVED***)
		if err != nil ***REMOVED***
			return err
		***REMOVED***
		logging.Logger.Println("Light found - toggled to on")
		light = nil
		return nil
	***REMOVED***
***REMOVED***

func (service *Service) ToggleLightsInRoom(room string) error ***REMOVED***
	groups, err := service.bridge.GetGroups()
	if err != nil ***REMOVED***
		logging.Logger.Println("Error getting groups")
		return err
	***REMOVED***
	for i := range groups ***REMOVED***
		group := groups[i]
		if group.Name == room ***REMOVED***
			if group.GroupState.AnyOn ***REMOVED***
				err := group.SetState(huego.State***REMOVED***On: false, TransitionTime: 0***REMOVED***)
				if err != nil ***REMOVED***
					return err
				***REMOVED***
				logging.Logger.Println("Group found - any lights on toggling to off")
				return nil
			***REMOVED*** else ***REMOVED***
				err := group.SetState(huego.State***REMOVED***On: true, Bri: 254, TransitionTime: 0***REMOVED***)
				if err != nil ***REMOVED***
					return err
				***REMOVED***
				logging.Logger.Println("Group found - all lights off toggling to on")
				return nil
			***REMOVED***
		***REMOVED***
	***REMOVED***
	groups = nil
	return errors.New("no room with name '" + room + "' found")
***REMOVED***

func (service *Service) AvailableGroups() []Group ***REMOVED***
	hueGroups, _ := service.bridge.GetGroups()
	var groupList []Group
	for _, group := range hueGroups ***REMOVED***
		groupStruct := Group***REMOVED***Name: group.Name***REMOVED***
		lightIds := group.Lights
		var lightList []Light
		for _, lightId := range lightIds ***REMOVED***
			lightIdInt, _ := strconv.Atoi(lightId)
			light, _ := service.bridge.GetLight(lightIdInt)
			lightList = append(lightList, Light***REMOVED***
				Name: light.Name,
				Id:   light.ID,
			***REMOVED***)
		***REMOVED***
		groupStruct.Lights = lightList
		groupList = append(groupList, groupStruct)
	***REMOVED***
	hueGroups = nil
	return groupList
***REMOVED***

type Group struct ***REMOVED***
	Name   string  `json:"name"`
	Lights []Light `json:"lights"`
***REMOVED***

type Light struct ***REMOVED***
	Name string `json:"name"`
	Id   int    `json:"id"`
***REMOVED***
