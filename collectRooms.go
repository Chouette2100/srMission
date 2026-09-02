package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/Chouette2100/srapi/v2"
)

type Room struct {
	MainName  string
	URL       string
	RoomID    int
	LiveID    int
	Starttime time.Time
}

func collectRooms(
	mission string,
	noofrooms int,
) (
	rooms []Room,
	err error,
) {
	var lives []srapi.Lives2
	lives, err = srapi.GetLiveOnlives3(http.DefaultClient, []int{
		102, 103, 104, 105, 107, 108, 109, 110, 112, 113,
		200,
	})

	for _, live := range lives {
		room := Room{
			MainName:  live.MainName,
			URL:       "https://showroom-live.com/r/" + live.RoomURLKey,
			RoomID:    live.RoomID,
			LiveID:    live.LiveID,
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

	if len(rooms) > noofrooms {
		rooms = rooms[0:noofrooms]
	}

	return rooms, err
}
