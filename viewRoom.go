package main

import (
	"fmt"
	"log"
	// "net/http"
	// "strconv"
	// "strings"
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
		elt.style.left = (rect.left - 400) + 'px';

		// bottom/right が設定されている場合を考慮して解除
		elt.style.right = 'auto';
		elt.style.bottom = 'auto';
		elt.style.margin = '0';
	}`)
	if err != nil {
		log.Printf("Error: failed to move the dialog: %v\n", err)
	}

	// コメントボックスが表示されるまで待つ
	// el, err := page.Timeout(10 * time.Second).Element(".st-gift_box.active.gift-box")
	sleep(1)
	cb, err := page.Element(".st-comment__box")
	if err != nil {
		log.Printf("Error: failed to find the comment box element: %v\n", err)
	}

	// コメントボックスの位置を変える(コメント投稿の邪魔にならないようにする)
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

	err = achieveAndReceiveMission(page, mission)
	if err != nil {
		return fmt.Errorf("achieveAndReceiveMission: %w", err)
	}

	// コメント投稿を行う
	sendComment(page, comment)

	// 視聴時間が経過するまで待機
	time.Sleep(time.Until(ttime))
	log.Printf("viewRoom: waiting finished at %s\n", time.Now().Format("15:04:05"))

	return nil
}
