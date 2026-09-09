package main

import (
	"os"
	"testing"
)

func TestParsePercent(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", -1},
		{"whitespace", "   ", -1},
		{"zero", "0%", 0},
		{"fifty", "50%", 50},
		{"one hundred", "100%", 100},
		{"px value", "100px", 100},
		{"decimal ignored", "45.5%", 45},
		{"with spaces", "  75 %  ", 75},
		{"no number", "completed", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePercent(tt.in)
			if got != tt.want {
				t.Errorf("parsePercent(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestAdRewardWaitSeconds(t *testing.T) {
	tests := []struct {
		retryCount int
		wantMin    int
		wantMax    int
	}{
		{-1, 5, 5},
		{0, 5, 5},
		{1, 10, 10},
		{10, 50, 100},
		{30, 50, 100},
	}

	for _, tt := range tests {
		got := adRewardWaitSeconds(tt.retryCount)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("adRewardWaitSeconds(%d) = %d, want between %d and %d", tt.retryCount, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestAdRewardNoProgressWaitSeconds(t *testing.T) {
	orig := os.Getenv("SR_ADREWARD_NO_PROGRESS_WAIT_SEC")
	defer os.Setenv("SR_ADREWARD_NO_PROGRESS_WAIT_SEC", orig)

	os.Unsetenv("SR_ADREWARD_NO_PROGRESS_WAIT_SEC")
	if got := adRewardNoProgressWaitSeconds(); got != 50 {
		t.Errorf("default = %d, want 50", got)
	}

	os.Setenv("SR_ADREWARD_NO_PROGRESS_WAIT_SEC", "60")
	if got := adRewardNoProgressWaitSeconds(); got != 60 {
		t.Errorf("custom = %d, want 60", got)
	}

	os.Setenv("SR_ADREWARD_NO_PROGRESS_WAIT_SEC", "invalid")
	if got := adRewardNoProgressWaitSeconds(); got != 50 {
		t.Errorf("invalid fallback = %d, want 50", got)
	}
}

func TestIsDebugScreenshotEnabled(t *testing.T) {
	orig := os.Getenv("SR_ADREWARD_DEBUG_SCREENSHOTS")
	defer os.Setenv("SR_ADREWARD_DEBUG_SCREENSHOTS", orig)

	cases := []struct {
		value string
		want  bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"ON", true},
	}

	for _, c := range cases {
		os.Setenv("SR_ADREWARD_DEBUG_SCREENSHOTS", c.value)
		got := isDebugScreenshotEnabled()
		if got != c.want {
			t.Errorf("isDebugScreenshotEnabled(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestIsDebugHTMLDumpEnabled(t *testing.T) {
	orig := os.Getenv("SR_ADREWARD_DEBUG_HTML")
	defer os.Setenv("SR_ADREWARD_DEBUG_HTML", orig)

	cases := []struct {
		value string
		want  bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"ON", true},
	}

	for _, c := range cases {
		os.Setenv("SR_ADREWARD_DEBUG_HTML", c.value)
		got := isDebugHTMLDumpEnabled()
		if got != c.want {
			t.Errorf("isDebugHTMLDumpEnabled(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestDumpPageHTML(t *testing.T) {
	orig := os.Getenv("SR_ADREWARD_DEBUG_HTML")
	defer os.Setenv("SR_ADREWARD_DEBUG_HTML", orig)

	// フラグ OFF 時は何もしないことを確認。
	os.Setenv("SR_ADREWARD_DEBUG_HTML", "0")
	dumpPageHTML(nil, "test")

	// フラグ ON でも page=nil なら何もしない。
	os.Setenv("SR_ADREWARD_DEBUG_HTML", "1")
	dumpPageHTML(nil, "test")
}
func TestDebugScreenshotDir(t *testing.T) {
	orig := os.Getenv("SR_ADREWARD_DEBUG_SCREENSHOT_DIR")
	defer os.Setenv("SR_ADREWARD_DEBUG_SCREENSHOT_DIR", orig)

	os.Unsetenv("SR_ADREWARD_DEBUG_SCREENSHOT_DIR")
	if got := debugScreenshotDir(); got != "screenshots" {
		t.Errorf("default dir = %q, want screenshots", got)
	}

	os.Setenv("SR_ADREWARD_DEBUG_SCREENSHOT_DIR", "/tmp/ss")
	if got := debugScreenshotDir(); got != "/tmp/ss" {
		t.Errorf("custom dir = %q, want /tmp/ss", got)
	}
}
