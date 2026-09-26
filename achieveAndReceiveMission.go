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

const (
	progressModeNone        = "none"
	progressModeAchieveText = "achieve_text"
	progressModeReceivedNum = "received_num"

	rewardModeNone        = "none"
	rewardModeOnce        = "once"
	rewardModeIncremental = "incremental"

	receivePolicyImmediate        = "immediate"
	receivePolicyManualUntilHours = "manual_until_hours"

	commentModeNone     = "none"
	commentModeThanks39 = "thanks39"
	commentModeGreeting = "greeting"
	commentModeText     = "text"
)

func missionProgressSelector(mission Mission) (string, error) {
	switch mission.ProgressMode {
	case progressModeNone:
		return "", nil
	case progressModeAchieveText:
		return mission.Selector + "  .achieve-text", nil
	case progressModeReceivedNum:
		return mission.Selector + "  .received-num", nil
	default:
		return "", fmt.Errorf("unknown progress mode: %s", mission.ProgressMode)
	}
}

func parseMissionProgress(mode string, raw string) (achieved int, total int, err error) {
	text := strings.TrimSpace(raw)
	if mode == progressModeReceivedNum {
		text = strings.TrimSpace(strings.TrimPrefix(text, "受取済"))
	}
	parts := strings.Split(text, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid progress text: %q", raw)
	}
	achieved, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid achieved progress %q: %w", parts[0], err)
	}
	total, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid total progress %q: %w", parts[1], err)
	}
	return achieved, total, nil
}

func shouldCheckMissionState(mission Mission) bool {
	return mission.ProgressMode != progressModeNone || mission.RewardMode != rewardModeNone
}

func shouldReceiveMissionReward(mission Mission, now time.Time) bool {
	switch mission.ReceivePolicy {
	case receivePolicyImmediate:
		return true
	case receivePolicyManualUntilHours:
		for _, hour := range mission.ReceiveHours {
			if now.Hour() == hour {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func missionCommentText(mission Mission) string {
	switch mission.CommentMode {
	case commentModeNone:
		return "nil"
	case commentModeThanks39:
		return "39"
	case commentModeGreeting:
		return "j"
	case commentModeText:
		return mission.CommentText
	default:
		return "nil"
	}
}

func achieveAndReceiveMission(page *rod.Page, themaID string) (err error) {

	// theme.ID == "Daily" && op.ID == "completed" の場合にtrueにする。
	// これがtrueの場合はDailyミッションは実質終了している(報酬は受け取っていない場合もある)
	bend := false

	theme, ok := themaList[themaID]
	if !ok {
		return fmt.Errorf("unknown theme id: %s", themaID)
	}
	// page.MustWaitIdle()
	selectMission(page, theme.Name)
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
	selector := ""
	i := 0
	for ; i < len(theme.Missions); i++ {
		// sleep(1.0)
		selector = theme.Missions[i].Selector
		op := theme.Missions[i]
		// log.Printf("%s-%d(%s): selector=%s, op=%+v\n", theme.ID, i, op.ID, selector, op)
		breceived := false
		tno := 0
		ano := 0
		rno := 0
		if shouldCheckMissionState(op) {
			// 進捗を確認する必要がある場合
			sleep(0.5)
			if op.ProgressMode != progressModeNone {
				// 進捗が分割されているミッション
				selectorTxt, err := missionProgressSelector(op)
				if err != nil {
					return fmt.Errorf("failed to get progress selector for %s/%s: %w", theme.ID, op.ID, err)
				}
				tgt, err := page.Timeout(10 * time.Second).Element(selectorTxt)
				if err != nil {
					return fmt.Errorf("failed to Elements(selector) for tgt %d[%s]: %w", i, selectorTxt, err)
				}
				if _, err = tgt.WaitInteractable(); err != nil {
					return fmt.Errorf("button not interactable: %w", err)
				}
				at, err := tgt.Timeout(10 * time.Second).Text()
				if err != nil {
					return fmt.Errorf("failed to get text for tgt %d: %w", i, err)
				}
				ano, tno, err = parseMissionProgress(op.ProgressMode, at)
				if err != nil {
					return fmt.Errorf("failed to parse progress for %s/%s: %w", theme.ID, op.ID, err)
				}
				// log.Printf("SW2026-%d: %d/%d\n", i, ano, tno)
				if ano == tno {
					breceived = true
				}
			}
			if op.RewardMode != rewardModeNone {
				// sleep(1.0)
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
						// 受取ボタン、ボタンを押せば報酬を受け取れる
						log.Printf("Button %d is receivable          ...\n", i)
						receivable++
						if shouldReceiveMissionReward(op, time.Now()) {
							breceived = true
							log.Printf("Button %d                clicking...\n", i)
							if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
								log.Printf("Error clicking button %d: %v\n", i, err)
							}
							if theme.ID == "Daily" && op.ID == "completed" {
								// この報酬を受け取れた場合はViews20を含む全てのミッションは必ず終了している。
								// ただしこのことは終了したミッションの報酬を受け取ったという意味ではない
								bend = true
								log.Printf("bend = true: Daily mission completed, all missions are considered completed\n")
							}

						} else {
							log.Printf("Button %d receivable but kept pending by policy\n", i)
							treceived--
						}
						if op.RewardMode == rewardModeIncremental {
							// 進捗の過程で報酬を受け取れるケース
							// DailyのView20がこれに相当するが、ここを通ると視聴ボーナスの受取に１分間かかる！
							// sleep(1.0)
							el, err := page.Timeout(15 * time.Second).Element(selector + " button span")
							if err != nil {
								log.Printf("Error finding element for selector %s: %v\n", selector+" button span", err)
							} else {
								text, err := el.Text()
								if err != nil {
									log.Printf("Error getting text for element %s: %v\n", selector+" button span", err)
								} else {
									rno, _ = strconv.Atoi(text)
									log.Printf("%s-%d(%s): rno=%d\n", theme.ID, i, op.ID, rno)
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
			log.Printf("%s-%d(%s): %d(+%d)/%d\n", theme.ID, i, op.ID, ano, rno, tno)

			if !breceived {
				comment := missionCommentText(op)
				if comment != "nil" {
					// sleep(1.0)
					sendComment(page, comment)
				}
				if op.Gift > 0 {
					// sleep(1.0)
					if op.GiftType != "StarsAndSeeds" {
							err = throwGift(page, op.GiftType, fmt.Sprintf("%d", op.Gift))
							if err != nil {
								log.Printf("Error throwing gift: %v\n", err)
							} else {
								log.Printf("Gift %s thrown successfully for mission %d\n", 
									op.GiftType, i)
							}
					} else {
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
		}
		log.Printf("receivable: %d, received: %d, others: %d\n",
			receivable, received, others)
	}
	if received == treceived || bend {
		if bend {
			// view20の報酬を受け取る
			err = findElementAndClick(page, theme.Missions[1].Selector+" .achieve-button", 1.0, 0.5)
			if err != nil {
				log.Printf("Error clicking view20 achieve-button: %v\n", err)
				// エラーが起きてもミッションは終了しているのでそのまま終了してかまわない
			} else {
				log.Printf("Clicked view20 achieve-button successfully\n")
			}
		}
		err = fmt.Errorf(cmsg)
		return err
	}
	return
}
