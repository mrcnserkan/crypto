/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrcnserkan/crypto/models"
)

type AlertChecker struct {
	alertManager *models.AlertManager
	coinGecko    *CoinGecko
	stopChan     chan struct{}
}

func NewAlertChecker(alertManager *models.AlertManager) *AlertChecker {
	return &AlertChecker{
		alertManager: alertManager,
		coinGecko:    NewCoinGecko(),
		stopChan:     make(chan struct{}),
	}
}

func (ac *AlertChecker) Start() {
	ac.stopChan = make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		rateLimiter := time.NewTicker(2 * time.Second)

		defer ticker.Stop()
		defer rateLimiter.Stop()

		for {
			select {
			case <-ac.stopChan:
				return
			case <-ticker.C:
				alerts := ac.alertManager.GetAlerts()
				for _, alert := range alerts {
					select {
					case <-rateLimiter.C:
						ac.checkAlert(alert)
					case <-ac.stopChan:
						return
					}
				}
			}
		}
	}()
}

func (ac *AlertChecker) Stop() {
	close(ac.stopChan)
}

func (ac *AlertChecker) checkAlert(alert models.Alert) {
	coin, err := ac.coinGecko.GetCoinDetail(alert.CoinID)
	if err != nil {
		fmt.Printf("Error checking alert for %s: %v\n", alert.CoinID, err)
		return
	}

	currentPrice := coin.MarketData.CurrentPrice["usd"]
	if (alert.Condition == "above" && currentPrice >= alert.Price) ||
		(alert.Condition == "below" && currentPrice <= alert.Price) {
		ac.sendNotification(alert, currentPrice)
		ac.alertManager.RemoveAlert(alert.CoinID)
	}
}

func (ac *AlertChecker) sendNotification(alert models.Alert, currentPrice float64) {
	message := fmt.Sprintf("🚨 Price Alert: %s is %s $%.2f (Target: $%.2f)",
		strings.ToUpper(alert.CoinID),
		alert.Condition,
		currentPrice,
		alert.Price)

	fmt.Printf("\n%s\n", message)
	// TODO: Implement desktop notifications using platform-specific libraries
	// Windows: github.com/go-toast/toast
	// Linux: github.com/esiqveland/notify
	// macOS: github.com/gen2brain/beeep
}
