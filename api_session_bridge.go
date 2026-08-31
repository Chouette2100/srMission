package main

import (
	"fmt"

	"net/http"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/juju/persistent-cookiejar"

	"github.com/Chouette2100/srapi/v2"
)

func PrepareAPIClientFromCurrentBrowser(cookiename string, pageURL string) (client *http.Client, jar *cookiejar.Jar, err error) {
	if srBrowser == nil {
		return nil, nil, fmt.Errorf("browser is not initialized")
	}
	if pageURL == "" {
		pageURL = "https://www.showroom-live.com/"
	}

	page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bridge page: %w", err)
	}
	defer page.Close()

	if err = applyJapaneseLocale(page); err != nil {
		return nil, nil, err
	}
	if err = page.Navigate(pageURL); err != nil {
		return nil, nil, fmt.Errorf("failed to navigate bridge page: %w", err)
	}
	if err = page.WaitLoad(); err != nil {
		return nil, nil, fmt.Errorf("failed to wait bridge page load: %w", err)
	}

	return PrepareAPIClientFromBrowser(page, cookiename, pageURL)
}

// PrepareAPIClientFromBrowser exports the current browser session into a persistent-cookiejar,
// then creates an API client that can reuse the same SHOWROOM login state.
//
// The API side already knows how to fetch csrf_token from SHOWROOM, so this helper only
// prepares the shared cookie session. Call srapi.ApiCsrftoken(client) afterwards.
func PrepareAPIClientFromBrowser(page *rod.Page, cookiename string, pageURL string) (client *http.Client, jar *cookiejar.Jar, err error) {
	client, jar, err = srapi.CreateNewClient(cookiename)
	if err != nil {
		return nil, nil, fmt.Errorf("srapi.CreateNewClient: %w", err)
	}

	if err = CopyRodPageCookiesToJar(page, jar, pageURL); err != nil {
		return nil, nil, fmt.Errorf("CopyRodPageCookiesToJar: %w", err)
	}

	return client, jar, nil
}

// FetchAPICSRFToken retrieves csrf_token through the API client.
//
// SHOWROOM exposes /api/csrf_token, so this is the supported way to obtain the token
// once the browser session has been exported into the API client.
func FetchAPICSRFToken(client *http.Client) (string, error) {
	csrftoken, err := srapi.ApiCsrftoken(client)
	if err != nil {
		return "", fmt.Errorf("srapi.ApiCsrftoken: %w", err)
	}
	return csrftoken, nil
}
