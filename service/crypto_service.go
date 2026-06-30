/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mrcnserkan/crypto/models"
	"github.com/mrcnserkan/crypto/utils"
)

const (
	DEFAULT_PAGE            = "1"
	DEFAULT_CURRENCY        = "usd"
	DEFAULT_CURRENCY_SYMBOL = "$"
	PER_PAGE                = "10"
	maxHTTPRetries          = 3
)

var BaseURL = "https://api.coingecko.com/api/v3"

type Interval struct {
	Name  string
	Value string
	Days  int
}

var (
	Intervals = []Interval{
		{Name: "1d", Value: "daily", Days: 1},
		{Name: "7d", Value: "daily", Days: 7},
		{Name: "14d", Value: "daily", Days: 14},
		{Name: "30d", Value: "daily", Days: 30},
		{Name: "90d", Value: "daily", Days: 90},
		{Name: "180d", Value: "daily", Days: 180},
		{Name: "1y", Value: "daily", Days: 365},
		{Name: "max", Value: "daily", Days: -1},
	}
)

type CoinGecko struct {
	client *http.Client
}

func NewCoinGecko() *CoinGecko {
	return &CoinGecko{
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (cg *CoinGecko) get(requestURL string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxHTTPRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err := cg.client.Get(requestURL)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return body, nil
		case http.StatusTooManyRequests:
			lastErr = fmt.Errorf("rate limit exceeded")
			continue
		default:
			return nil, fmt.Errorf("API error: %s", resp.Status)
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed after %d attempts", maxHTTPRetries)
	}
	return nil, lastErr
}

func (cg *CoinGecko) GetMarkets(currency string, perPage int, page int) ([]models.Coin, error) {
	requestURL := fmt.Sprintf("%s/coins/markets?vs_currency=%s&order=market_cap_desc&per_page=%d&page=%d&sparkline=false&price_change_percentage=24h,7d",
		BaseURL, strings.ToLower(currency), perPage, page)

	bodyBytes, err := cg.get(requestURL)
	if err != nil {
		return nil, err
	}

	var coins []models.Coin
	if err := json.Unmarshal(bodyBytes, &coins); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return coins, nil
}

func (cg *CoinGecko) GetMarketsByIDs(currency string, ids []string) ([]models.Coin, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	normalizedIDs := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.ToLower(strings.TrimSpace(id))
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalizedIDs = append(normalizedIDs, id)
	}
	if len(normalizedIDs) == 0 {
		return nil, nil
	}

	currency = utils.NormalizeCurrency(currency)
	allCoins := make([]models.Coin, 0, len(normalizedIDs))

	for start := 0; start < len(normalizedIDs); start += utils.MaxMarketIDsPerBatch {
		end := start + utils.MaxMarketIDsPerBatch
		if end > len(normalizedIDs) {
			end = len(normalizedIDs)
		}

		coins, err := cg.getMarketsByIDsChunk(currency, normalizedIDs[start:end])
		if err != nil {
			return nil, err
		}
		allCoins = append(allCoins, coins...)
	}

	return allCoins, nil
}

func (cg *CoinGecko) getMarketsByIDsChunk(currency string, ids []string) ([]models.Coin, error) {
	requestURL := fmt.Sprintf("%s/coins/markets?vs_currency=%s&ids=%s&order=market_cap_desc&sparkline=false&price_change_percentage=24h",
		BaseURL, currency, strings.Join(ids, ","))

	bodyBytes, err := cg.get(requestURL)
	if err != nil {
		return nil, err
	}

	var coins []models.Coin
	if err := json.Unmarshal(bodyBytes, &coins); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return coins, nil
}

func (cg *CoinGecko) GetCoinDetail(id string) (models.CoinDetail, error) {
	requestURL := fmt.Sprintf("%s/coins/%s?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false",
		BaseURL, strings.ToLower(strings.TrimSpace(id)))

	bodyBytes, err := cg.get(requestURL)
	if err != nil {
		return models.CoinDetail{}, err
	}

	coinDetail, err := models.UnmarshalCoinDetail(bodyBytes)
	if err != nil {
		return models.CoinDetail{}, fmt.Errorf("decode error: %v", err)
	}

	return coinDetail, nil
}

func (cg *CoinGecko) SearchCoins(query string) (models.SearchResponse, error) {
	requestURL := fmt.Sprintf("%s/search?query=%s", BaseURL, url.QueryEscape(query))

	bodyBytes, err := cg.get(requestURL)
	if err != nil {
		return models.SearchResponse{}, err
	}

	searchResult, err := models.UnmarshalSearch(bodyBytes)
	if err != nil {
		return models.SearchResponse{}, fmt.Errorf("decode error: %v", err)
	}

	return searchResult, nil
}

func (cg *CoinGecko) GetCoinPriceHistory(id, currency string, interval string) ([][]float64, error) {
	selectedInterval := selectInterval(interval)

	requestURL := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=%s&days=%d&interval=%s",
		BaseURL, strings.ToLower(strings.TrimSpace(id)), strings.ToLower(currency), selectedInterval.Days, selectedInterval.Value)

	bodyBytes, err := cg.get(requestURL)
	if err != nil {
		return nil, err
	}

	var result struct {
		Prices [][]float64 `json:"prices"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return result.Prices, nil
}

func (cg *CoinGecko) GetCoinOHLC(id, currency string, interval string) ([]models.OHLC, error) {
	selectedInterval := selectInterval(interval)

	requestURL := fmt.Sprintf("%s/coins/%s/ohlc?vs_currency=%s&days=%d",
		BaseURL, strings.ToLower(strings.TrimSpace(id)), strings.ToLower(currency), selectedInterval.Days)

	bodyBytes, err := cg.get(requestURL)
	if err != nil {
		return nil, err
	}

	var rawData [][]float64
	if err := json.Unmarshal(bodyBytes, &rawData); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	ohlcData := make([]models.OHLC, 0, len(rawData))
	for _, data := range rawData {
		if len(data) >= 5 {
			ohlcData = append(ohlcData, models.OHLC{
				Time:  int64(data[0]),
				Open:  data[1],
				High:  data[2],
				Low:   data[3],
				Close: data[4],
			})
		}
	}

	return ohlcData, nil
}

// SelectInterval returns the interval preset for the given name (defaults to 7d).
func SelectInterval(interval string) Interval {
	return selectInterval(interval)
}

func selectInterval(interval string) Interval {
	for _, i := range Intervals {
		if i.Name == interval {
			return i
		}
	}
	return Intervals[1] // Default to 7d
}

// SelectIntervalForRange picks the smallest preset interval that covers the given span.
func SelectIntervalForRange(from, to time.Time) Interval {
	if to.Before(from) {
		from, to = to, from
	}
	spanDays := int(to.Sub(from).Hours()/24) + 1
	if spanDays < 1 {
		spanDays = 1
	}
	best := Intervals[len(Intervals)-1] // max
	for _, i := range Intervals {
		if i.Days == -1 {
			continue
		}
		if i.Days >= spanDays {
			best = i
			break
		}
	}
	return best
}

// DaysFromNow returns days from t until now (minimum 1).
func DaysFromNow(t time.Time) int {
	days := int(time.Since(t).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

type AlertChecker struct {
	alertManager *models.AlertManager
	coinGecko    *CoinGecko
	stopChan     chan struct{}
	doneChan     chan struct{}
	mu           sync.Mutex
}

func NewAlertChecker(alertManager *models.AlertManager) *AlertChecker {
	return &AlertChecker{
		alertManager: alertManager,
		coinGecko:    NewCoinGecko(),
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

	go func() {
		defer close(doneChan)

		ticker := time.NewTicker(5 * time.Minute)
		rateLimiter := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		defer rateLimiter.Stop()

		ac.runAlertChecks(rateLimiter.C, stopChan)

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				if len(ac.alertManager.GetAlerts()) == 0 {
					return
				}
				ac.runAlertChecks(rateLimiter.C, stopChan)
			}
		}
	}()
}

func (ac *AlertChecker) runAlertChecks(rateLimiter <-chan time.Time, stopChan <-chan struct{}) {
	alerts := ac.alertManager.GetAlerts()
	for _, alert := range alerts {
		select {
		case <-rateLimiter:
			ac.checkAlert(alert)
		case <-stopChan:
			return
		}
	}
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

func (ac *AlertChecker) checkAlert(alert models.Alert) {
	coin, err := ac.coinGecko.GetCoinDetail(alert.CoinID)
	if err != nil {
		fmt.Printf("Error checking alert for %s: %v\n", alert.CoinID, err)
		return
	}

	currency := utils.NormalizeCurrency(alert.Currency)

	currentPrice, err := utils.PriceFromCurrencyMap(coin.MarketData.CurrentPrice, currency)
	if err != nil {
		fmt.Printf("Error checking alert for %s: %v\n", alert.CoinID, err)
		return
	}

	if (alert.Condition == "above" && currentPrice >= alert.Price) ||
		(alert.Condition == "below" && currentPrice <= alert.Price) {
		ac.sendNotification(alert, currentPrice)
		if err := ac.alertManager.RemoveTriggeredAlert(alert); err != nil {
			fmt.Printf("Error removing triggered alert for %s: %v\n", alert.CoinID, err)
		}
	}
}

func (ac *AlertChecker) sendNotification(alert models.Alert, currentPrice float64) {
	currency := utils.NormalizeCurrency(alert.Currency)
	currencySymbol := utils.CurrencySymbol(currency)

	message := fmt.Sprintf("🚨 Price Alert: %s is %s %s%.2f (Target: %s%.2f %s)",
		strings.ToUpper(alert.CoinID),
		alert.Condition,
		currencySymbol,
		currentPrice,
		currencySymbol,
		alert.Price,
		strings.ToUpper(currency))

	fmt.Printf("\n%s\n", message)
}
