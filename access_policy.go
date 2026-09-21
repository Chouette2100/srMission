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
	LastAccessedAt string `yaml:"last_accessed_at"`
	LastURL        string `yaml:"last_url"`
	LastThemeID    string `yaml:"last_theme_id"`
}

func loadAccessHistory(path string) (AccessHistory, error) {
	h := AccessHistory{}
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
	return h, nil
}

func (h AccessHistory) save(path string) error {
	raw, err := yaml.Marshal(h)
	if err != nil {
		return fmt.Errorf("marshal access history: %w", err)
	}
	if err := os.WriteFile(path, raw, 0644); err != nil {
		return fmt.Errorf("write access history %s: %w", path, err)
	}
	return nil
}

func (h AccessHistory) lastAccessTime() (time.Time, bool, error) {
	if strings.TrimSpace(h.LastAccessedAt) == "" {
		return time.Time{}, false, nil
	}
	t, err := time.Parse(time.RFC3339, h.LastAccessedAt)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid last_accessed_at %q: %w", h.LastAccessedAt, err)
	}
	return t, true, nil
}

func missionThemeID(mission string) string {
	if mission == "daily" {
		return "Daily"
	}
	return "NonDaily"
}

func filterRoomsByAccessPolicy(rooms []Room, themeID string, at time.Time, lastAccess time.Time, hasLast bool) []Room {
	filtered := make([]Room, 0, len(rooms))
	shadowLast := lastAccess
	shadowHasLast := hasLast

	for _, room := range rooms {
		allowed, nextAllowedAt := canAccessByThemePolicy(themeID, at, shadowLast, shadowHasLast)
		if !allowed {
			if !nextAllowedAt.IsZero() {
				log.Printf("filterRoomsByAccessPolicy: skip by policy. theme=%s url=%s last=%s next=%s\n",
					themeID,
					room.URL,
					shadowLast.Format("2006-01-02 15:04:05"),
					nextAllowedAt.Format("2006-01-02 15:04:05"),
				)
			} else {
				log.Printf("filterRoomsByAccessPolicy: skip by policy. theme=%s url=%s\n", themeID, room.URL)
			}
			continue
		}

		filtered = append(filtered, room)
		shadowLast = at
		shadowHasLast = true
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
