package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const missionConfigPath = "mission_themes.yml"

type Mission struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name"`
	Selector      string `yaml:"selector"`
	ProgressMode  string `yaml:"progress_mode"`
	RewardMode    string `yaml:"reward_mode"`
	ReceivePolicy string `yaml:"receive_policy"`
	ReceiveHours  []int  `yaml:"receive_hours"`
	CommentMode   string `yaml:"comment_mode"`
	CommentText   string `yaml:"comment_text"`
	GiftType      string `yaml:"gift_type"`
	Gift          int    `yaml:"gift"`
}

type Thema struct {
	ID       string    `yaml:"id"`
	Order    int       `yaml:"order"`
	Name     string    `yaml:"name"`
	Missions []Mission `yaml:"missions"`
}

type MissionThemeConfig struct {
	Themes []Thema `yaml:"themes"`
}

type ThemaList map[string]Thema

var themaConfig MissionThemeConfig
var themaList ThemaList

func init() {
	if err := loadThemaConfig(missionConfigPath); err != nil {
		panic(err)
	}
}

func loadThemaConfig(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read mission config %s: %w", path, err)
	}

	var cfg MissionThemeConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("parse mission config %s: %w", path, err)
	}
	if err := validateMissionThemeConfig(&cfg); err != nil {
		return fmt.Errorf("validate mission config %s: %w", path, err)
	}

	themaConfig = cfg
	themaList = make(ThemaList, len(cfg.Themes))
	for _, theme := range cfg.Themes {
		themaList[theme.ID] = theme
	}
	return nil
}

func validateMissionThemeConfig(cfg *MissionThemeConfig) error {
	if len(cfg.Themes) == 0 {
		return fmt.Errorf("themes is empty")
	}

	seenThemeIDs := make(map[string]struct{}, len(cfg.Themes))
	for themeIndex := range cfg.Themes {
		theme := &cfg.Themes[themeIndex]
		if theme.ID == "" {
			return fmt.Errorf("themes[%d].id is required", themeIndex)
		}
		if _, exists := seenThemeIDs[theme.ID]; exists {
			return fmt.Errorf("duplicate theme id: %s", theme.ID)
		}
		seenThemeIDs[theme.ID] = struct{}{}
		if theme.Name == "" {
			return fmt.Errorf("themes[%d].name is required", themeIndex)
		}
		if len(theme.Missions) == 0 {
			return fmt.Errorf("themes[%d].missions is empty", themeIndex)
		}

		seenMissionIDs := make(map[string]struct{}, len(theme.Missions))
		for missionIndex := range theme.Missions {
			mission := &theme.Missions[missionIndex]
			applyMissionDefaults(mission)
			if mission.ID == "" {
				return fmt.Errorf("themes[%d].missions[%d].id is required", themeIndex, missionIndex)
			}
			if _, exists := seenMissionIDs[mission.ID]; exists {
				return fmt.Errorf("duplicate mission id %s in theme %s", mission.ID, theme.ID)
			}
			seenMissionIDs[mission.ID] = struct{}{}
			if mission.Selector == "" {
				return fmt.Errorf("themes[%d].missions[%d].selector is required", themeIndex, missionIndex)
			}
			if err := validateMission(mission); err != nil {
				return fmt.Errorf("theme %s mission %s: %w", theme.ID, mission.ID, err)
			}
		}
	}

	return nil
}

func applyMissionDefaults(mission *Mission) {
	if mission.ProgressMode == "" {
		mission.ProgressMode = progressModeNone
	}
	if mission.RewardMode == "" {
		mission.RewardMode = rewardModeNone
	}
	if mission.ReceivePolicy == "" && mission.RewardMode != rewardModeNone {
		mission.ReceivePolicy = receivePolicyImmediate
	}
	if mission.CommentMode == "" {
		mission.CommentMode = commentModeNone
	}
}

func validateMission(mission *Mission) error {
	switch mission.ProgressMode {
	case progressModeNone, progressModeAchieveText, progressModeReceivedNum:
	default:
		return fmt.Errorf("unknown progress_mode: %s", mission.ProgressMode)
	}

	switch mission.RewardMode {
	case rewardModeNone, rewardModeOnce, rewardModeIncremental:
	default:
		return fmt.Errorf("unknown reward_mode: %s", mission.RewardMode)
	}

	switch mission.ReceivePolicy {
	case "":
		if mission.RewardMode != rewardModeNone {
			return fmt.Errorf("receive_policy is required when reward_mode=%s", mission.RewardMode)
		}
	case receivePolicyImmediate, receivePolicyManualUntilHours:
	default:
		return fmt.Errorf("unknown receive_policy: %s", mission.ReceivePolicy)
	}

	if mission.ReceivePolicy == receivePolicyManualUntilHours && len(mission.ReceiveHours) == 0 {
		return fmt.Errorf("receive_hours is required when receive_policy=%s", receivePolicyManualUntilHours)
	}

	switch mission.CommentMode {
	case commentModeNone, commentModeThanks39, commentModeGreeting:
	case commentModeText:
		if mission.CommentText == "" {
			return fmt.Errorf("comment_text is required when comment_mode=%s", commentModeText)
		}
	default:
		return fmt.Errorf("unknown comment_mode: %s", mission.CommentMode)
	}

	if mission.Gift < 0 {
		return fmt.Errorf("gift must be >= 0")
	}

	return nil
}
