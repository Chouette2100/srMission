package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const (
	adRewardPageURL     = "https://www.showroom-live.com/lottery/ad_reward"
	remainCountSelector = ".reward-tap-content .reward-define-info > dd:nth-child(2)"
	adCountSelectors    = ".ad-button-box .is-ad > dd:nth-child(2) > em:nth-child(1), .is-ad > dd:nth-child(2) > em:nth-child(1)"
	// rewardButtonSelector = ".reward-tap-content button.reward-score-circle:nth-child(2)"
	rewardButtonSelector = "button.reward-score-circle.border-gradient"
	adButtonSelector     = ".ad-button-box button.reward-score-circle:nth-child(1), .reward-tap-content button.reward-score-circle:nth-child(1)"
	// okButtonSelector     = ".modal.active .modal-actions > button:nth-child(2)"
	okButtonSelector = "div.active.modal-wrapper > section > div.modal-actions > button:nth-child(2)"
	// closeButtonSelector  = ".modal.active .close"
	closeButtonSelector = "div.active.is-result.modal-wrapper > button"
)

func viewReward(
	client *http.Client,
	csrftoken string,
) (
	err error,
) {
	_ = client
	_ = csrftoken

	if srBrowser == nil {
		return fmt.Errorf("browser is not initialized")
	}

	page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("failed to create adreward page: %w", err)
	}
	defer page.Close()

	if err = applyJapaneseLocale(page); err != nil {
		return err
	}
	if err = page.Navigate(adRewardPageURL); err != nil {
		return fmt.Errorf("failed to navigate ad_reward: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait ad_reward page load: %w", err)
	}
	// Vue SPA のレンダリングが落ち着くまで待つ。
	if err = page.WaitIdle(5 * time.Second); err != nil {
		log.Printf("viewReward: WaitIdle returned: %v\n", err)
	}

	adRetryCount := 0
	for {
		time.Sleep(3 * time.Second)

		remainCount, err := readCount(page, remainCountSelector)
		if err != nil {
			captureDebugScreenshot(page, "read-remain-count-failed")
			return fmt.Errorf("failed to read remain draw count: %w", err)
		}
		adCount, err := readCountWithFallback(page, adCountSelectors, 0)
		if err != nil {
			captureDebugScreenshot(page, "read-ad-count-failed")
			return fmt.Errorf("failed to read ad count: %w", err)
		}

		log.Printf("viewReward: remainDraw=%d, adRemaining=%d\n", remainCount, adCount)

		if remainCount == 0 && adCount == 0 {
			log.Printf("viewReward: no remaining rewards and ads, end\n")
			break
		}

		if remainCount > 0 {
			if err = collectReward(page); err != nil {
				captureDebugScreenshot(page, "collect-reward-failed")
				log.Printf("collectReward failed: %v\n", err)
			}
			// ページ側のカウント更新を待つ。
			time.Sleep(2 * time.Second)
			continue
		}

		if adCount > 0 {
			completed, err := watchAd(page)
			if err != nil {
				log.Printf("watchAd failed: %v\n", err)
			}

			if !completed {
				waitSec := adRewardWaitSeconds(adRetryCount)
				log.Printf("viewReward: ad did not complete, wait %d sec before next loop\n", waitSec)
				time.Sleep(time.Duration(waitSec) * time.Second)
				adRetryCount++
				continue
			}

			// 成功時は連続失敗カウントをリセットする。
			adRetryCount = 0
			continue
		}
	}

	return
}

func adRewardWaitSeconds(retryCount int) int {
	const uwait = 20     // 初回の待ち時間（秒）
	const maxwait = 600 // 待ち時間の最大値（秒）
	if retryCount < 0 {
		retryCount = 0
	}
	sec := uwait * (retryCount + 1)
	if sec > maxwait {
		sec = maxwait
	}
	return sec
}

func readCount(page *rod.Page, selector string) (int, error) {
	el, err := page.Timeout(15 * time.Second).Element(selector)
	if err != nil {
		return 0, err
	}
	text, err := el.Text()
	if err != nil {
		return 0, err
	}
	m := regexp.MustCompile(`\d+`).FindString(text)
	if m == "" {
		return 0, fmt.Errorf("no number in text: %q", text)
	}
	count, err := strconv.Atoi(m)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func readCountWithFallback(page *rod.Page, selector string, fallback int) (int, error) {
	el, err := page.Timeout(15 * time.Second).Element(selector)
	if err != nil {
		log.Printf("readCountWithFallback: selector %q not found, using fallback %d: %v\n", selector, fallback, err)
		return fallback, nil
	}
	text, err := el.Text()
	if err != nil {
		return fallback, nil
	}
	m := regexp.MustCompile(`\d+`).FindString(text)
	if m == "" {
		log.Printf("readCountWithFallback: no number in text %q, using fallback %d\n", text, fallback)
		return fallback, nil
	}
	count, err := strconv.Atoi(m)
	if err != nil {
		log.Printf("readCountWithFallback: failed to parse %q, using fallback %d: %v\n", m, fallback, err)
		return fallback, nil
	}
	return count, nil
}

func collectReward(page *rod.Page) error {
	// 報酬取得ボタンが表示・操作可能になるまで待つ。
	button, err := page.Timeout(10 * time.Second).Element(rewardButtonSelector)
	if err != nil {
		return fmt.Errorf("failed to find collect reward button: %w", err)
	}
	if _, err = button.WaitInteractable(); err != nil {
		return fmt.Errorf("collect reward button not interactable: %w", err)
	}
	if err = button.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click collect reward button: %w", err)
	}

	// ダイアログ .active が表示されるまで待つ。
	// if _, err := page.Timeout(10 * time.Second).Element(".modal.active"); err != nil {
	if _, err := page.Timeout(10 * time.Second).Element("div.active.modal-wrapper"); err != nil {
		return fmt.Errorf("failed to wait reward modal active: %w", err)
	}

	okButton, err := page.Timeout(10 * time.Second).Element(okButtonSelector)
	if err != nil {
		return fmt.Errorf("failed to find OK button in reward modal: %w", err)
	}
	if _, err = okButton.WaitInteractable(); err != nil {
		return fmt.Errorf("OK button in reward modal not interactable: %w", err)
	}
	if err = okButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click OK button in reward modal: %w", err)
	}

	// 2 つ目のダイアログ .active が表示されるまで待つ。
	// if _, err := page.Timeout(10 * time.Second).Element(".modal.active"); err != nil {
	if _, err := page.Timeout(10 * time.Second).Element("div.active.is-result.modal-wrapper"); err != nil {
		return fmt.Errorf("failed to wait second reward modal active: %w", err)
	}

	// closeButton, err := page.Timeout(10 * time.Second).Element(closeButtonSelector)
	closeButton, err := page.Timeout(10 * time.Second).Element(closeButtonSelector)
	if err != nil {
		return fmt.Errorf("failed to find close button in reward modal: %w", err)
	}
	if _, err = closeButton.WaitInteractable(); err != nil {
		return fmt.Errorf("close button in reward modal not interactable: %w", err)
	}
	if err = closeButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click close button in reward modal: %w", err)
	}

	return nil
}

func watchAd(parentPage *rod.Page) (completed bool, err error) {
	log.Printf("watchAd: start\n")

	adButton, err := parentPage.Timeout(10 * time.Second).Element(adButtonSelector)
	if err != nil {
		return false, fmt.Errorf("failed to find watch ad button: %w", err)
	}
	if _, err = adButton.WaitInteractable(); err != nil {
		return false, fmt.Errorf("watch ad button not interactable: %w", err)
	}
	if err = adButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return false, fmt.Errorf("failed to click watch ad button: %w", err)
	}

	adPage, err := waitAdPageOpen()
	if err != nil {
		captureDebugScreenshot(parentPage, "ad-page-open-failed")
		return false, fmt.Errorf("failed to wait ad page open: %w", err)
	}

	if err = handleAdMoveDialog(adPage); err != nil {
		captureDebugScreenshot(adPage, "ad-move-dialog-failed")
		_ = adPage.Close()
		return false, fmt.Errorf("failed to handle ad move dialog: %w", err)
	}

	completed, err = waitAdProgressComplete(adPage, 70*time.Second)
	if err != nil {
		captureDebugScreenshot(adPage, "ad-progress-error")
		_ = adPage.Close()
		return false, fmt.Errorf("failed to wait ad progress complete: %w", err)
	}
	if !completed {
		log.Printf("watchAd: ad progress did not complete within timeout\n")
		_ = adPage.Close()
		return false, nil
	}

	// タブが自動で閉じられていた場合は成功とみなす。
	if !isPageAlive(adPage) {
		log.Printf("watchAd: ad page closed after progress complete\n")
		_ = parentPage.WaitLoad()
		return true, nil
	}

	// タブが残っている場合は dismiss ボタンを押下して閉じる。
	dismiss, _, err := findDismissButton(adPage)
	if err != nil {
		captureDebugScreenshot(adPage, "dismiss-search-error")
		log.Printf("watchAd: failed to search dismiss button: %v, closing tab manually\n", err)
		_ = adPage.Close()
		_ = parentPage.WaitLoad()
		return true, nil
	}
	if dismiss == nil {
		captureDebugScreenshot(adPage, "dismiss-not-found")
		log.Printf("watchAd: dismiss button not found, closing tab manually\n")
		_ = adPage.Close()
		_ = parentPage.WaitLoad()
		return true, nil
	}
	if err = dismiss.Click(proto.InputMouseButtonLeft, 1); err != nil {
		captureDebugScreenshot(adPage, "dismiss-click-failed")
		log.Printf("watchAd: failed to click dismiss button, closing tab manually: %v\n", err)
		_ = adPage.Close()
		_ = parentPage.WaitLoad()
		return true, nil
	}

	_ = parentPage.WaitLoad()
	return true, nil
}

func waitAdPageOpen() (*rod.Page, error) {
	const jsURLRegex = `showroom-live\.com/lottery/ad_reward/\d+/watch#goog_rewarded`
	for i := 0; i < 20; i++ {
		pages, err := srBrowser.Pages()
		if err == nil {
			adPage, findErr := pages.FindByURL(jsURLRegex)
			if findErr == nil {
				return adPage, nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("ad watch tab not found")
}

func handleAdMoveDialog(adPage *rod.Page) error {
	hasDialog, _, err := adPage.Has("#dialog-wrapper")
	if err != nil {
		err = fmt.Errorf("failed to check ad move dialog existence: %w", err)
		return err
	}
	if !hasDialog {
		return nil
	}

	moveButton, err := adPage.Timeout(10 * time.Second).Element("#confirmation-buttons")
	if err != nil {
		err = fmt.Errorf("failed to find move button in ad move dialog: %w", err)
		return err
	}
	if err = moveButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		err = fmt.Errorf("failed to click move button in ad move dialog: %w", err)
		return err
	}

	return nil
}

func waitAdProgressComplete(adPage *rod.Page, timeout time.Duration) (bool, error) {
	time.Sleep(2 * time.Second) // プログレスバーが表示されるまでの猶予
	noProgressWaitSec := adRewardNoProgressWaitSeconds()
	searchDeadline := time.Now().Add(10 * time.Second)
	overallDeadline := time.Now().Add(timeout)

	// プログレスバーが表示されるまで最大 10 秒間ポーリング（iframe 内も検索）
	for time.Now().Before(searchDeadline) {
		// ページが広告配信失敗などで自動的に閉じられた場合は完了とみなす。
		if !isPageAlive(adPage) {
			log.Printf("waitAdProgressComplete: ad page closed or navigation failed\n")
			return true, nil
		}

		progressEl, _, err := findProgressBar(adPage)
		if err != nil {
			return false, fmt.Errorf("failed to find progress bar: %w", err)
		}
		if progressEl != nil {
			return monitorProgressBar(progressEl, overallDeadline)
		}

		time.Sleep(1 * time.Second)
	}

	// 10 秒経ってもプログレスバーが見つからない場合は「プログレスバーなし」として待機
	log.Printf("waitAdProgressComplete: progress bar not found after initial search\n")
	diagnoseProgressBar(adPage)
	captureDebugScreenshot(adPage, "no-progress-bar")
	dumpPageHTML(adPage, "no-progress-bar")
	log.Printf("waitAdProgressComplete: progress bar not found, wait %d sec and treat as completed\n", noProgressWaitSec)
	time.Sleep(time.Duration(noProgressWaitSec) * time.Second)
	return true, nil
}

func findProgressBar(adPage *rod.Page) (*rod.Element, *rod.Page, error) {
	el, frame, err := findInFrames(adPage, "#progress-bar")
	if err != nil {
		return nil, nil, err
	}
	if el != nil && frame != nil {
		log.Printf("findProgressBar: found #progress-bar in iframe\n")
	}
	return el, frame, nil
}

func findDismissButton(adPage *rod.Page) (*rod.Element, *rod.Page, error) {
	return findInFrames(adPage, "#dismiss-button-element")
}

func findInFrames(adPage *rod.Page, selector string) (*rod.Element, *rod.Page, error) {
	// メインドキュメント
	if has, el, _ := adPage.Has(selector); has {
		return el, adPage, nil
	}

	// 各 iframe 内を検索
	iframes, err := adPage.Elements("iframe")
	if err != nil {
		return nil, nil, err
	}
	for i, iframe := range iframes {
		frame, err := iframe.Frame()
		if err != nil {
			log.Printf("findInFrames: failed to get frame for iframe[%d]: %v\n", i, err)
			continue
		}
		if frame == nil {
			continue
		}
		if has, el, _ := frame.Has(selector); has {
			log.Printf("findInFrames: found %q in iframe[%d]\n", selector, i)
			return el, frame, nil
		}
	}

	return nil, nil, nil
}

func monitorProgressBar(progressEl *rod.Element, deadline time.Time) (bool, error) {
	log.Printf("waitAdProgressComplete: progress bar found, monitoring until 100%%\n")
	for time.Now().Before(deadline) {
		widthObj, err := progressEl.Eval(`function() {
			const inner = this.querySelector('#progress-bar-inner');
			if (inner) {
				return inner.style.width || window.getComputedStyle(inner).width || '';
			}
			return this.style.width || window.getComputedStyle(this).width || '';
		}`)
		if err != nil {
			log.Printf("waitAdProgressComplete: failed to get width: %v\n", err)
		} else {
			width := widthObj.Value.String()
			if p := parsePercent(width); p >= 100 {
				log.Printf("waitAdProgressComplete: progress reached 100%%\n")
				return true, nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return false, nil
}

func adRewardNoProgressWaitSeconds() int {
	const defaultWaitSec = 50
	v := os.Getenv("SR_ADREWARD_NO_PROGRESS_WAIT_SEC")
	if v == "" {
		return defaultWaitSec
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		log.Printf("invalid SR_ADREWARD_NO_PROGRESS_WAIT_SEC=%q, fallback to %d\n", v, defaultWaitSec)
		return defaultWaitSec
	}
	return n
}

func isPageAlive(page *rod.Page) bool {
	if page == nil {
		return false
	}
	_, err := page.Eval(`() => location.href`)
	return err == nil
}

func captureDebugScreenshot(page *rod.Page, name string) {
	if !isDebugScreenshotEnabled() {
		return
	}
	if page == nil {
		return
	}

	dir := debugScreenshotDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("captureDebugScreenshot: failed to create dir: %v\n", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405.000")
	filename := filepath.Join(dir, fmt.Sprintf("%s_%s.png", timestamp, name))
	img, err := page.Screenshot(true, nil)
	if err != nil {
		log.Printf("captureDebugScreenshot: failed to take screenshot: %v\n", err)
		return
	}
	if err := os.WriteFile(filename, img, 0644); err != nil {
		log.Printf("captureDebugScreenshot: failed to write screenshot: %v\n", err)
		return
	}
	log.Printf("captureDebugScreenshot: saved %s\n", filename)
}

func isDebugScreenshotEnabled() bool {
	switch os.Getenv("SR_ADREWARD_DEBUG_SCREENSHOTS") {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

func debugScreenshotDir() string {
	if v := os.Getenv("SR_ADREWARD_DEBUG_SCREENSHOT_DIR"); v != "" {
		return v
	}
	return "screenshots"
}

func parsePercent(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}
	m := regexp.MustCompile(`\d+`).FindString(s)
	if m == "" {
		return -1
	}
	v, err := strconv.Atoi(m)
	if err != nil {
		return -1
	}
	return v
}

func dumpPageHTML(page *rod.Page, name string) {
	if !isDebugHTMLDumpEnabled() {
		return
	}
	if page == nil {
		return
	}
	html, err := page.Eval(`() => document.documentElement.outerHTML`)
	if err != nil {
		log.Printf("dumpPageHTML: failed to get HTML: %v\n", err)
		return
	}
	filename := fmt.Sprintf("debug_%s_%s.html", time.Now().Format("20060102_150405"), name)
	if err := os.WriteFile(filename, []byte(html.Value.String()), 0644); err != nil {
		log.Printf("dumpPageHTML: failed to write HTML: %v\n", err)
		return
	}
	log.Printf("dumpPageHTML: saved %s (%d bytes)\n", filename, len(html.Value.String()))
}

func isDebugHTMLDumpEnabled() bool {
	return isTruthyEnv(os.Getenv("SR_ADREWARD_DEBUG_HTML"))
}

func diagnoseProgressBar(adPage *rod.Page) {
	log.Printf("=== progress bar diagnosis start ===")

	// 1. document 全体から progress 関連要素を検索
	result, err := adPage.Eval(`() => {
		const found = [];
		const keywords = ['progress-bar', 'progress', 'dismiss-button'];
		const getClass = (node) => {
			if (!node.className) return '';
			if (typeof node.className === 'string') return node.className;
			if (node.className.baseVal !== undefined) return node.className.baseVal;
			return String(node.className);
		};
		const walk = (node, depth) => {
			if (!node || depth > 20) return;
			if (node.nodeType === 1) { // Element
				const cls = getClass(node);
				const id = node.id || '';
				const tag = node.tagName || '';
				const match = keywords.some(k => id.includes(k) || cls.includes(k) || tag.toLowerCase().includes(k));
				if (match) {
					found.push({
						tag: tag,
						id: id,
						class: cls,
						depth: depth,
						width: node.style ? node.style.width : '',
						computedWidth: window.getComputedStyle(node).width,
						parentTag: node.parentElement ? node.parentElement.tagName : '',
						parentId: node.parentElement ? node.parentElement.id : '',
					});
				}
				// shadow DOM も掘る
				if (node.shadowRoot) {
					walk(node.shadowRoot, depth + 1);
				}
			}
			let child = node.firstChild;
			while (child) {
				walk(child, depth + 1);
				child = child.nextSibling;
			}
		};
		walk(document.documentElement, 0);
		return JSON.stringify(found, null, 2);
	}`)
	if err != nil {
		log.Printf("diagnoseProgressBar: eval error: %v\n", err)
	} else {
		log.Printf("diagnoseProgressBar: found elements:\n%s\n", result.Value.String())
	}

	// 2. iframe の確認
	iframes, err := adPage.Elements("iframe")
	if err != nil {
		log.Printf("diagnoseProgressBar: iframe check error: %v\n", err)
	} else {
		log.Printf("diagnoseProgressBar: number of iframes: %d\n", len(iframes))
		for i, iframe := range iframes {
			src, _ := iframe.Attribute("src")
			id, _ := iframe.Attribute("id")

			idStr := ""
			if id != nil {
				idStr = *id
			}
			srcStr := ""
			if src != nil {
				srcStr = *src
			}
			log.Printf("diagnoseProgressBar: iframe[%d] id=%q src=%q\n", i, idStr, srcStr)
		}
	}

	// 3. document.documentElement.outerHTML の長さを確認
	htmlLen, err := adPage.Eval(`() => document.documentElement.outerHTML.length`)
	if err != nil {
		log.Printf("diagnoseProgressBar: html length error: %v\n", err)
	} else {
		log.Printf("diagnoseProgressBar: document HTML length: %s\n", htmlLen.Value.String())
	}

	log.Printf("=== progress bar diagnosis end ===")
}
