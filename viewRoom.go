package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-rod/rod/lib/proto"

	"github.com/Chouette2100/srapi/v2"
)

const cmsg = "viewRoom: mission completed for room"

func viewRoom(
	client *http.Client,
	csrftoken string,
	mission string,
	room Room,
	viewingTime int,
	comment string,
) (
	err error,
) {

	dtmin := 3 // 最小待ち時間

	log.Printf("viewRoom: url=%s, viewingTime=%d, comment=%s\n", room.URL, viewingTime, comment)
	if srBrowser == nil {
		return fmt.Errorf("browser is not initialized")
	}
	if viewingTime <= 0 {
		return fmt.Errorf("viewingTime must be > 0")
	} else if viewingTime < dtmin*2 {
		viewingTime = dtmin * 2
	}

	page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()
	if err = applyJapaneseLocale(page); err != nil {
		return err
	}

	if err = page.Navigate(room.URL); err != nil {
		return fmt.Errorf("failed to navigate room: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait room page load: %w", err)
	}

	if comment != "nil" {
		time.Sleep(time.Duration(dtmin) * time.Second)
		srapi.ApiLivePostLiveComment(client, comment, csrftoken, room.LiveID)
		time.Sleep(time.Duration(viewingTime-dtmin) * time.Second)
	} else {
		time.Sleep(time.Duration(viewingTime) * time.Second)
	}

	var pmission *srapi.Mission
	pmission, err = srapi.ApiMission(client, strconv.Itoa(room.RoomID))
	if err != nil {
		return fmt.Errorf("srapi.ApiMission: %w", err)
	}

	for i, genre := range pmission.GenreList {
		log.Printf("[%d] %s\n", i, genre.Name)

		log.Printf("  Night\n")
		for j, single := range genre.Night.SingleMission {
			log.Printf("    Single[%d]  %d / %d %s\n", j, single.CurrentLevel, single.MaxLevel, single.Title)
		}
		log.Printf("    Composite   %d / %d %s\n",
			genre.Night.CompositeMission.CurrentLevel,
			genre.Night.CompositeMission.MaxLevel,
			genre.Night.CompositeMission.Title,
		)
		for j, continuous := range genre.Night.ContinuousMission {
			log.Printf("    Continuous[%d]  %d / %d %s\n", j, continuous.CurrentLevel, continuous.MaxLevel, continuous.Title)
		}

		log.Printf("  Day\n")
		for j, single := range genre.Day.SingleMission {
			log.Printf("    Single[%d]  %d / %d %s\n", j, single.CurrentLevel, single.MaxLevel, single.Title)
		}
		log.Printf("    Composite   %d / %d %s\n",
			genre.Day.CompositeMission.CurrentLevel,
			genre.Day.CompositeMission.MaxLevel,
			genre.Day.CompositeMission.Title,
		)
		for j, continuous := range genre.Day.ContinuousMission {
			log.Printf("    Continuous[%d]  %d / %d %s\n", j, continuous.CurrentLevel, continuous.MaxLevel, continuous.Title)
		}
	}

	if mission == "daily" && (pmission.GenreList[0].Day.ContinuousMission[0].CurrentLevel ==
		pmission.GenreList[0].Day.ContinuousMission[0].MaxLevel ||
		pmission.GenreList[0].Night.ContinuousMission[0].CurrentLevel ==
			pmission.GenreList[0].Night.ContinuousMission[0].MaxLevel) {
		log.Printf(" %s Mission completed", mission)
		err = fmt.Errorf(cmsg)
	}

	return
}
