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

type Mission struct {
	Name        string
	Selector    string
	AchieveText bool
	Received    bool
	Comment     string
	GiftType    string
	Gift        int
}

type Thema struct {
	Order    int
	Name     string
	Selector string
	Missions []Mission
}

type ThemaList map[string]Thema

var themaList = ThemaList{
	"Daily": Thema{
		Order: 0,
		Name:  "デイリー（昼/夜）",
		Missions: []Mission{
			{Name: "Completed",Selector: ".missions > ol:nth-child(3) > li:nth-child(1)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Name: "20 views", Selector: "div.content:nth-child(4) > ol:nth-child(2) > li:nth-child(1)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: "div.content:nth-child(4) > ol:nth-child(4) > li:nth-child(1)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 10},
			{Selector: "div.content:nth-child(4) > ol:nth-child(4) > li:nth-child(2)",
				AchieveText: false, Received: true, Comment: "39", GiftType: "", Gift: 0},
			{Selector: "div.content:nth-child(4) > ol:nth-child(4) > li:nth-child(3)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: "div.content:nth-child(4) > ol:nth-child(4) > li:nth-child(4)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 0},
		},
	},
	"SW2026-Sep.": Thema{
		Order: 1,
		Name:  "SW2026ミッション",
		Missions: []Mission{
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(1)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(2)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(3)",
				AchieveText: false, Received: true, Comment: "39", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(4)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 10},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(5)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(6)",
				AchieveText: true, Received: true, Comment: "39", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(7)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 10},
			{Name: "Campaign", Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(8)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(9)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(10)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(11)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(12)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(13)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(14)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(15)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(16)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Name: "Campaign Period I", Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(1)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(2)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(3)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(4)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(5)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(6)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(2) > li:nth-child(7)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Name: "Campaign Period II", Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(1)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(2)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(3)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(4)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(5)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(6)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(5) > ol:nth-child(4) > li:nth-child(7)",
				AchieveText: false, Received: false, Comment: "nil", GiftType: "", Gift: 0},
		},
	},
	"SW2026-NewCommer": Thema{
		Order: 2,
		Name:  "SW2026 新人ライバー応援ミッション",
		Missions: []Mission{
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(1)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(2)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(3)",
				AchieveText: true, Received: true, Comment: "39", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(4)",
				AchieveText: false, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(5)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(6)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(7)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "", Gift: 0},
			{Selector: ".missions > div:nth-child(4) > ol:nth-child(2) > li:nth-child(8)",
				AchieveText: true, Received: true, Comment: "nil", GiftType: "starsAndSeeds", Gift: 10},
		},
	},
}

func achieveAndReceiveMission(page *rod.Page, themaID string) (err error) {

	theme := themaList[themaID]
	// ミッションリストを選択する（「デイリー（昼/夜）」というテキストを含む li 要素を直接指定）
	page.MustWaitIdle()

	/*
		// li, err := page.Timeout(10 * time.Second).ElementX(theme.Selector)
		li, err := page.Timeout(10 * time.Second).Element(theme.Selector)
		if err != nil {
			return fmt.Errorf("failed to find the li element: %w", err)
		}
	*/

	if theme.Order > 0 {
		snext, err := page.Timeout(10 * time.Second).Element("#mission-list .slider-btn.slider-next")
		if err != nil {
			return fmt.Errorf("failed to find the li element: %w", err)
		}

		if _, err = snext.WaitInteractable(); err != nil {
			return fmt.Errorf("snext element not interactable: %w", err)
		}

		for i := 0; i < theme.Order; i++ {
			if err := snext.Click(proto.InputMouseButtonLeft, 1); err != nil {
				return fmt.Errorf("failed to click the snext element: %w", err)
			}
			if i < theme.Order-1 {
				sleep(0.4)
			}
		}
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
	treceived := 0
	receivable := 0
	others := 0
	i := 0
	for ; i < len(theme.Missions); i++ {
		selector := theme.Missions[i].Selector
		op := theme.Missions[i]
		log.Printf("SW2026-%d: selector=%s, op=%+v\n", i, selector, op)
		breceived := false
		tno := 0
		ano := 0
		rno := 0
		if op.AchieveText || op.Received {
			// 進捗を確認する必要がある場合
			sleep(0.5)
			if op.AchieveText {
				// 進捗が分割されているミッション
				selector_txt := selector + "  .achieve-text"
				if theme.Order == 0 && i == 1 {
					selector_txt = selector + "  .received-num"
				}
				tgt, err := page.Timeout(10 * time.Second).Element(selector_txt)
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
				if theme.Order == 0 && i == 1 {
					at = strings.TrimPrefix(at, "受取済")
				}
				ata := strings.Split(at, "/")
				ano, _ = strconv.Atoi(ata[0])
				tno, _ = strconv.Atoi(ata[1])
				// log.Printf("SW2026-%d: %d/%d\n", i, ano, tno)
				if ano == tno {
					breceived = true
				}
				/*
					if op.Comment != "nil" && ano < tno {
						sendComment(page, op.Comment)
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
				*/
			}
			if op.Received {
				treceived++
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
						log.Printf("Button %d is receivable          ...\n", i)
						breceived = true
						receivable++
						hour := time.Now().Hour()
						if theme.Order != 0 || i != 4 || hour == 2 || hour == 14 {
							log.Printf("Button %d                clicking...\n", i)
							if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
								log.Printf("Error clicking button %d: %v\n", i, err)
							}
						} else {
							treceived--
						}
						if theme.Order == 0 && i == 1 {
							el, err := page.Timeout(15 * time.Second).Element(selector + " button span")
							if err != nil {
								log.Printf("Error finding element for selector %s: %v\n", selector+" button span", err)
							} else {
								text, err := el.Text()
								if err != nil {
									log.Printf("Error getting text for element %s: %v\n", selector+" button span", err)
								} else {
									rno, _ = strconv.Atoi(text)
									log.Printf("SW2026-%d: rno=%d\n", i, rno)
									ano += rno
									if ano < tno {
										breceived = false
									}
								}
							}
						}
						page.MustWaitIdle()

					} else if strings.Contains(*classAttr, "received") {
						breceived = true
						log.Printf("Button %d is received\n", i)
						received++
					} else {
						log.Printf("Button %d is other: %s\n", i, *classAttr)
						others++
					}
				}
			}
			log.Printf("SW2026-%d: %d(+%d)/%d\n", i, ano, rno, tno)

			if !breceived {
				if op.Comment != "nil" {
					sendComment(page, op.Comment)
				}
				if op.Gift > 0 {
					giftlist, err := checkGiftInventory(page)
					if err != nil {
						log.Printf("Error checking gift inventory: %v\n", err)
					} else {
						err = throwGift(page, giftlist[0].Name, fmt.Sprintf("%d", op.Gift))
						if err != nil {
							log.Printf("Error throwing gift: %v\n", err)
						} else {
							log.Printf("Gift %s thrown successfully for mission %d\n", giftlist[0].Name, i)
						}
					}
				}
			}
		}
		log.Printf("receivable: %d, received: %d, others: %d\n",
			receivable, received, others)
	}
	if received == treceived {
		err = fmt.Errorf(cmsg)
		return err
	}
	return
}
