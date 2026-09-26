package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	// "github.com/go-rod/rod/lib/proto"
)

func selectMission(page *rod.Page, themeName string) error {

	page.MustWaitIdle()
	elements, err := page.Timeout(10 * time.Second).Elements("#mission-list ul.slide-container li")
	if err != nil {
		return fmt.Errorf("failed to get mission list elements: %w", err)
	}
	i := 0
	for ; i < len(elements); i++ {
		page.MustWaitIdle()
		selector := fmt.Sprintf("#mission-list ul.slide-container li:nth-child(%d)", i+1)
		txt, err := findElementAndGetText(page, selector, 0.4, 0.4)
		if err != nil {
			return fmt.Errorf("failed to get text for selector %s: %w", selector, err)
		}
		if strings.Contains(txt, themeName) {
			return nil
		}
		err = findElementAndClick(page, "#mission-list .slider-btn.slider-next", 0.4, 0.4)
		if err != nil {
			return fmt.Errorf("failed to click next button: %w", err)
		}
	}
	return fmt.Errorf("theme name '%s' not found in mission list", themeName)
}
