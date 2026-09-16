package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func checkReceivedDiscovery(page *rod.Page) (err error) {

	// ミッションリストを選択する（「デイリー（昼/夜）」というテキストを含む li 要素を直接指定）
	page.MustWaitIdle()

	li, err := page.Timeout(10 * time.Second).ElementX(
		"//li[contains(text(), 'SW2026ミッション')]")
	if err != nil {
		return fmt.Errorf("failed to find the li element: %w", err)
	}

	if _, err = li.WaitInteractable(); err != nil {
		return fmt.Errorf("li element not interactable: %w", err)
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
	// page.MustWaitIdle()
	sleep(1) // 進捗更新のための待機（これがないとボタンの一覧を取るところでエラーになる）

	//  ループで回す
	received := 0
	receivable := 0
	others := 0
	i := 0
	for ; i < 30; i++ {
		selector, op, err := mkSelector("SW2026", i)
		if err != nil {
			return fmt.Errorf("failed to make selector for button %d: %w", i, err)
		}
		log.Printf("SW2026-%d: selector=%s, op=%+v\n", i, selector, op)
		if op.AchieveText || op.Received || op.Comment || op.Gift > 0 {
			sleep(0.5)
			if op.AchieveText {
				tgt, err := page.Timeout(10 * time.Second).Element(selector + " .achieve-text")
				if err != nil {
					return fmt.Errorf("failed to Elements(selector) for tgt %d: %w", i, err)
				}
				if _, err = tgt.WaitInteractable(); err != nil {
					return fmt.Errorf("button not interactable: %w", err)
				}
				at, err := tgt.Timeout(10 * time.Second).Text()
				if err != nil {
					return fmt.Errorf("failed to get text for tgt %d: %w", i, err)
				}
				ata := strings.Split(at, "/")
				ano, _ := strconv.Atoi(ata[0])
				tno, _ := strconv.Atoi(ata[1])
				log.Printf("SW2026-%d: %d/%d\n", i, ano, tno)
				if op.Comment && ano < tno {
					sendComment(page, "39")
				} else if op.Gift > 0 && ano < tno {
					giftlist, err := checkGiftInventory(page)
					if err != nil {
						log.Printf("Error checking gift inventory: %v\n", err)
					} else {
						if giftlist[0].Count >= op.Gift {
							err = throwGift(page, giftlist[0].Name, fmt.Sprintf("%d", op.Gift))
							if err != nil {
								log.Printf("Error throwing gift: %v\n", err)
							}
						}
					}
				}
			}
			if op.Received {
				// 受け取り可能なボタンをクリックする処理
				btn, err := page.Timeout(10 * time.Second).Element(selector + " .achieve-button")
				if err != nil {
					return fmt.Errorf("failed to find element for button %d: %w", i, err)
				}

				// 各ボタンに対して状態を確認
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
				if classAttr != nil {
					if strings.Contains(*classAttr, "receivable") {
						log.Printf("Button %d is receivable, clicking...\n", i)
						receivable++
						if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
							log.Printf("Error clicking button %d: %v\n", i, err)
						}
						page.MustWaitIdle()

					} else if strings.Contains(*classAttr, "received") {
						log.Printf("Button %d is received\n", i)
						received++
					} else {
						log.Printf("Button %d is other: %s\n", i, *classAttr)
						others++
					}
				}
			}
		}
		log.Printf("receivable: %d, received: %d, others: %d\n",
			receivable, received, others)
		if received == 4 {
			err = fmt.Errorf(cmsg)
			return err
		}
	}
	return
}
