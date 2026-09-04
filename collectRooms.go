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
	type Genres struct{ IDs []int }
	var genreList = map[string]Genres{
		"daily": {IDs: []int{
			// 0, // 人気
			112, // ミュージック
			102, // アイドル
			103, // タレント
			104, // 声優
			105, // 芸人
			107, // バーチャル
			108, // モデル
			109, // 俳優
			110, // アナウンサー
			112, // クリエイター
			200, // ライバー
			// 704, // メンズ
			// 703, // カラオケ
		}},
		"newcommer": {IDs: []int{
			762, // New1day
			763, // New7day
			764, // New30day
			// 758, // 注目の新人
		}},
		// 「きっかけ配信」はジャンルではない！
		"discovery" : {IDs: []int{}},
	}
	genre := genreList[mission]
	var lives []srapi.Lives2
	lives, err = srapi.GetLiveOnlives3(http.DefaultClient, genre.IDs)

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
