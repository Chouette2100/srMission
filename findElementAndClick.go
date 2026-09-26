package main

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"time"
)

func findElementAndClick(page *rod.Page, selector string, wait0 float64, wait1 float64) error {

	// 処理前のウェイト
	if wait0 > 0.099 {
		sleep(wait0)
	}

	// セレクターに対するエレメントを取得する
	tgt, err := page.Timeout(10 * time.Second).Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element for selector '%s': %w", selector, err)
	}

	// エレメントが操作可能になるまで待機する
	if _, err = tgt.WaitInteractable(); err != nil {
		return fmt.Errorf("element not interactable for selector %s: %w", selector, err)
	}

	// エレメントをクリックする
	if err := tgt.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click element for selector %s: %w", selector, err)
	}

	// 処理後のウェイト
	if wait1 > 0.099 {
		sleep(wait1)
	}
	return nil
}

func findElementAndGetText(page *rod.Page, selector string, wait0 float64, wait1 float64) (
	txt string, err error,
) {
	// ミッションリストを選択する（「デイリー（昼/夜）」というテキストを含む li 要素を直接指定
	// のつもりだったが、実際やってみると安定性に欠ける、スライダーボタンを使う方が現実的）

	// 処理前のウェイト
	if wait0 > 0.099 {
		sleep(wait0)
	}

	// セレクターに対するエレメントを取得する
	tgt, err := page.Timeout(10 * time.Second).Element(selector)
	if err != nil {
		return "", fmt.Errorf("failed to find element for selector '%s': %w", selector, err)
	}

	// エレメントが操作可能になるまで待機する
	if _, err = tgt.WaitInteractable(); err != nil {
		return "", fmt.Errorf("element not interactable for selector %s: %w", selector, err)
	}

	// エレメントをクリックする
	if txt, err = tgt.Text(); err != nil {
		return "", fmt.Errorf("failed to get text for selector %s: %w", selector, err)
	}

	// 処理後のウェイト
	if wait1 > 0.099 {
		sleep(wait1)
	}
	return txt, nil
}