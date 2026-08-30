package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
)

var srBrowser *rod.Browser
var srLauncher *launcher.Launcher

func applyJapaneseLocale(page *rod.Page) error {
	if _, err := page.SetExtraHeaders([]string{
		"Accept-Language", "ja-JP,ja;q=0.9,en-US;q=0.8,en;q=0.7",
	}); err != nil {
		return fmt.Errorf("failed to set extra headers: %w", err)
	}

	uaObj, err := page.Eval(`() => navigator.userAgent`)
	if err != nil {
		return fmt.Errorf("failed to get user agent: %w", err)
	}
	ua := uaObj.Value.String()
	if ua != "" {
		if err = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent:      ua,
			AcceptLanguage: "ja-JP,ja",
		}); err != nil {
			return fmt.Errorf("failed to set user agent override: %w", err)
		}
	}

	if err = (proto.EmulationSetLocaleOverride{Locale: "ja_JP"}).Call(page); err != nil {
		return fmt.Errorf("failed to set locale override: %w", err)
	}
	if err = (proto.EmulationSetTimezoneOverride{TimezoneID: "Asia/Tokyo"}).Call(page); err != nil {
		return fmt.Errorf("failed to set timezone override: %w", err)
	}

	return nil
}

func findBrowserBin() (string, error) {
	// SR_BROWSER_BIN is optional.
	// In some environments (e.g. NixOS), explicitly pinning a browser path can behave differently
	// from auto-detected binaries, so prefer leaving it unset unless you need to force a specific one.
	if v := os.Getenv("SR_BROWSER_BIN"); v != "" {
		if p, err := exec.LookPath(v); err == nil {
			return p, nil
		}
		// If SR_BROWSER_BIN is an absolute path, keep the value for a clearer launch error.
		return v, nil
	}

	candidates := []string{
		"chromium",
		"chromium-browser",
		"google-chrome-stable",
		"google-chrome",
		"chrome",
	}

	for _, name := range candidates {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("browser binary not found; set SR_BROWSER_BIN or install chromium/google-chrome")
}

func closeBrowser() {
	if srBrowser != nil {
		if err := srBrowser.Close(); err != nil {
			log.Printf("closeBrowser: %v\n", err)
		}
		srBrowser = nil
	}
	if srLauncher != nil {
		// Remove temporary user-data-dir to avoid leaving cookies on disk.
		srLauncher.Cleanup()
		srLauncher = nil
	}
}

func ensureBrowser() (err error) {
	if srBrowser != nil {
		return nil
	}

	// Browser binary selection policy:
	// 1) use SR_BROWSER_BIN when explicitly set
	// 2) otherwise auto-detect common names (chromium/google-chrome)
	// If startup becomes unstable, first try unsetting SR_BROWSER_BIN.
	browserBin, err := findBrowserBin()
	if err != nil {
		return err
	}

	launcherEnv := append(os.Environ(),
		"LANG=ja_JP.UTF-8",
		"LC_ALL=ja_JP.UTF-8",
		"LANGUAGE=ja_JP:ja",
		"TZ=Asia/Tokyo",
	)

	headless := false
	if os.Getenv("SR_HEADLESS") == "1" || os.Getenv("SR_HEADLESS") == "true" {
		headless = true
	}

	l := launcher.New().
		Headless(headless).
		Bin(browserBin).
		Env(launcherEnv...).
		Set(flags.Flag("lang"), "ja-JP").
		Set(flags.Flag("window-size"), "1360,1020").
		Set(flags.Flag("force-device-scale-factor"), "1")

	// WebGL tuning is opt-in. Some combinations can crash on specific host drivers.
	// SR_WEBGL_MODE: off(default) | hardware | swiftshader
	switch os.Getenv("SR_WEBGL_MODE") {
	case "hardware":
		l = l.
			Set(flags.Flag("enable-webgl")).
			Set(flags.Flag("ignore-gpu-blocklist")).
			Set(flags.Flag("enable-accelerated-2d-canvas")).
			Set(flags.Flag("use-angle"), "gl").
			Set(flags.Flag("use-gl"), "desktop")
	case "swiftshader":
		l = l.
			Set(flags.Flag("enable-webgl")).
			Set(flags.Flag("ignore-gpu-blocklist")).
			Set(flags.Flag("enable-accelerated-2d-canvas")).
			Set(flags.Flag("use-angle"), "swiftshader").
			Set(flags.Flag("use-gl"), "angle")
	}

	// 画面内の表示言語も日本語寄りにする
	l = l.Set(flags.Preferences, `{"intl":{"accept_languages":"ja-JP,ja"}}`)

	controlURL, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser with %s: %w", browserBin, err)
	}
	browser := rod.New().ControlURL(controlURL)
	if err = browser.Connect(); err != nil {
		l.Cleanup()
		return fmt.Errorf("failed to connect browser: %w", err)
	}
	srBrowser = browser
	srLauncher = l
	return nil
}

func srLogin(
	acct string,
	pswd string,
) (
	err error,
) {

	log.Printf("srLogin: acct=%s, pswd=***\n", acct)
	if err = ensureBrowser(); err != nil {
		return err
	}

	page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()
	if err = applyJapaneseLocale(page); err != nil {
		return err
	}

	if err = page.Navigate("https://www.showroom-live.com/"); err != nil {
		return fmt.Errorf("failed to navigate showroom: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait page load: %w", err)
	}

	const myPageSelector = `#js-wrapper > div.pc-header > a:nth-child(6)`
	if _, myPageErr := page.Timeout(3 * time.Second).Element(myPageSelector); myPageErr == nil {
		log.Printf("srLogin: already logged in\n")
		return nil
	}

	const loginButtonSelector = `#js-wrapper > div.pc-header > button`
	loginButton, findLoginButtonErr := page.Timeout(10 * time.Second).Element(loginButtonSelector)
	if findLoginButtonErr != nil {
		return fmt.Errorf("failed to find login button: %w", findLoginButtonErr)
	}
	if err = loginButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click login button: %w", err)
	}

	const acctInputSelector = `#js-login-form > div:nth-child(2) > div:nth-child(1) > input`
	const pswdInputSelector = `#js-login-form > div:nth-child(2) > div:nth-child(2) > input`
	const submitButtonSelector = `#js-login-form > div:nth-child(3)`

	acctInput, acctErr := page.Timeout(10 * time.Second).Element(acctInputSelector)
	if acctErr != nil {
		return fmt.Errorf("failed to find account input: %w", acctErr)
	}
	pswdInput, pswdErr := page.Timeout(10 * time.Second).Element(pswdInputSelector)
	if pswdErr != nil {
		return fmt.Errorf("failed to find password input: %w", pswdErr)
	}

	if err = acctInput.Input(acct); err != nil {
		return fmt.Errorf("failed to input account: %w", err)
	}
	if err = pswdInput.Input(pswd); err != nil {
		return fmt.Errorf("failed to input password: %w", err)
	}

	submitButton, submitErr := page.Timeout(10 * time.Second).Element(submitButtonSelector)
	if submitErr != nil {
		return fmt.Errorf("failed to find submit button: %w", submitErr)
	}
	if err = submitButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to submit login: %w", err)
	}

	if _, loginCheckErr := page.Timeout(20 * time.Second).Element(myPageSelector); loginCheckErr != nil {
		return fmt.Errorf("login verification failed: %w", loginCheckErr)
	}

	log.Printf("srLogin: login success\n")

	return
}
