package main
import (
	"fmt"
	"log"
	"sort"
	"time"
	"net/http"

	"github.com/Chouette2100/srapi/v2"
)
	type Room struct {
		MainName string
		URL  string
		RoomID int
		Starttime time.Time
	}

func collectRooms(
	mission string,
) (
	rooms []Room,
	err error,
){
	var lives []srapi.Lives2
	lives, err = srapi.GetLiveOnlives3(http.DefaultClient, []int{200})

	for i, live := range lives {
		if i >= 2 {
			break
		}
		room := Room{
			MainName: live.MainName,
			URL:  "https://showroom-live.com/r/" + live.RoomURLKey,
			RoomID: live.RoomID,
			Starttime: time.Unix(live.StartedAt, 0),
		}
		rooms = append(rooms, room)
	}
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, fmt.Errorf("failed to collectRooms(%s): %w", mission, err)
	}

	sort.Slice(rooms, func(i, j int) bool {
	 	return rooms[i].Starttime.After(rooms[j].Starttime)
	}) 

	return rooms, err
}