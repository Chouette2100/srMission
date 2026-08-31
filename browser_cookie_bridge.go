package main

import (
	"fmt"
	"net/http"
	"github.com/juju/persistent-cookiejar"
	"net/url"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// CopyRodPageCookiesToJar copies cookies from a rod page into a persistent-cookiejar.
//
// Use this after browser login when you want a separate API client to reuse the same
// SHOWROOM session.
// HttpOnly cookies are included; they cannot be read from JavaScript, but rod/CDP can
// fetch them and convert them into net/http cookies.
func CopyRodPageCookiesToJar(page *rod.Page, jar *cookiejar.Jar, pageURL string) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}
	if jar == nil {
		return fmt.Errorf("jar is nil")
	}

	if pageURL == "" {
		info, err := page.Info()
		if err != nil {
			return fmt.Errorf("failed to get page info: %w", err)
		}
		pageURL = info.URL
	}

	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("failed to parse page url %q: %w", pageURL, err)
	}

	cookies, err := page.Cookies([]string{pageURL})
	if err != nil {
		return fmt.Errorf("failed to get page cookies: %w", err)
	}

	httpCookies := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		httpCookie := ToHTTPCookie(cookie)
		if httpCookie != nil {
			httpCookies = append(httpCookies, httpCookie)
		}
	}

	jar.SetCookies(baseURL, httpCookies)
	return nil
}

// CopyRodCookiesToJar copies cookies for the given URLs into the jar.
// This is useful when you want to export cookies for multiple SHOWROOM domains.
func CopyRodCookiesToJar(page *rod.Page, jar *cookiejar.Jar, urls []string) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}
	if jar == nil {
		return fmt.Errorf("jar is nil")
	}
	if len(urls) == 0 {
		return fmt.Errorf("urls is empty")
	}

	cookies, err := page.Cookies(urls)
	if err != nil {
		return fmt.Errorf("failed to get page cookies: %w", err)
	}

	for _, rawURL := range urls {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("failed to parse url %q: %w", rawURL, err)
		}

		var perURLCookies []*http.Cookie
		for _, cookie := range cookies {
			if cookie == nil {
				continue
			}
			if httpCookie := ToHTTPCookie(cookie); httpCookie != nil {
				perURLCookies = append(perURLCookies, httpCookie)
			}
		}

		jar.SetCookies(parsedURL, perURLCookies)
	}

	return nil
}

// ToHTTPCookie converts a rod cookie into an http.Cookie.
func ToHTTPCookie(cookie *proto.NetworkCookie) *http.Cookie {
	if cookie == nil {
		return nil
	}

	httpCookie := &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Domain:   cookie.Domain,
		Path:     cookie.Path,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HTTPOnly,
	}
	if !cookie.Session && cookie.Expires > 0 {
		httpCookie.Expires = time.Unix(0, int64(float64(cookie.Expires)*float64(time.Second)))
	}

	return httpCookie
}
