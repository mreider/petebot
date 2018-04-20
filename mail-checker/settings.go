package main

import (
	"encoding/json"
	"io/ioutil"
)

// Service settings.
type Settings struct {
	ServerInternalAddress string              `json:"SERVER_INTERNAL_ADDRESS"`
	TopicName             string              `json:"TOPIC_NAME"`
	DatabaseUri           string              `json:"DATABASE_URI"`
	Patterns              map[string][]string `json:"PATTERNS"`
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
