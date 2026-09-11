package main
import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func getProgressValue(page *rod.Page) (int, error) {
	targetDiv, err := page.Timeout(10 * time.Second).Element(".achieve-section .received-num")
	if err != nil {
		return 0, fmt.Errorf("failed to find the target div: %w", err)
	}
	_, err = targetDiv.WaitInteractable()
	if err != nil {
		return 0, fmt.Errorf("target div not interactable: %w", err)
	}

	// 2. テキストを取得 ("受取済18/20")
	text, err := targetDiv.Text()

	// 3. 正規表現で "/" の前の数値を取り出す
	// "受取済" という文字列を除去し、"/" で分割するアプローチ
	re := regexp.MustCompile(`(\d+)/`)
	matches := re.FindStringSubmatch(text)
	
	if len(matches) < 2 {
		return 0, fmt.Errorf("数値が見つかりませんでした: %s", text)
	}

	matches[0] = strings.Replace(matches[0], "/","",1)

	return strconv.Atoi(matches[0])
}
