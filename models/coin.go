/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    coins, err := UnmarshalCoins(bytes)
//    bytes, err = coins.Marshal()

package models

import "encoding/json"

type Coins []Coin

func UnmarshalCoins(data []byte) (Coins, error) {
	var r Coins
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Coins) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Coin struct {
	ID                                 string   `json:"id"`
	Symbol                             string   `json:"symbol"`
	Name                               string   `json:"name"`
	Image                              string   `json:"image"`
	CurrentPrice                       float64  `json:"current_price"`
	MarketCap                          int64    `json:"market_cap"`
	MarketCapRank                      int64    `json:"market_cap_rank"`
	FullyDilutedValuation              *int64   `json:"fully_diluted_valuation"`
	TotalVolume                        int64    `json:"total_volume"`
	High24H                            float64  `json:"high_24h"`
	Low24H                             float64  `json:"low_24h"`
	PriceChange24H                     float64  `json:"price_change_24h"`
	PriceChangePercentage24H           float64  `json:"price_change_percentage_24h"`
	MarketCapChange24H                 float64  `json:"market_cap_change_24h"`
	MarketCapChangePercentage24H       float64  `json:"market_cap_change_percentage_24h"`
	CirculatingSupply                  float64  `json:"circulating_supply"`
	TotalSupply                        *float64 `json:"total_supply"`
	MaxSupply                          *float64 `json:"max_supply"`
	Ath                                float64  `json:"ath"`
	AthChangePercentage                float64  `json:"ath_change_percentage"`
	AthDate                            string   `json:"ath_date"`
	Atl                                float64  `json:"atl"`
	AtlChangePercentage                float64  `json:"atl_change_percentage"`
	AtlDate                            string   `json:"atl_date"`
	Roi                                *Roi     `json:"roi"`
	LastUpdated                        string   `json:"last_updated"`
	PriceChangePercentage24HInCurrency float64  `json:"price_change_percentage_24h_in_currency"`
	PriceChangePercentage7DInCurrency  float64  `json:"price_change_percentage_7d_in_currency"`
}

type Roi struct {
	Times      float64 `json:"times"`
	Currency   string  `json:"currency"`
	Percentage float64 `json:"percentage"`
}
