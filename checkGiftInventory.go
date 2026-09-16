package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"log"
	"strconv"
	"strings"
	"time"
)

type Gift struct {
	Name  string
	Count int
}

func checkGiftInventory(page *rod.Page) (giftlist []Gift, err error) {
	giftlist = []Gift{
		{Name: "3000669", Count: 0},
		{Name: "3000670", Count: 0},
		{Name: "3000671", Count: 0},
		{Name: "3000672", Count: 0},
		{Name: "3000673", Count: 0},
		{Name: "3000841", Count: 0},
		{Name: "3000842", Count: 0},
		{Name: "3000843", Count: 0},
		{Name: "3000844", Count: 0},
		{Name: "3000845", Count: 0},
	}
	//  .st-gift_box div[gift_id="3000670"] .button
	// .st-gift_box div[gift_id="3000670"] .num      text "x 120"

	for i, gift := range giftlist {
		selector := fmt.Sprintf(".st-gift_box div[gift_id=\"%s\"] .num", gift.Name)
		elem, err := page.Timeout(10 * time.Second).Element(selector)
		if err != nil {
			return nil, fmt.Errorf("failed to find element for gift %s: %w", gift.Name, err)
		}
		text, err := elem.Text()
		if err != nil {
			log.Printf("Error: failed to get text for gift %s: %v\n", gift.Name, err)
		}
		// text is expected to be in the format "x 120"
		var count int
		texta := strings.Split(text, " ")
		if len(texta) != 2 {

			log.Printf("Error: unexpected text format for gift %s: %s\n", gift.Name, text)
			continue
		} else {
			count, err = strconv.Atoi(texta[1])
			if err != nil {
				log.Printf("Error: failed to parse count for gift %s: %v\n", gift.Name, err)
			}
			giftlist[i].Count = count
		}
	}

	// giftlist[]をCountの降順にソートする
	for i := 0; i < len(giftlist)-1; i++ {
		for j := 0; j < len(giftlist)-i-1; j++ {
			if giftlist[j].Count < giftlist[j+1].Count {
				giftlist[j], giftlist[j+1] = giftlist[j+1], giftlist[j]
			}
		}
	}

	for _, gift := range giftlist {
		log.Printf("Gift ID: %s, Count: %d\n", gift.Name, gift.Count)
	}

	return giftlist, nil
}
// 指定したセットのギフトを投げる
func throwGift(page *rod.Page, giftID string, count string) error {

	sleep(1)
	err := page.WaitIdle(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for page idle: %w", err)
	}

	selector := fmt.Sprintf("li.gift div[gift_id=\"%s\"] button", giftID)
	log.Printf("throwGift: clicking gift button for gift %s\n", selector)
	gbtn, err := page.Timeout(20 * time.Second).Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element for gift %s: %w", giftID, err)
	}
	if _, err = gbtn.WaitInteractable(); err != nil {
		return fmt.Errorf("button not interactable for gift %s: %w", giftID, err)
	}
	if err := gbtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click button for gift %s: %w", giftID, err)
	}

	switch count {
	case "10":
		selector = ".st-gift_quantity_selector__list li:last-child .st-gift_quantity_selector__button"
	default:
		return fmt.Errorf("invalid count: %s", count)
	}

	btn, err := page.Timeout(10 * time.Second).Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element for gift %s with count %s: %w", giftID, count, err)
	}
	if _, err = btn.WaitInteractable(); err != nil {
		return fmt.Errorf("button not interactable for gift %s with count %s: %w", giftID, count, err)
	}
	if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click button for gift %s with count %s: %w", giftID, count, err)
	}

	cfm, err := page.Timeout(5 * time.Second).Element(".st-bulk_gift_confirm_modal__submit")
	if err != nil {
		return fmt.Errorf("confirm button not found %s: %w", giftID, err)
	}
	if _, err = cfm.WaitInteractable(); err != nil {
		return fmt.Errorf("confirm button not interactable for gift %s: %w", giftID, err)
	}
	if err := cfm.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click confirm button for gift %s: %w", giftID, err)
	}
	log.Printf("throwGift: gift %s with count %s thrown successfully\n", giftID, count)
	return nil
}