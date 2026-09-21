package main

import (
	"fmt"
	"log"
	// "net/http"
	// "strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	// "github.com/Chouette2100/srapi/v2"
)

const cmsg = "viewRoom: mission completed for room"

func viewRoom(
	page *rod.Page,
	// client *http.Client,
	// csrftoken string,
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

	if err = page.Navigate("https://www.showroom-live.com/r/" + room.URL); err != nil {
		return fmt.Errorf("failed to navigate room: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait room page load: %w", err)
	}

	// 視聴完了時刻を求める
	ttime := time.Now().Add(time.Duration(viewingTime) * time.Second)
	log.Printf("viewRoom: waiting until %s (viewingTime=%d seconds)\n", ttime.Format("15:04:05"), viewingTime)

	// 告知があれば閉じる
	sleep(0.5)
	rcc, err := page.Timeout(10 * time.Second).Element(
		".room-campaign-close")
	if err == nil {

		if _, err = rcc.WaitInteractable(); err != nil {
			return fmt.Errorf("room-campaign-close button not interactable: %w", err)
		}

		if err := rcc.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("failed to click the room-campaign-close button: %w", err)
		}
		page.MustWaitIdle() // 描画が落ち着くのを待つ
	}
	// 告知のモーダルダイアログを閉じる
	sleep(1)
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

	// 視聴完了時刻を求める
	// ttime := time.Now().Add(time.Duration(viewingTime) * time.Second)
	// log.Printf("viewRoom: waiting until %s (viewingTime=%d seconds)\n", ttime.Format("15:04:05"), viewingTime)

	// 一番下までスクロールして、トグルボタンを見えるようにする
	// page.MustEval(`window.scrollTo(0, document.body.scrollHeight)`)
	sleep(1)
	page.Mouse.Scroll(0, 100000, 20)
	page.MustWaitIdle() // 描画が落ち着くのを待つ

	// "ミッション"のダイアログのみ表示するために、トグルボタンの状態を同期する
	err = SyncActiveButton(page, []string{"ミッション", "コメント", "ギフト"})
	if err != nil {
		return fmt.Errorf("SyncActiveButton: %w", err)
	}
	// 一番上までスクロールして、"ミッション"ボタンを表示させる
	// page.MustEval(`window.scrollTo(0, 0)`)
	sleep(1)
	page.Mouse.Scroll(0, -100000, 20)
	page.MustWaitIdle()

	// ギフトボックスが表示されるまで待つ
	// el, err := page.Timeout(10 * time.Second).Element(".st-gift_box.active.gift-box")
	sleep(1)
	el, err := page.Element(".st-gift_box.active.gift-box, .st-fan__status")
	if err != nil {
		log.Printf("Error: failed to find the gift box element: %v\n", err)
	}

	// ギフトボックスの位置を変える(コメント投稿の邪魔にならないようにする)
	_, err = el.Eval(`(el) => {
		const elt = this;
		const rect = elt.getBoundingClientRect();

		// position: fixed にして画面座標ベースで配置
		elt.style.position = 'fixed';
		elt.style.top = (rect.top - 400) + 'px';
		elt.style.left = rect.left + 'px';

		// bottom/right が設定されている場合を考慮して解除
		elt.style.right = 'auto';
		elt.style.bottom = 'auto';
		elt.style.margin = '0';
	}`)
	if err != nil {
		log.Printf("Error: failed to move the dialog: %v\n", err)
	}

	// ギフトボックスが表示されるまで待つ
	// el, err := page.Timeout(10 * time.Second).Element(".st-gift_box.active.gift-box")
	sleep(1)
	cb, err := page.Element(".st-comment__box")
	if err != nil {
		log.Printf("Error: failed to find the comment box element: %v\n", err)
	}

	// ギフトボックスの位置を変える(コメント投稿の邪魔にならないようにする)
	_, err = cb.Eval(`(cb) => {
		const rect = this.getBoundingClientRect();

		// position: fixed にして画面座標ベースで配置
		this.style.position = 'fixed';
		this.style.top = (rect.top - 200) + 'px';
		this.style.left = rect.left + 'px';

		// bottom/right が設定されている場合を考慮して解除
		this.style.right = 'auto';
		this.style.bottom = 'auto';
		this.style.margin = '0';
	}`)
	if err != nil {
		log.Printf("Error: failed to move the comment box: %v\n", err)
	}

	switch mission {
	case "daily":
		// err = checkReceivedDaily(page)
		err = achieveAndReceiveMission(page, "Daily")
		if err != nil {
			return fmt.Errorf("checkReceivedDaily: %w", err)
		}
	case "discovery":
		// err = checkReceivedDiscovery(page)
		err = achieveAndReceiveMission(page, "SW2026-Sep.")
		if err != nil {
			return fmt.Errorf("checkReceivedDiscovery: %w", err)
		}
	case "newcommer":
		err = achieveAndReceiveMission(page, "SW2026-NewCommer")
		if err != nil {
			return fmt.Errorf("achieveAndReceiveMission: %w", err)
		}
	case "newcommer-old":

		// 「新人ライバー応援キャンペーン」というテキストを含む li 要素を直接指定
		sleep(1)
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
		for i := 0; i < len(buttons); i++ {
			sleep(1)
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

	// コメント投稿を行う
	sendComment(page, comment)

	// 視聴時間が経過するまで待機
	time.Sleep(time.Until(ttime))
	log.Printf("viewRoom: waiting finished at %s\n", time.Now().Format("15:04:05"))

	return nil
}
