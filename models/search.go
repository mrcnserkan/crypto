/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    search, err := UnmarshalSearch(bytes)
//    bytes, err = search.Marshal()

package models

import "encoding/json"

func UnmarshalSearch(data []byte) (Search, error) {
	var r Search
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Search) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Search struct {
	Coins      []CoinSearch  `json:"coins"`
	Exchanges  []Exchange    `json:"exchanges"`
	Icos       []interface{} `json:"icos"`
	Categories []Category    `json:"categories"`
	Nfts       []Nft         `json:"nfts"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CoinSearch struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank int64  `json:"market_cap_rank"`
	Thumb         string `json:"thumb"`
	Large         string `json:"large"`
}

type Exchange struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	MarketType MarketType `json:"market_type"`
	Thumb      string     `json:"thumb"`
	Large      string     `json:"large"`
}

type Nft struct {
	ID     *string `json:"id"`
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	Thumb  string  `json:"thumb"`
}

type MarketType string

const (
	Spot MarketType = "spot"
)
