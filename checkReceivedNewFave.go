  package main
  import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
  )

  func checkReceivedNewFave(page *rod.Page) (err error) {

		// ミッションリストを選択する（「デイリー（昼/夜）」というテキストを含む li 要素を直接指定）
		page.MustWaitIdle()

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

		// 進捗を更新する
		btn, err := page.Timeout(10 * time.Second).Element(".reload-button")
		if err != nil {
			return fmt.Errorf("failed to find the reload button: %w", err)
		}

		if _, err = btn.WaitInteractable(); err != nil {
			return fmt.Errorf("reload button not interactable: %w", err)
		}

		if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("failed to click the reload button: %w", err)
		}
		page.MustWaitIdle()
		sleep(5) // 進捗更新のための待機（これがないとボタンの一覧を取るところでエラーになる）

		// 1. 共通する親要素からボタンをすべて取得するセレクタを指定
		// nth-child(n) を使わず、クラス名や構造で絞り込むのがコツです
		selector := "#mission-list .achieve-button"

		// 3. ループで回す
		received := 0
		receivable := 0
		others := 0
		progress := 0
		i := 0
		for ; i < 6; i++ {
			sleep(1)
			// 2. Elements() で全要素を取得
			buttons, err := page.Timeout(10 * time.Second).Elements(selector)
			if err != nil {
				return fmt.Errorf("failed to Elements(selector): %w", err)
			}
			if i >= len(buttons) {
				log.Printf("Button index %d out of range, total buttons: %d\n", i, len(buttons))
				break
			}

			// 各ボタンに対して状態を確認
			btn := buttons[i]
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
					hour := time.Now().Hour()
					if i != 4 || hour == 2 || hour == 14 { // 「キラキラ星 x500」は適切なタイミングに受け取る必要がある
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
					if i == 2 && strings.Contains(*classAttr, "challenging") {
						// まだ一度もギフトを投げていない
						giftlist, err := checkGiftInventory(page)
						if err != nil {
							log.Printf("Error checking gift inventory: %v\n", err)
						} else {
							if giftlist[0].Count >= 10 {
								err = throwGift(page, giftlist[0].Name, "10")
								if err != nil {
									log.Printf("Error throwing gift: %v\n", err)
								}
							}	
						}
					}
				}

				if i == 1 { // 視聴したルーム数を確認する
					progress, err = getProgressValue(page)
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
		log.Printf("receivable: %d, received: %d, others: %d (progress: %d)\n",
			receivable, received, others, progress)
		if progress == 20 {
			err = fmt.Errorf(cmsg)
			return err
		}
		return nil
	}
