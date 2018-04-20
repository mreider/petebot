package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var UsersSeparatorsRe = regexp.MustCompile(`[,;]`)

// Service settings.
type Settings struct {
	ServerInternalAddress                     string `json:"SERVER_INTERNAL_ADDRESS"`
	ServerPublicAddress                       string `json:"SERVER_PUBLIC_ADDRESS"`
	StravaOauthClientSecret                   string `json:"STRAVA_OAUTH_CLIENT_SECRET"`
	StravaOauthClientId                       int    `json:"STRAVA_OAUTH_CLIENT_ID"`
	SlackOauthClientId                        string `json:"SLACK_OAUTH_CLIENT_ID"`
	SlackOauthClientSecret                    string `json:"SLACK_OAUTH_CLIENT_SECRET"`
	DatabaseUri                               string `json:"DATABASE_URI"`
	StravaCommonMonitoringIntervalInMinutes   int    `json:"STRAVA_COMMON_MONITORING_INTERVAL_IN_MINUTES"`
	StravaUnlockedMonitoringIntervalInMinutes int    `json:"STRAVA_UNLOCKED_MONITORING_INTERVAL_IN_MINUTES"`
	AdminStravaUsers                          string `json:"ADMIN_STRAVA_USERS"`

	adminUsers map[int]bool
}

func (s *Settings) IsAdmin(stravaUserId int) bool {
	return s.adminUsers[stravaUserId]
}

// Load merged settings from specified default and local settings files.
func LoadSettings(settingsDefaultPath, settingsLocalPath string) (*Settings, error) {
	settings, err := loadRawSettings(settingsDefaultPath)
	if err != nil {
		return nil, err
	}
	localSettings, err := loadRawSettings(settingsLocalPath)
	if err != nil {
		return nil, err
	}
	for key, value := range localSettings {
		settings[key] = value
	}

	mergedJson, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	var mergedSettings *Settings
	err = json.Unmarshal(mergedJson, &mergedSettings)
	if err != nil {
		return nil, err
	}

	mergedSettings.adminUsers = make(map[int]bool)
	users := UsersSeparatorsRe.Split(mergedSettings.AdminStravaUsers, -1)
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}

		userId, err := strconv.Atoi(user)
		if err == nil {
			mergedSettings.adminUsers[userId] = true
		} else {
			log.Printf("Admin user Id should be integer: %s\n", user)
		}
	}

	return mergedSettings, nil
}

// Load raw settings from the specified file.
func loadRawSettings(settingsPath string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		return nil, err
	}
	var settings map[string]interface{}
	err = json.Unmarshal(content, &settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
