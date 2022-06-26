/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mrcnserkan/crypto/models"
	"github.com/mrcnserkan/crypto/utils"
)

const DEFAULT_PAGE string = "1"
const DEFAULT_CURRENCY string = "usd"
const DEFAULT_CURRENCY_SYMBOL string = "$"
const PER_PAGE string = "10"
const LIST_URL string = "https://api.coingecko.com/api/v3/coins/markets?vs_currency=%s&per_page=%s&page=%s&order=market_cap_desc&price_change_percentage=24h%%2C7d"
const COIN_DETAIL_URL string = "https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&community_data=false&developer_data=false"
const SEARCH_URL string = "https://api.coingecko.com/api/v3/search?query=%s"

func GetList(page string, perPage string, currency string) models.Coins {
	if page == "" {
		page = DEFAULT_PAGE
	}
	if perPage == "" {
		perPage = PER_PAGE
	}
	if currency == "" {
		currency = DEFAULT_CURRENCY
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Get(fmt.Sprintf(LIST_URL, currency, perPage, page))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		coins, err := models.UnmarshalCoins(bodyBytes)
		if err != nil {
			log.Fatal(err)
		}
		return coins
	}
	return nil
}

func GetDetail(coinName string) models.CoinDetail {
	if coinName == "" {
		coinName = "bitcoin"
	}
	coinName = utils.ClearCoinName(coinName)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Get(fmt.Sprintf(COIN_DETAIL_URL, coinName))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		coinDetail, err := models.UnmarshalCoinDetail(bodyBytes)
		if err != nil {
			log.Fatal(err)
		}
		return coinDetail
	}
	return models.CoinDetail{}
}

func Search(query string) models.Search {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Get(fmt.Sprintf(SEARCH_URL, query))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		searchResult, err := models.UnmarshalSearch(bodyBytes)
		if err != nil {
			log.Fatal(err)
		}
		return searchResult
	}
	return models.Search{}
}
