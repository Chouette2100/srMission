package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
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
	if err = page.Navigate("https://www.showroom-live.com/lottery/ad_reward"); err != nil {
		return fmt.Errorf("failed to navigate ad_reward: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait ad_reward page load: %w", err)
	}
	if err = page.Reload(); err != nil {
		return fmt.Errorf("failed to reload ad_reward page: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait reloaded ad_reward page load: %w", err)
	}

	adRetryCount := 0
	for {
		time.Sleep(5 * time.Second)
		remainCount, err := readCount(page, ".reward-define-info > dd:nth-child(2)")
		if err != nil {
			return fmt.Errorf("failed to read remain draw count: %w", err)
		}
		adCount, err := readCount(page, ".is-ad > dd:nth-child(2) > em:nth-child(1)")
		if err != nil {
			return fmt.Errorf("failed to read ad count: %w", err)
		}

		log.Printf("viewReward: remainDraw=%d, adRemaining=%d\n", remainCount, adCount)

		if remainCount == 0 && adCount == 0 {
			log.Printf("viewReward: no remaining rewards and ads, end\n")
			break
		}

		if remainCount > 0 {
			if err = collectReward(page); err != nil {
				err = fmt.Errorf("collectReward: %w", err)
				return err
			}
			// ページ側のカウント更新を待つ。
			time.Sleep(2 * time.Second)
			continue
		}

		if adCount > 0 {
			retryNeeded, err := watchAd(page)
			if err != nil {
				return fmt.Errorf("watchAd: %w", err)
			}

			if retryNeeded {
				waitSec := adRewardWaitSeconds(adRetryCount)
				log.Printf("viewReward: ad tab closed or incomplete, retry after %d sec\n", waitSec)
				time.Sleep(time.Duration(waitSec) * time.Second)
				adRetryCount++
				continue
			}

			// Successful completion without fallback wait resets the retry-based wait duration.
			adRetryCount = 0
			continue
		}
	}

	return
}

func adRewardWaitSeconds(retryCount int) int {
	const uwait = 5     // 初回の待ち時間（秒）
	const maxwait = 100 // 待ち時間の最大値（秒）
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

func collectReward(page *rod.Page) error {
	const rewardButtonSelector = "button.reward-score-circle:nth-child(2)"
	// Element() で要素を待つ。タイムアウトは独立させる。
	if _, err := page.Timeout(10 * time.Second).Element(rewardButtonSelector); err != nil {
		return fmt.Errorf("failed to find collect reward button: %w", err)
	}
	// 新しいタイムアウトで要素を取得し直し、WaitInteractable() + Click() に十分な時間を確保する。
	button, err := page.Timeout(10 * time.Second).Element(rewardButtonSelector)
	if err != nil {
		return fmt.Errorf("failed to get collect reward button: %w", err)
	}
	if _, err = button.WaitInteractable(); err != nil {
		return fmt.Errorf("collect reward button not interactable: %w", err)
	}
	if err = button.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click collect reward button: %w", err)
	}

	const okButtonSelector = ".modal-actions > button:nth-child(2)"
	if _, err := page.Timeout(10 * time.Second).Element(okButtonSelector); err != nil {
		return fmt.Errorf("failed to find OK button in reward modal: %w", err)
	}
	okButton, err := page.Timeout(10 * time.Second).Element(okButtonSelector)
	if err != nil {
		return fmt.Errorf("failed to get OK button in reward modal: %w", err)
	}
	if _, err = okButton.WaitInteractable(); err != nil {
		log.Printf("collectReward: OK button not interactable, continuing: %v\n", err)
		return nil
	}
	if err = okButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		log.Printf("collectReward: failed to click OK button, continuing: %v\n", err)
		return nil
	}

	const closeButtonSelector = ".close"
	if _, err := page.Timeout(10 * time.Second).Element(closeButtonSelector); err != nil {
		return fmt.Errorf("failed to find close button in reward modal: %w", err)
	}
	closeButton, err := page.Timeout(10 * time.Second).Element(closeButtonSelector)
	if err != nil {
		return fmt.Errorf("failed to get close button in reward modal: %w", err)
	}
	if _, err = closeButton.WaitInteractable(); err != nil {
		log.Printf("collectReward: close button not interactable, continuing: %v\n", err)
		return nil
	}
	if err = closeButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		log.Printf("collectReward: failed to click close button, continuing: %v\n", err)
		return nil
	}

	return nil
}

func watchAd(parentPage *rod.Page) (retryNeeded bool, err error) {
	const adButtonSelector = "button.reward-score-circle:nth-child(1)"

	log.Printf("watchAd: Next - Element(adButtonSelector)\n")
	// Element() で要素を待つ。タイムアウトは独立させる。
	if _, err = parentPage.Timeout(10 * time.Second).Element(adButtonSelector); err != nil {
		return false, fmt.Errorf("failed to find watch ad button: %w", err)
	}
	// 新しいタイムアウトで要素を取得し直し、WaitInteractable() + Click() に十分な時間を確保する。
	adButton, err := parentPage.Timeout(10 * time.Second).Element(adButtonSelector)
	if err != nil {
		return false, fmt.Errorf("failed to get watch ad button: %w", err)
	}
	if _, err = adButton.WaitInteractable(); err != nil {
		return false, fmt.Errorf("watch ad button not interactable: %w", err)
	}
	if err = adButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return false, fmt.Errorf("failed to click watch ad button: %w", err)
	}

	adPage, err := waitAdPageOpen()
	if err != nil {
		log.Printf("watchAd: failed to wait ad page open: %v\n", err)
		return true, nil
	}
	defer adPage.Close()

	if err = handleAdMoveDialog(adPage); err != nil {
		log.Printf("watchAd: failed to handle ad move dialog: %v\n", err)
		return true, nil
	}

	completed, err := waitAdProgressComplete(adPage, 70*time.Second)
	if err != nil {
		log.Printf("watchAd: failed to wait ad progress complete: %v\n", err)
		return true, nil
	}
	if !completed {
		log.Printf("watchAd: ad progress did not complete within timeout\n")
		return true, nil
	}

	// 3.B.i: ウェイト後にタブが自動で閉じられていた場合は成功とみなす。
	if _, aliveErr := adPage.Eval(`() => location.href`); aliveErr != nil {
		log.Printf("watchAd: ad page closed after wait, treating as completed\n")
		_ = parentPage.WaitLoad()
		return false, nil
	}

	// 3.A / 3.B.ii: タブが残っている場合は dismiss ボタンを押下して閉じる。
	dismiss, err := adPage.Timeout(10 * time.Second).Element("#dismiss-button-element")
	if err != nil {
		log.Printf("watchAd: failed to find dismiss button: %v\n", err)
		return true, nil
	}
	if err = dismiss.Click(proto.InputMouseButtonLeft, 1); err != nil {
		log.Printf("watchAd: failed to click dismiss button: %v\n", err)
		return true, nil
	}

	_ = parentPage.WaitLoad()
	return false, nil
}

func waitAdPageOpen() (*rod.Page, error) {
	const jsURLRegex = `showroom-live\\.com/lottery/ad_reward/.*/watch#goog_rewarded`
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
	time.Sleep(2 * time.Second) // Wait for the progress bar to appear
	noProgressWaitSec := adRewardNoProgressWaitSeconds()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Page can close itself when ad serving fails; treat it as completed (3.B.i).
		urlObj, evalErr := adPage.Eval(`() => location.href`)
		if evalErr != nil {
			log.Printf("waitAdProgressComplete: ad page closed or navigation failed: %v\n", evalErr)
			return true, nil
		}
		_ = urlObj

		hasProgress, progressEl, err := adPage.Has("#progress-bar")
		if err != nil {
			err = fmt.Errorf("failed to check progress bar existence: %w", err)
			return false, err
		}
		if !hasProgress {
			log.Printf("waitAdProgressComplete: progress bar not found, wait %d sec and treat as completed\n", noProgressWaitSec)
			time.Sleep(time.Duration(noProgressWaitSec) * time.Second)
			return true, nil
		}
		if hasProgress {
			text, textErr := progressEl.Text()
			if textErr == nil {
				if p := parsePercent(text); p >= 100 {
					return true, nil
				}
			}

			valueObj, valueErr := adPage.Eval(`() => {
				const el = document.querySelector('#progress-bar');
				if (!el) return '';
				return el.getAttribute('aria-valuenow') || '';
			}`)
			if valueErr == nil {
				if p := parsePercent(valueObj.Value.String()); p >= 100 {
					return true, nil
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

	return false, nil
}

func adRewardNoProgressWaitSeconds() int {
	const defaultWaitSec = 30
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
