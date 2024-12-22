/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Alert struct {
	CoinID    string    `json:"coin_id"`
	Price     float64   `json:"price"`
	Condition string    `json:"condition"` // "above" or "below"
	CreatedAt time.Time `json:"created_at"`
}

type AlertData struct {
	Alerts []Alert `json:"alerts"`
}

type AlertManager struct {
	alerts    []Alert
	alertFile string
}

func NewAlertManager(configDir string) *AlertManager {
	return &AlertManager{
		alerts:    make([]Alert, 0),
		alertFile: filepath.Join(configDir, "alerts.json"),
	}
}

func (am *AlertManager) AddAlert(alert Alert) error {
	// Check if alert already exists
	for _, existingAlert := range am.alerts {
		if existingAlert.CoinID == alert.CoinID &&
			existingAlert.Price == alert.Price &&
			existingAlert.Condition == alert.Condition {
			return fmt.Errorf("alert already exists for %s at %.2f %s",
				alert.CoinID, alert.Price, alert.Condition)
		}
	}

	alert.CreatedAt = time.Now()
	am.alerts = append(am.alerts, alert)
	return am.Save()
}

func (am *AlertManager) RemoveAlert(coinID string) error {
	for i, alert := range am.alerts {
		if alert.CoinID == coinID {
			am.alerts = append(am.alerts[:i], am.alerts[i+1:]...)
			return am.Save()
		}
	}
	return fmt.Errorf("alert not found for coin: %s", coinID)
}

func (am *AlertManager) GetAlerts() []Alert {
	return am.alerts
}

func (am *AlertManager) Load() error {
	data, err := os.ReadFile(am.alertFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Try to load new format first
	var alerts []Alert
	err = json.Unmarshal(data, &alerts)
	if err == nil {
		am.alerts = alerts
		return nil
	}

	// Try to load old format
	var alertData AlertData
	err = json.Unmarshal(data, &alertData)
	if err != nil {
		return err
	}

	am.alerts = alertData.Alerts
	return am.Save() // Save in new format
}

func (am *AlertManager) Save() error {
	data, err := json.MarshalIndent(am.alerts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(am.alertFile, data, 0644)
}
