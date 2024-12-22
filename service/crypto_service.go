/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mrcnserkan/crypto/models"
)

const (
	BaseURL                 = "https://api.coingecko.com/api/v3"
	DEFAULT_PAGE            = "1"
	DEFAULT_CURRENCY        = "usd"
	DEFAULT_CURRENCY_SYMBOL = "$"
	PER_PAGE                = "10"
)

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

func (cg *CoinGecko) GetMarkets(currency string, perPage int, page int) ([]models.Coin, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=%s&order=market_cap_desc&per_page=%d&page=%d&sparkline=false&price_change_percentage=24h,7d",
		BaseURL, strings.ToLower(currency), perPage, page)

	resp, err := cg.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
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
	url := fmt.Sprintf("%s/coins/%s?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false",
		BaseURL, strings.ToLower(id))

	resp, err := cg.client.Get(url)
	if err != nil {
		return models.CoinDetail{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.CoinDetail{}, fmt.Errorf("API error: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
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
	url := fmt.Sprintf("%s/search?query=%s", BaseURL, query)

	resp, err := cg.client.Get(url)
	if err != nil {
		return models.SearchResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.SearchResponse{}, fmt.Errorf("API error: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
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
	var selectedInterval Interval
	for _, i := range Intervals {
		if i.Name == interval {
			selectedInterval = i
			break
		}
	}
	if selectedInterval.Name == "" {
		selectedInterval = Intervals[1] // Default to 7d
	}

	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=%s&days=%d&interval=%s",
		BaseURL, strings.ToLower(id), strings.ToLower(currency), selectedInterval.Days, selectedInterval.Value)

	resp, err := cg.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result struct {
		Prices [][]float64 `json:"prices"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return result.Prices, nil
}

func (cg *CoinGecko) GetCoinOHLC(id, currency string, interval string) ([]models.OHLC, error) {
	var selectedInterval Interval
	for _, i := range Intervals {
		if i.Name == interval {
			selectedInterval = i
			break
		}
	}
	if selectedInterval.Name == "" {
		selectedInterval = Intervals[1] // Default to 7d
	}

	url := fmt.Sprintf("%s/coins/%s/ohlc?vs_currency=%s&days=%d",
		BaseURL, strings.ToLower(id), strings.ToLower(currency), selectedInterval.Days)

	resp, err := cg.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var rawData [][]float64
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &rawData); err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	ohlcData := make([]models.OHLC, len(rawData))
	for i, data := range rawData {
		if len(data) >= 5 {
			ohlcData[i] = models.OHLC{
				Time:  int64(data[0]),
				Open:  data[1],
				High:  data[2],
				Low:   data[3],
				Close: data[4],
			}
		}
	}

	return ohlcData, nil
}
