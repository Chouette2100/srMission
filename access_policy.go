package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const accessHistoryPath = "access_history.yml"

type AccessHistory struct {
	Records map[string]AccessRecord `yaml:"records"`
}

type AccessRecord struct {
	ViewedAt string `yaml:"viewed_at"`
	ThemeID  string `yaml:"theme_id,omitempty"`
}

type legacyAccessHistory struct {
	LastAccessedAt string `yaml:"last_accessed_at"`
	LastURL        string `yaml:"last_url"`
	LastThemeID    string `yaml:"last_theme_id"`
}

func newAccessHistory() AccessHistory {
	return AccessHistory{Records: map[string]AccessRecord{}}
}

func loadAccessHistory(path string) (AccessHistory, error) {
	h := newAccessHistory()
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return h, nil
		}
		return h, fmt.Errorf("read access history %s: %w", path, err)
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return h, nil
	}
	if err := yaml.Unmarshal(raw, &h); err != nil {
		return h, fmt.Errorf("parse access history %s: %w", path, err)
	}
	h.ensureRecords()
	if len(h.Records) == 0 {
		legacy := legacyAccessHistory{}
		if err := yaml.Unmarshal(raw, &legacy); err != nil {
			return h, fmt.Errorf("parse access history legacy %s: %w", path, err)
		}
		if strings.TrimSpace(legacy.LastURL) != "" && strings.TrimSpace(legacy.LastAccessedAt) != "" {
			h.Records[canonicalRoomURL(legacy.LastURL)] = AccessRecord{
				ViewedAt: legacy.LastAccessedAt,
				ThemeID:  legacy.LastThemeID,
			}
		}
	}

	canonicalized := make(map[string]AccessRecord, len(h.Records))
	for rawURL, rec := range h.Records {
		key := canonicalRoomURL(rawURL)
		if key == "" || strings.TrimSpace(rec.ViewedAt) == "" {
			continue
		}
		canonicalized[key] = rec
	}
	h.Records = canonicalized
	return h, nil
}

func (h AccessHistory) save(path string) error {
	h.ensureRecords()
	raw, err := yaml.Marshal(h)
	if err != nil {
		return fmt.Errorf("marshal access history: %w", err)
	}
	if err := os.WriteFile(path, raw, 0644); err != nil {
		return fmt.Errorf("write access history %s: %w", path, err)
	}
	return nil
}

func (h *AccessHistory) ensureRecords() {
	if h.Records == nil {
		h.Records = map[string]AccessRecord{}
	}
}

func (h AccessHistory) lastAccessTimeByURL(url string) (time.Time, bool, error) {
	key := canonicalRoomURL(url)
	rec, ok := h.Records[key]
	if !ok || strings.TrimSpace(rec.ViewedAt) == "" {
		return time.Time{}, false, nil
	}
	t, err := time.Parse(time.RFC3339, rec.ViewedAt)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid viewed_at %q for url %s: %w", rec.ViewedAt, key, err)
	}
	return t, true, nil
}

func (h *AccessHistory) setLastAccessByURL(url string, themeID string, at time.Time) {
	h.ensureRecords()
	key := canonicalRoomURL(url)
	if key == "" {
		return
	}
	h.Records[key] = AccessRecord{
		ViewedAt: at.Format(time.RFC3339),
		ThemeID:  themeID,
	}
}

func missionThemeID(mission string) string {
	if mission == "daily" {
		return "Daily"
	}
	return "NonDaily"
}

func filterRoomsByAccessPolicy(rooms []Room, themeID string, at time.Time, history AccessHistory) []Room {
	filtered := make([]Room, 0, len(rooms))
	shadowHistory := newAccessHistory()
	for k, v := range history.Records {
		shadowHistory.Records[k] = v
	}

	for _, room := range rooms {
		lastAccess, hasLast, err := shadowHistory.lastAccessTimeByURL(room.URL)
		if err != nil {
			log.Printf("filterRoomsByAccessPolicy: invalid history for url=%s err=%v\n", room.URL, err)
			continue
		}

		allowed, nextAllowedAt := canAccessByThemePolicy(themeID, at, lastAccess, hasLast)
		if !allowed {
			if !nextAllowedAt.IsZero() {
				log.Printf("filterRoomsByAccessPolicy: skip by policy. theme=%s url=%s last=%s next=%s\n",
					themeID,
					room.URL,
					lastAccess.Format("2006-01-02 15:04:05"),
					nextAllowedAt.Format("2006-01-02 15:04:05"),
				)
			} else {
				log.Printf("filterRoomsByAccessPolicy: skip by policy. theme=%s url=%s\n", themeID, room.URL)
			}
			continue
		}

		filtered = append(filtered, room)
		shadowHistory.setLastAccessByURL(room.URL, themeID, at)
	}

	return filtered
}

func canAccessByThemePolicy(themeID string, now time.Time, lastAccess time.Time, hasLast bool) (bool, time.Time) {
	if !hasLast {
		return true, time.Time{}
	}
	if isAccessAllowed(themeID, now, lastAccess) {
		return true, time.Time{}
	}

	next := now
	for i := 0; i < 8; i++ {
		next = nextPolicyBoundary(themeID, next)
		if isAccessAllowed(themeID, next, lastAccess) {
			return false, next
		}
	}
	return false, time.Time{}
}

func isAccessAllowed(themeID string, now time.Time, lastAccess time.Time) bool {
	cutoff := accessCutoff(themeID, now)
	return !lastAccess.After(cutoff)
}

func accessCutoff(themeID string, now time.Time) time.Time {
	y, m, d := now.Date()
	loc := now.Location()

	if themeID == "Daily" {
		hour := now.Hour()
		switch {
		case hour >= 3 && hour < 15:
			return time.Date(y, m, d, 3, 0, 0, 0, loc)
		case hour >= 15:
			return time.Date(y, m, d, 15, 0, 0, 0, loc)
		default:
			yesterday := now.AddDate(0, 0, -1)
			yy, ym, yd := yesterday.Date()
			return time.Date(yy, ym, yd, 15, 0, 0, 0, loc)
		}
	}

	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func nextPolicyBoundary(themeID string, now time.Time) time.Time {
	y, m, d := now.Date()
	loc := now.Location()

	if themeID == "Daily" {
		today3 := time.Date(y, m, d, 3, 0, 0, 0, loc)
		today15 := time.Date(y, m, d, 15, 0, 0, 0, loc)
		if now.Before(today3) {
			return today3
		}
		if now.Before(today15) {
			return today15
		}
		return time.Date(y, m, d, 3, 0, 0, 0, loc).AddDate(0, 0, 1)
	}

	return time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, 1)
}
