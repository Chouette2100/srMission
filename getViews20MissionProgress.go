package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-rod/rod"
)

func getViews20MissionProgress(page *rod.Page, selector string) (int, error) {
	// sleep(1.0)
	el, err := page.Timeout(15 * time.Second).Element(selector + " button span")
	if err != nil {
		// log.Printf("Error finding element for selector %s: %v\n", selector+" button span", err)
		// もっと厳密にやるのであれば button のラベルが受取済になっていることを確かめる
		return 0, nil
	} else {
		text, err := el.Text()
		if err != nil {
			log.Printf("Error getting text for element %s: %v\n", selector+" button span", err)
		} else {
			rno, _ := strconv.Atoi(text)
			return rno, nil
		}
	}
	return 0, fmt.Errorf("failed to get views20 mission progress for selector %s", selector)
}
