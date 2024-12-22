/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package models

import "encoding/json"

type Coin struct {
	ID                                 string         `json:"id"`
	Symbol                             string         `json:"symbol"`
	Name                               string         `json:"name"`
	Image                              string         `json:"image"`
	CurrentPrice                       float64        `json:"current_price"`
	MarketCap                          float64        `json:"market_cap"`
	MarketCapRank                      int            `json:"market_cap_rank"`
	TotalVolume                        float64        `json:"total_volume"`
	High24h                            float64        `json:"high_24h"`
	Low24h                             float64        `json:"low_24h"`
	PriceChange24h                     float64        `json:"price_change_24h"`
	PriceChangePercentage24h           float64        `json:"price_change_percentage_24h"`
	MarketCapChange24h                 float64        `json:"market_cap_change_24h"`
	MarketCapChangePercentage24h       float64        `json:"market_cap_change_percentage_24h"`
	CirculatingSupply                  float64        `json:"circulating_supply"`
	TotalSupply                        float64        `json:"total_supply"`
	MaxSupply                          float64        `json:"max_supply"`
	Ath                                float64        `json:"ath"`
	AthChangePercentage                float64        `json:"ath_change_percentage"`
	AthDate                            string         `json:"ath_date"`
	Atl                                float64        `json:"atl"`
	AtlChangePercentage                float64        `json:"atl_change_percentage"`
	AtlDate                            string         `json:"atl_date"`
	LastUpdated                        string         `json:"last_updated"`
	SparklineIn7d                      *SparklineIn7d `json:"sparkline_in_7d,omitempty"`
	PriceChangePercentage1hInCurrency  float64        `json:"price_change_percentage_1h_in_currency"`
	PriceChangePercentage24HInCurrency float64        `json:"price_change_percentage_24h_in_currency"`
	PriceChangePercentage7DInCurrency  float64        `json:"price_change_percentage_7d_in_currency"`
}

type SparklineIn7d struct {
	Price []float64 `json:"price"`
}

type Category struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	MarketCap float64 `json:"market_cap"`
	Volume24h float64 `json:"volume_24h"`
	Change24h float64 `json:"price_change_24h"`
}

type SearchResponse struct {
	Coins      []CoinSearch `json:"coins"`
	Categories []Category   `json:"categories"`
	Nfts       []Nft        `json:"nfts"`
}

type CoinSearch struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank int    `json:"market_cap_rank"`
	Thumb         string `json:"thumb"`
	Large         string `json:"large"`
}

type Nft struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Thumb  string `json:"thumb"`
}

type OHLC struct {
	Time  int64   `json:"time"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

func UnmarshalSearch(data []byte) (SearchResponse, error) {
	var r SearchResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SearchResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
