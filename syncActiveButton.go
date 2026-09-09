package main

import (
    "strings"
    "github.com/go-rod/rod"
)
func SyncActiveButton(page *rod.Page, targetText string) error {
    // 1. 対象となるボタン群をすべて取得
    buttons := page.MustElements(".st-activate__button")

    for _, btn := range buttons {
        text := btn.MustText()
        isTarget := strings.Contains(text, targetText)
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
    return nil
}
