package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

func viewRoom(
	url string,
	viewingTime int,
	comment string,
) (
	err error,
) {
	log.Printf("viewRoom: url=%s, viewingTime=%d, comment=%s\n", url, viewingTime, comment)
	if srBrowser == nil {
		return fmt.Errorf("browser is not initialized")
	}
	if viewingTime <= 0 {
		return fmt.Errorf("viewingTime must be > 0")
	}

	page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()
	if err = applyJapaneseLocale(page); err != nil {
		return err
	}

	if err = page.Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate room: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait room page load: %w", err)
	}

	time.Sleep(time.Duration(viewingTime) * time.Second)
	return
}
