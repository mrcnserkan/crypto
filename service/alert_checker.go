package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mrcnserkan/crypto/models"
	"github.com/mrcnserkan/crypto/utils"
)

const defaultAlertCheckInterval = 5 * time.Minute

// AlertChecker periodically evaluates price alerts.
type AlertChecker struct {
	alertManager *models.AlertManager
	coinGecko    *CoinGecko
	stopChan     chan struct{}
	doneChan     chan struct{}
	mu           sync.Mutex
	interval     time.Duration
}

func NewAlertChecker(alertManager *models.AlertManager) *AlertChecker {
	return &AlertChecker{
		alertManager: alertManager,
		coinGecko:    NewCoinGecko(),
		interval:     defaultAlertCheckInterval,
	}
}

func (ac *AlertChecker) SetInterval(d time.Duration) {
	if d > 0 {
		ac.interval = d
	}
}

func (ac *AlertChecker) EnsureRunning() {
	if len(ac.alertManager.GetAlerts()) == 0 {
		return
	}
	ac.Start()
}

func (ac *AlertChecker) Start() {
	ac.mu.Lock()
	if ac.stopChan != nil {
		ac.mu.Unlock()
		return
	}

	ac.stopChan = make(chan struct{})
	ac.doneChan = make(chan struct{})
	stopChan := ac.stopChan
	doneChan := ac.doneChan
	ac.mu.Unlock()

	go ac.runLoop(stopChan, doneChan)
}

func (ac *AlertChecker) runLoop(stopChan, doneChan chan struct{}) {
	defer close(doneChan)

	ticker := time.NewTicker(ac.interval)
	defer ticker.Stop()

	ac.runAlertChecks(stopChan)

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			if len(ac.alertManager.GetAlerts()) == 0 {
				return
			}
			ac.runAlertChecks(stopChan)
		}
	}
}

func (ac *AlertChecker) runAlertChecks(stopChan <-chan struct{}) {
	alerts := ac.alertManager.GetAlerts()
	if len(alerts) == 0 {
		return
	}

	byCurrency := make(map[string][]models.Alert)
	for _, alert := range alerts {
		currency := utils.NormalizeCurrency(alert.Currency)
		byCurrency[currency] = append(byCurrency[currency], alert)
	}

	for currency, currencyAlerts := range byCurrency {
		select {
		case <-stopChan:
			return
		default:
		}

		coinIDs := make([]string, 0, len(currencyAlerts))
		seen := make(map[string]struct{})
		for _, alert := range currencyAlerts {
			if _, ok := seen[alert.CoinID]; ok {
				continue
			}
			seen[alert.CoinID] = struct{}{}
			coinIDs = append(coinIDs, alert.CoinID)
		}

		prices, err := ac.coinGecko.GetSimplePrices(coinIDs, currency)
		if err != nil {
			fmt.Printf("Error fetching prices for alerts: %v\n", err)
			continue
		}

		for _, alert := range currencyAlerts {
			currentPrice, ok := prices[alert.CoinID]
			if !ok {
				continue
			}
			if ac.isTriggered(alert, currentPrice) {
				ac.sendNotification(alert, currentPrice)
				if err := ac.alertManager.RemoveTriggeredAlert(alert); err != nil {
					fmt.Printf("Error removing triggered alert for %s: %v\n", alert.CoinID, err)
				}
			}
		}
	}
}

func (ac *AlertChecker) isTriggered(alert models.Alert, currentPrice float64) bool {
	if alert.Condition == "above" {
		return currentPrice >= alert.Price
	}
	return currentPrice <= alert.Price
}

func (ac *AlertChecker) Stop() {
	ac.mu.Lock()
	if ac.stopChan == nil {
		ac.mu.Unlock()
		return
	}

	close(ac.stopChan)
	doneChan := ac.doneChan
	ac.stopChan = nil
	ac.doneChan = nil
	ac.mu.Unlock()

	<-doneChan
}

func (ac *AlertChecker) RunOnce() {
	ac.runAlertChecks(nil)
}

func (ac *AlertChecker) sendNotification(alert models.Alert, currentPrice float64) {
	currency := utils.NormalizeCurrency(alert.Currency)
	currencySymbol := utils.CurrencySymbol(currency)

	message := fmt.Sprintf("Price Alert: %s is %s %s%.2f (Target: %s%.2f %s)",
		strings.ToUpper(alert.CoinID),
		alert.Condition,
		currencySymbol,
		currentPrice,
		currencySymbol,
		alert.Price,
		strings.ToUpper(currency))

	fmt.Printf("\n%s\n", message)
}
