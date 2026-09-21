package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const urlBlacklistPath = "url_blacklist.yml"

type URLBlockEntry struct {
	URL    string `yaml:"url"`
	Reason string `yaml:"reason"`
}

type URLBlacklistConfig struct {
	Entries []URLBlockEntry `yaml:"entries"`
}

var urlBlacklist map[string]URLBlockEntry

func loadURLBlacklist(path string) error {
	urlBlacklist = map[string]URLBlockEntry{}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read URL blacklist %s: %w", path, err)
	}

	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil
	}

	var cfg URLBlacklistConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("parse URL blacklist %s: %w", path, err)
	}

	for _, entry := range cfg.Entries {
		k := canonicalRoomURL(entry.URL)
		if k == "" {
			continue
		}
		urlBlacklist[k] = entry
	}
	return nil
}

func canonicalRoomURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "https://www.showroom-live.com/r/")
	s = strings.TrimPrefix(s, "http://www.showroom-live.com/r/")
	s = strings.TrimPrefix(s, "/r/")
	s = strings.TrimPrefix(s, "r/")
	s = strings.Trim(s, "/")
	return s
}

func isRoomURLBlacklisted(url string) (URLBlockEntry, bool) {
	entry, ok := urlBlacklist[canonicalRoomURL(url)]
	return entry, ok
}
