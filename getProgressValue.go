package main
import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-rod/rod"
)

func getProgressValue(button *rod.Element) (int, error) {
	// 1. buttonの次の兄弟要素である div.received-num をXPathで取得
	// "following-sibling::div[1]" は「直後のdiv」を指します
	// targetDiv, err := button.Element("xpath:./following-sibling::div[1]")
	// targetDiv, err := button.Element("+ .received-num")
	// targetDiv, err := button.ElementX("./following-sibling::div[1]")
	// targetDiv, err := button.ElementByJS("this.nextElementSibling")
	targetDiv, err := button.MustParent().Element("button + .received-num")

	if err != nil {
		err = fmt.Errorf("failed to find the sibling div: %w", err)
		return 0, err
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
