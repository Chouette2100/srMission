package main

import (
	"fmt"
	"log"
	"net/http"
	// "sort"
	"time"

	"github.com/Chouette2100/srapi/v2"
)

/*
type Room struct {
	MainName  string
	URL       string
	RoomID    int
	LiveID    int
	Starttime time.Time
}
*/

func collectEventRooms(
	eventurlkey string,
	noofrooms int,
	collectedAt time.Time,
	accessHistory AccessHistory,
) (
	rooms []Room,
	err error,
) {
	var prooms *srapi.EventRooms
	prooms, err = srapi.GetEventRoomsByApi(http.DefaultClient, eventurlkey, 1, 1000)

	for _, proom := range prooms.Rooms {
		// if chkDup(live.RoomID, mission) {
		// 	continue
		// }
		if proom.IsOnLive {
			room := Room{
				MainName: proom.RoomName,
				URL:      proom.RoomURLKey,
				RoomID:   proom.RoomID,
			}
			if blocked, ok := isRoomURLBlacklisted(room.URL); ok {
				log.Printf("collectRooms: skip blacklisted url=%s reason=%s\n", room.URL, blocked.Reason)
				continue
			}
			rooms = append(rooms, room)
		}
	}
	if err != nil {
		log.Printf("Error: %v\n", err)
		return nil, fmt.Errorf("failed to collectRooms(%s): %w", eventurlkey, err)
	}
	themeID := missionThemeID("event")
	rooms = filterRoomsByAccessPolicy(rooms, themeID, collectedAt, accessHistory)

	// sort.Slice(rooms, func(i, j int) bool {
	// 	return rooms[i].Starttime.After(rooms[j].Starttime)
	// })

	if len(rooms) > noofrooms {
		rooms = rooms[0:noofrooms]
	}

	return rooms, err
}
