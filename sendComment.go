 package main

import (
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func sendComment(page *rod.Page, comment string) {
	if comment != "nil" {
		//	time.Sleep(time.Duration(dtmin) * time.Second)
		//	srapi.ApiLivePostLiveComment(client, comment, csrftoken, room.LiveID)
		if comment == "j" {
			// 朝、昼、夜で挨拶を変える
			hour := time.Now().Hour()
			if hour > 3 && hour < 10 {
				comment = "おはようございます"
			} else if hour < 18 {
				comment = "こんにちは"
			} else {
				comment = "こんばんは"
			}
		}
		cinput, err := page.Timeout(10 * time.Second).Element(".comment-input input")
		if err == nil {
			if _, err = cinput.WaitInteractable(); err == nil {
				err = cinput.Input(comment)
				if err != nil {
					log.Printf("Error: failed to input comment: %v\n", err)
				} else {
					cbtn, err := page.Timeout(10 * time.Second).Element(".st-comment__button")
					if err == nil {
						if _, err = cbtn.WaitInteractable(); err == nil {
							err = cbtn.Click(proto.InputMouseButtonLeft, 1)
							if err != nil {
								log.Printf("Error: failed to click comment button: %v\n", err)
							} else {
								log.Printf("Comment posted: %s\n", comment)
							}
						}
					}
				}
			}
		}
	}
}