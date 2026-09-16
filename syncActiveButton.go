package main

import (
    "fmt"
    "strings"
    "github.com/go-rod/rod"
)
func SyncActiveButton(page *rod.Page, targetText []string) error {
    // 1. 対象となるボタン群をすべて取得
    buttons := page.MustElements(".st-activate__button")

    nhit := 0
    for _, btn := range buttons {
        sleep(0.4)
        text := btn.MustText()
        isTarget := false
        for _, t := range targetText {
            // if strings.Contains(text, t) {
            if text == t {
                isTarget = true
                nhit++
                break
            }
        }
        hasClassActive := strings.Contains(*btn.MustAttribute("class"), "active")

        if isTarget {
            // ターゲットなら、activeでなければクリック
            if !hasClassActive {
                btn.MustClick()
            }
        } else {
            // ターゲット以外なら、activeならクリック（解除）
            if hasClassActive {
                btn.MustClick()
            }
        }
    }
    ndlg := len(targetText)
    if nhit != ndlg {
        return fmt.Errorf("expected %d target buttons, but found %d", ndlg, nhit)
    }
    return nil
}
