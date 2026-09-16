package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func collectMyNextFave(page *rod.Page) (rooms []Room, err error) {

	rooms = []Room{}

	if err = page.Navigate("https://www.showroom-live.com/"); err != nil {
		return rooms, fmt.Errorf("failed to navigate top page: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return rooms, fmt.Errorf("failed to wait top page load: %w", err)
	}

	// 1. 指定したセレクタに一致する要素をすべて取得
	elements, err := page.Timeout(10 * time.Second).Elements("a[data-category-id='15']")
	if err != nil {
		return rooms, fmt.Errorf("failed to get elements: %w", err)
	}

	for _, el := range elements {
		// 2. href属性を取得
		href := el.MustAttribute("href")
		if href != nil {
			// 3. 文字列操作でID部分を抽出
			// 例: "/r/nmb48_12add_62" -> "nmb48_12add_62"
			url := strings.TrimPrefix(*href, "/r/")
			log.Printf("抽出したURL: %s\n", url)
			room := Room{MainName: url, URL: url}
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}
