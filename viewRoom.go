package main

import (
	"fmt"
	"log"
	"net/http"
	// "strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/Chouette2100/srapi/v2"
)

const cmsg = "viewRoom: mission completed for room"

func viewRoom(
	page *rod.Page,
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

	// page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	// if err != nil {
	// 	return fmt.Errorf("failed to create page: %w", err)
	// }
	// defer page.Close()
	// if err = applyJapaneseLocale(page); err != nil {
	// 	return err
	// }

	if err = page.Navigate(room.URL); err != nil {
		return fmt.Errorf("failed to navigate room: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait room page load: %w", err)
	}

	ttime := time.Now().Add(time.Duration(viewingTime) * time.Second)
	log.Printf("viewRoom: waiting until %s (viewingTime=%d seconds)\n", ttime.Format("15:04:05"), viewingTime)

	// 告知のモーダルダイアログを閉じる
	mdb, err := page.Timeout(10 * time.Second).Element(
		".st-gift_bulk_sending_intro__close")
	if err == nil {

		if _, err = mdb.WaitInteractable(); err != nil {
			return fmt.Errorf("close button not interactable: %w", err)
		}

		if err := mdb.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("failed to click the close button: %w", err)
		}
		page.MustWaitIdle() // 描画が落ち着くのを待つ
	}

	// 一番下までスクロールして、トグルボタンを見えるようにする
	// page.MustEval(`window.scrollTo(0, document.body.scrollHeight)`)
	page.Mouse.Scroll(0, 100000, 1)
	page.MustWaitIdle() // 描画が落ち着くのを待つ

	// "ミッション"のダイアログのみ表示するために、トグルボタンの状態を同期する
	err = SyncActiveButton(page, "ミッション")
	if err != nil {
		return fmt.Errorf("SyncActiveButton: %w", err)
	}
	// 一番上までスクロールして、"ミッション"ボタンを表示させる
	// page.MustEval(`window.scrollTo(0, 0)`)
	page.Mouse.Scroll(0, -100000, 1)
	page.MustWaitIdle()

	if comment != "nil" {
		time.Sleep(time.Duration(dtmin) * time.Second)
		srapi.ApiLivePostLiveComment(client, comment, csrftoken, room.LiveID)
	}

	// ttimeまで待機する
	time.Sleep(time.Until(ttime))
	log.Printf("viewRoom: waiting finished at %s\n", time.Now().Format("15:04:05"))

	/*
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
	*/

	switch mission {
	case "daily":
		// 「デイリー（昼/夜）」というテキストを含む li 要素を直接指定
		li, err := page.Timeout(10 * time.Second).ElementX(
			"//li[contains(text(), 'デイリー（昼/夜）')]")
		if err != nil {
			return fmt.Errorf("failed to find the li element: %w", err)
		}

		if _, err = li.WaitInteractable(); err != nil {
			return fmt.Errorf("button not interactable: %w", err)
		}

		if err := li.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("failed to click the li element: %w", err)
		}
		page.MustWaitIdle()

		// 1. 共通する親要素からボタンをすべて取得するセレクタを指定
		// nth-child(n) を使わず、クラス名や構造で絞り込むのがコツです
		selector := "#mission-list .achieve-button"

		// 2. Elements() で全要素を取得
		buttons, err := page.Timeout(10 * time.Second).Elements(selector)
		if err != nil {
			return err
		}

		// 3. ループで回す
		received := 0
		receivable := 0
		others := 0
		progress := 0
		for i := range len(buttons) {
			nbuttons, err := page.Timeout(10 * time.Second).Elements(selector)
			if err != nil {
				return err
			}

			// 各ボタンに対して状態を確認
			btn := nbuttons[i]
			if _, err = btn.WaitInteractable(); err != nil {
				// return fmt.Errorf("btn not interactable: %w", err)
				log.Printf("Button %d not interactable: %v\n", i, err)
			}

			classAttr, err := btn.Attribute("class")
			if err != nil {
				log.Printf("Error getting class attribute for button %d: %v\n", i, err)
				continue
			}


			// nilチェックと判定
			// if classAttr != nil && strings.Contains(*classAttr, "receivable") {
			if classAttr != nil {
				if strings.Contains(*classAttr, "receivable") {
					log.Printf("Button %d is receivable, clicking...\n", i)
					receivable++
					if i != 4 { // 「キラキラ星 x500」は適切なタイミングに受け取る必要がある
						// クリック処理
						if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
							log.Printf("Error clicking button %d: %v\n", i, err)
							page.MustWaitIdle()
							continue // エラーが出ても次へ進むなどの制御が可能
						}
						page.MustWaitIdle()
					}
				} else if strings.Contains(*classAttr, "received") {
					log.Printf("Button %d is received\n", i)
					received++
				} else {
					log.Printf("Button %d is other: %s\n", i, *classAttr)
					others++
				}

				if i == 1 { // 視聴したルーム数を確認する
					progress, err = getProgressValue(btn)
					if err != nil {
						log.Printf("Error getting progress value from button %d: %v\n", i, err)
					} else {
						log.Printf("Progress value from button %d: %d\n", i, progress)
					}
				}
				// クリック後に画面が更新される場合は待機を入れる
				// page.WaitIdle(1 * time.Second)
				page.MustWaitIdle()
			}
		}
		log.Printf("Total buttons: %d, receivable: %d, received: %d, progress: %d, others: %d\n",
			len(buttons), receivable, received, progress, others)
		if progress == 20 {
			err = fmt.Errorf(cmsg)
			return err
		}

	case "newcommer":

		// 「新人ライバー応援キャンペーン」というテキストを含む li 要素を直接指定
		li, err := page.Timeout(10 * time.Second).ElementX(
			"//li[contains(text(), '新人ライバー応援キャンペーン')]")
		if err != nil {
			return fmt.Errorf("failed to find the li element: %w", err)
		}

		if _, err = li.WaitInteractable(); err != nil {
			return fmt.Errorf("button not interactable: %w", err)
		}

		if err := li.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("failed to click the li element: %w", err)
		}
		page.MustWaitIdle()

		// 1. 共通する親要素からボタンをすべて取得するセレクタを指定
		// nth-child(n) を使わず、クラス名や構造で絞り込むのがコツです
		selector := ".missions > div:nth-child(4) > ol:nth-child(2) > li > div > div > button"

		// time.Sleep(5 * time.Second) // ページが完全に読み込まれるまで待機

		// 2. Elements() で全要素を取得
		buttons, err := page.Timeout(10 * time.Second).Elements(selector)
		if err != nil {
			return err
		}

		// 3. ループで回す
		received := 0
		receivable := 0
		others := 0
		for i := range len(buttons) {
			nbuttons, err := page.Timeout(10 * time.Second).Elements(selector)
			if err != nil {
				return err
			}
			btn := nbuttons[i]
			if _, err = btn.WaitInteractable(); err != nil {
				return fmt.Errorf("btn not interactable: %w", err)
			}

			// 各ボタンに対して状態を確認
			classAttr, _ := btn.Attribute("class")

			// nilチェックと判定
			// if classAttr != nil && strings.Contains(*classAttr, "receivable") {
			if classAttr != nil {
				if strings.Contains(*classAttr, "receivable") {
					log.Printf("Button %d is receivable, clicking...\n", i)
					receivable++
					// クリック処理
					if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
						log.Printf("Error clicking button %d: %v\n", i, err)
						page.MustWaitIdle()
						continue // エラーが出ても次へ進むなどの制御が可能
					}
					page.MustWaitIdle()
				} else if strings.Contains(*classAttr, "received") {
					log.Printf("Button %d is received\n", i)
					received++
				} else {
					log.Printf("Button %d is other: %s\n", i, *classAttr)
					others++
				}

				// クリック後に画面が更新される場合は待機を入れる
				// page.WaitIdle(1 * time.Second)
			}
		}
		log.Printf("Total buttons: %d, receivable: %d, received: %d, others: %d\n", len(buttons), receivable, received, others)
		if received >= 10 {
			err = fmt.Errorf(cmsg)
			return err
		}
	default:
		return fmt.Errorf("unknown mission type: %s", mission)
	}

	/*
		// 2. Elements() で全要素を取得
		buttons, err = page.Elements(selector)
		if err != nil {
			return err
		}

		// 3. ループで回す
		nb := 0
		for i, btn := range buttons {
			// 各ボタンに対して状態を確認
			classAttr, _ := btn.Attribute("class")

			// nilチェックと判定
			if classAttr != nil && strings.Contains(*classAttr, "receivable") {
				log.Printf("Button %d is receivable, clicking...\n", i)
				nb++
				// クリック処理
				if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
					continue // エラーが出ても次へ進むなどの制御が可能
				}

				// クリック後に画面が更新される場合は待機を入れる
				page.WaitIdle(1 * time.Second)
			}
		}
		log.Printf("Total received buttons clicked: %d\n", nb)
	*/

	/*
		if (mission == "daily" || mission == "newcommer") &&
			(pmission.GenreList[0].Day.ContinuousMission[0].CurrentLevel ==
				pmission.GenreList[0].Day.ContinuousMission[0].MaxLevel ||
				pmission.GenreList[0].Night.ContinuousMission[0].CurrentLevel ==
					pmission.GenreList[0].Night.ContinuousMission[0].MaxLevel) {
			log.Printf(" %s Mission completed", mission)
			err = fmt.Errorf(cmsg)
		}
	*/
	return nil
}
