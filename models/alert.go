/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Alert struct {
	CoinID    string    `json:"coin_id"`
	Price     float64   `json:"price"`
	Condition string    `json:"condition"` // "above" or "below"
	Currency  string    `json:"currency,omitempty"`
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

func normalizeAlert(alert *Alert) {
	alert.CoinID = strings.ToLower(strings.TrimSpace(alert.CoinID))
	alert.Condition = strings.ToLower(strings.TrimSpace(alert.Condition))
	if alert.Currency == "" {
		alert.Currency = "usd"
	} else {
		alert.Currency = strings.ToLower(strings.TrimSpace(alert.Currency))
	}
}

func (am *AlertManager) AddAlert(alert Alert) error {
	if alert.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}
	if alert.Condition != "above" && alert.Condition != "below" {
		return fmt.Errorf("condition must be 'above' or 'below'")
	}

	normalizeAlert(&alert)

	for _, existingAlert := range am.alerts {
		if existingAlert.CoinID == alert.CoinID &&
			existingAlert.Price == alert.Price &&
			existingAlert.Condition == alert.Condition &&
			existingAlert.Currency == alert.Currency {
			return fmt.Errorf("alert already exists for %s at %.2f %s (%s)",
				alert.CoinID, alert.Price, alert.Condition, strings.ToUpper(alert.Currency))
		}
	}

	alert.CreatedAt = time.Now()
	am.alerts = append(am.alerts, alert)
	return am.Save()
}

func (am *AlertManager) RemoveAlert(coinID string) error {
	return am.RemoveAlertsForCoin(coinID)
}

// RemoveAlertsForCoin removes all alerts for a coin.
func (am *AlertManager) RemoveAlertsForCoin(coinID string) error {
	coinID = strings.ToLower(strings.TrimSpace(coinID))

	var remaining []Alert
	removed := false
	for _, alert := range am.alerts {
		if alert.CoinID == coinID {
			removed = true
			continue
		}
		remaining = append(remaining, alert)
	}
	if !removed {
		return fmt.Errorf("alert not found for coin: %s", coinID)
	}

	am.alerts = remaining
	return am.Save()
}

func (am *AlertManager) RemoveTriggeredAlert(alert Alert) error {
	normalizeAlert(&alert)

	for i, existingAlert := range am.alerts {
		if existingAlert.CoinID == alert.CoinID &&
			existingAlert.Price == alert.Price &&
			existingAlert.Condition == alert.Condition &&
			existingAlert.Currency == alert.Currency {
			am.alerts = append(am.alerts[:i], am.alerts[i+1:]...)
			return am.Save()
		}
	}
	return fmt.Errorf("triggered alert not found for coin: %s", alert.CoinID)
}

// RemoveAlertByTarget removes a specific alert by coin, price, and condition.
func (am *AlertManager) RemoveAlertByTarget(coinID string, price float64, condition string) error {
	coinID = strings.ToLower(strings.TrimSpace(coinID))
	condition = strings.ToLower(strings.TrimSpace(condition))

	var remaining []Alert
	removed := false
	for _, alert := range am.alerts {
		if alert.CoinID == coinID && alert.Price == price && alert.Condition == condition {
			removed = true
			continue
		}
		remaining = append(remaining, alert)
	}
	if !removed {
		return fmt.Errorf("alert not found for %s at %.2f %s", coinID, price, condition)
	}
	am.alerts = remaining
	return am.Save()
}

func (am *AlertManager) GetAlerts() []Alert {
	if len(am.alerts) == 0 {
		return nil
	}
	return append([]Alert(nil), am.alerts...)
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
		am.normalizeLoadedAlerts()
		if am.needsMigration(alerts) {
			return am.Save()
		}
		return nil
	}

	// Try to load old format
	var alertData AlertData
	err = json.Unmarshal(data, &alertData)
	if err != nil {
		return err
	}

	am.alerts = alertData.Alerts
	am.normalizeLoadedAlerts()
	return am.Save()
}

func (am *AlertManager) needsMigration(alerts []Alert) bool {
	for _, alert := range alerts {
		if alert.Currency == "" {
			return true
		}
	}
	return false
}

func (am *AlertManager) normalizeLoadedAlerts() {
	for i := range am.alerts {
		normalizeAlert(&am.alerts[i])
	}
}

func (am *AlertManager) Save() error {
	data, err := json.MarshalIndent(am.alerts, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(am.alertFile, data)
}
