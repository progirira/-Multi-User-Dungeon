package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Room struct {
	Name          string
	IsLocked      bool
	ThingsInRoom  map[string]bool // Bool state - Is in the room(True) or Taken(false)
	RoomsAbleToGo []string
	LookAround    func(things []string, rooms []string) string
	Go            func(rooms []string) string
}

var OrderedThings []string

var World map[string]Room

var ActionToThing = map[string]string{
	"рюкзак":    "надеть",
	"конспекты": "взять",
	"ключи":     "взять",
}

var FuncToAction map[string]func(string) string

type BagType struct {
	Name         string
	IsWithPlayer bool
}

type Player struct {
	Location            Room
	Bag                 BagType
	ThingsInBagAndState map[string]bool // Bool state - Is with player(True) or Not(false)
	GameDone            bool
}

var CurrentPlayer Player

func (player *Player) setGameDone() {
	player.GameDone = true
}

func (player *Player) ifGameDone() bool {
	return player.GameDone
}

func (room *Room) removeThingFromRoom(thing string) {
	room.ThingsInRoom[thing] = false
}

func (player *Player) ReplaceAThing(thing string) bool {
	switch {
	case thing == player.Bag.Name:
		player.Bag.IsWithPlayer = true
		player.Location.removeThingFromRoom(thing)
		return true
	case player.Bag.IsWithPlayer:
		player.ThingsInBagAndState[thing] = true
		player.Location.removeThingFromRoom(thing)
		return true
	}
	return false
}

func printDoneAction(action string, thing string) string {
	return FuncToAction[action](thing)
}

func getListOfThingsInRoom(mapOfThings map[string]bool) []string {
	resultList := make([]string, 0, len(mapOfThings))
	for _, key := range OrderedThings {
		if value, keyExist := mapOfThings[key]; keyExist && value {
			resultList = append(resultList, key)
		}
	}
	return resultList
}

func returnListWithoutElement(things []string, thingToDelete string) []string {
	resultList := make([]string, 0, len(things)-1)
	for _, thing := range things {
		if thing != thingToDelete {
			resultList = append(resultList, thing)
		}
	}
	return resultList
}

func LookAroundInBedroom(things []string, rooms []string) string {
	if len(things) == 0 {
		return fmt.Sprintf("пустая комната. можно пройти - %s", strings.Join(rooms, ", "))
	}
	thingsOnTable := things
	thingsOnChair := []string{}

	if !CurrentPlayer.Bag.IsWithPlayer {
		thingsOnTable = returnListWithoutElement(things, CurrentPlayer.Bag.Name)
		thingsOnChair = []string{CurrentPlayer.Bag.Name}
	}

	result := ""
	switch {
	case len(thingsOnTable) != 0 && len(thingsOnChair) != 0:
		result = fmt.Sprintf("на столе: %s, на стуле: %s. ", strings.Join(thingsOnTable, ", "),
			strings.Join(thingsOnChair, ", "))
	case len(thingsOnTable) != 0:
		result = fmt.Sprintf("на столе: %s. ", strings.Join(thingsOnTable, ", "))
	default:
		result = fmt.Sprintf(", на стуле: %s. ", strings.Join(thingsOnChair, ", "))
	}
	return result + fmt.Sprintf("можно пройти - %s", strings.Join(rooms, ", "))
}

func LookAroundInKitchen(things []string, rooms []string) string {

	if CurrentPlayer.Bag.IsWithPlayer {
		return fmt.Sprintf("ты находишься на кухне, на столе: %s, надо идти в универ. можно пройти - %s",
			strings.Join(things, ", "), strings.Join(rooms, ", "))
	} else {
		return fmt.Sprintf("ты находишься на кухне, на столе: %s, надо собрать рюкзак и идти в универ. можно пройти - %s",
			strings.Join(things, ", "), strings.Join(rooms, ", "))
	}
}

func LookAroundInHall(rooms []string) string {
	return fmt.Sprintf("ты находишься в коридоре. можно пройти - %s", strings.Join(rooms, ", "))
}

func WhenPlayerCame(beginning string, rooms []string) string {
	return beginning + strings.Join(rooms, ", ")
}

func UnlockTheDoor() {
	for _, roomName := range CurrentPlayer.Location.RoomsAbleToGo {
		room := World[roomName]
		room.IsLocked = false
		World[roomName] = room
	}
}

func SetGameDone() {
	CurrentPlayer.setGameDone()
}

func initGame() {
	FuncToAction = map[string]func(string) string{
		"надеть": func(thing string) string {
			return fmt.Sprintf("вы надели: %s", thing)
		},
		"взять": func(thing string) string {
			return fmt.Sprintf("предмет добавлен в инвентарь: %s", thing)
		},
		"выпить": func(thing string) string {
			return fmt.Sprintf("вы выпили: %s", thing)
		},
		"применить ключи": func(thing string) string {
			if thing != "дверь" {
				return "не к чему применить"
			}
			UnlockTheDoor()
			return "дверь открыта"
		},
	}
	OrderedThings = []string{"ключи", "конспекты", "рюкзак", "чай"}

	var kitchen = Room{
		Name:          "кухня",
		IsLocked:      false,
		ThingsInRoom:  map[string]bool{"чай": true},
		RoomsAbleToGo: []string{"коридор"},
		LookAround:    LookAroundInKitchen,
		Go: func(rooms []string) string {
			return WhenPlayerCame("кухня, ничего интересного. можно пройти - ", rooms)
		},
	}

	var hall = Room{
		Name:          "коридор",
		IsLocked:      false,
		ThingsInRoom:  map[string]bool{},
		RoomsAbleToGo: []string{"кухня", "комната", "улица"},
		LookAround: func(things []string, rooms []string) string {
			return LookAroundInHall(rooms)
		},
		Go: func(rooms []string) string {
			return WhenPlayerCame("ничего интересного. можно пройти - ", rooms)
		},
	}

	var bedroom = Room{
		Name:     "комната",
		IsLocked: false,
		ThingsInRoom: map[string]bool{"ключи": true,
			"конспекты": true,
			"рюкзак":    true,
		},
		RoomsAbleToGo: []string{"коридор"},
		LookAround:    LookAroundInBedroom,
		Go: func(rooms []string) string {
			return WhenPlayerCame("ты в своей комнате. можно пройти - ", rooms)
		},
	}

	var street = Room{
		Name:          "улица",
		IsLocked:      true,
		ThingsInRoom:  map[string]bool{},
		RoomsAbleToGo: []string{"домой"},
		LookAround: func(things []string, rooms []string) string {
			return "на улице весна..."
		},
		Go: func(rooms []string) string {
			SetGameDone()
			return WhenPlayerCame("на улице весна. можно пройти - ", rooms)
		},
	}

	World = map[string]Room{
		"кухня":   kitchen,
		"коридор": hall,
		"комната": bedroom,
		"улица":   street,
	}

	CurrentPlayer = Player{
		Location: World["кухня"],
		Bag: BagType{Name: "рюкзак",
			IsWithPlayer: false},
		ThingsInBagAndState: map[string]bool{"ключи": false,
			"конспекты": false,
			"рюкзак":    false,
		},
		GameDone: false,
	}
}

func ifThingIsWithPlayer(thing string) bool {
	return CurrentPlayer.ThingsInBagAndState[thing]
}

func ifRoomIsAble(room string, ableRooms []string) bool {
	return slices.Contains(ableRooms, room)
}

func ifRoomIsLocked(room string) bool {
	return World[room].IsLocked
}

func (player *Player) ChangeLocation(room Room) {
	player.Location = room
}

func doGo(roomName string) string {
	if ifRoomIsAble(roomName, CurrentPlayer.Location.RoomsAbleToGo) {
		if ifRoomIsLocked(roomName) {
			return "дверь закрыта"
		}
		CurrentPlayer.ChangeLocation(World[roomName])
		return World[roomName].Go(World[roomName].RoomsAbleToGo)
	} else {
		return "нет пути в " + roomName
	}
}

func ifActionIsUpToObject(commandAction string, object string) bool {
	action := ActionToThing[object]
	return action == commandAction
}

func ifThingIsInRoom(thing string) bool {
	return CurrentPlayer.Location.ThingsInRoom[thing]
}

func Apply(verb string, thingWhatApply string, thingWhereApply string) string {
	action := strings.Join([]string{verb, thingWhatApply}, " ")
	return FuncToAction[action](thingWhereApply)
}

func handleCommand(command string) string {
	var CommandWords = strings.Fields(command)
	var NumOfWords = len(CommandWords)

	switch {
	case NumOfWords == 1 && CommandWords[0] == "осмотреться":
		things := getListOfThingsInRoom(CurrentPlayer.Location.ThingsInRoom)
		rooms := CurrentPlayer.Location.RoomsAbleToGo
		return CurrentPlayer.Location.LookAround(things, rooms)

	case NumOfWords == 2:
		if CommandWords[0] == "идти" {
			return doGo(CommandWords[1])
		}
		if !ifThingIsInRoom(CommandWords[1]) {
			return "нет такого"
		}
		if ifActionIsUpToObject(CommandWords[0], CommandWords[1]) {
			success := CurrentPlayer.ReplaceAThing(CommandWords[1])
			if success {
				return printDoneAction(CommandWords[0], CommandWords[1])
			} else {
				return "некуда класть"
			}
		}

	case NumOfWords == 3 && CommandWords[0] == "применить":
		if !ifThingIsWithPlayer(CommandWords[1]) {
			return "нет предмета в инвентаре - " + CommandWords[1]
		}
		return Apply(CommandWords[0], CommandWords[1], CommandWords[2])
	}
	return "неизвестная команда"
}

func main() {
	initGame()
	reader := bufio.NewReader(os.Stdin)
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Invalid input: %s", err)
		}
		command = strings.TrimSuffix(command, "\n")
		fmt.Println(handleCommand(command))

		if CurrentPlayer.ifGameDone() {
			break
		}
	}
}
