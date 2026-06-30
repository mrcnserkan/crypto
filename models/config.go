package models

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig holds user preferences stored in ~/.crypto/config.json.
type AppConfig struct {
	Currency            string `json:"currency,omitempty"`
	ChartWidth          int    `json:"chart_width,omitempty"`
	ChartHeight         int    `json:"chart_height,omitempty"`
	AlertCheckIntervalM int    `json:"alert_check_interval_minutes,omitempty"`
	NoColor             bool   `json:"no_color,omitempty"`
}

type ConfigStore struct {
	filePath string
	Config   AppConfig
}

func NewConfigStore(configDir string) *ConfigStore {
	return &ConfigStore{filePath: filepath.Join(configDir, "config.json")}
}

func (cs *ConfigStore) Load() error {
	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &cs.Config)
}

func (cs *ConfigStore) Save() error {
	data, err := json.MarshalIndent(cs.Config, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(cs.filePath, data)
}

func (cs *ConfigStore) CurrencyOrDefault(defaultCurrency string) string {
	if cs.Config.Currency != "" {
		return normalizeCurrency(cs.Config.Currency)
	}
	return normalizeCurrency(defaultCurrency)
}

func (cs *ConfigStore) ChartWidthOrDefault(defaultWidth int) int {
	if cs.Config.ChartWidth > 0 {
		return cs.Config.ChartWidth
	}
	return defaultWidth
}

func (cs *ConfigStore) ChartHeightOrDefault(defaultHeight int) int {
	if cs.Config.ChartHeight > 0 {
		return cs.Config.ChartHeight
	}
	return defaultHeight
}

func (cs *ConfigStore) AlertIntervalOrDefault(defaultMinutes int) int {
	if cs.Config.AlertCheckIntervalM > 0 {
		return cs.Config.AlertCheckIntervalM
	}
	return defaultMinutes
}
