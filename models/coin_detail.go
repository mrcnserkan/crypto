/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    coinDetail, err := UnmarshalCoinDetail(bytes)
//    bytes, err = coinDetail.Marshal()

package models

import "encoding/json"

func UnmarshalCoinDetail(data []byte) (CoinDetail, error) {
	var r CoinDetail
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *CoinDetail) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type CoinDetail struct {
	ID                           string              `json:"id"`
	Symbol                       string              `json:"symbol"`
	Name                         string              `json:"name"`
	AssetPlatformID              interface{}         `json:"asset_platform_id"`
	Platforms                    Platforms           `json:"platforms"`
	BlockTimeInMinutes           int64               `json:"block_time_in_minutes"`
	HashingAlgorithm             string              `json:"hashing_algorithm"`
	Categories                   []string            `json:"categories"`
	PublicNotice                 interface{}         `json:"public_notice"`
	AdditionalNotices            []interface{}       `json:"additional_notices"`
	Description                  Description         `json:"description"`
	Links                        Links               `json:"links"`
	Image                        Image               `json:"image"`
	CountryOrigin                string              `json:"country_origin"`
	GenesisDate                  string              `json:"genesis_date"`
	SentimentVotesUpPercentage   float64             `json:"sentiment_votes_up_percentage"`
	SentimentVotesDownPercentage float64             `json:"sentiment_votes_down_percentage"`
	MarketCapRank                int64               `json:"market_cap_rank"`
	CoingeckoRank                int64               `json:"coingecko_rank"`
	CoingeckoScore               float64             `json:"coingecko_score"`
	DeveloperScore               float64             `json:"developer_score"`
	CommunityScore               float64             `json:"community_score"`
	LiquidityScore               float64             `json:"liquidity_score"`
	PublicInterestScore          float64             `json:"public_interest_score"`
	MarketData                   MarketData          `json:"market_data"`
	PublicInterestStats          PublicInterestStats `json:"public_interest_stats"`
	StatusUpdates                []interface{}       `json:"status_updates"`
	LastUpdated                  string              `json:"last_updated"`
}

type Description struct {
	En string `json:"en"`
}

type Image struct {
	Thumb string `json:"thumb"`
	Small string `json:"small"`
	Large string `json:"large"`
}

type Links struct {
	Homepage                    []string    `json:"homepage"`
	BlockchainSite              []string    `json:"blockchain_site"`
	OfficialForumURL            []string    `json:"official_forum_url"`
	ChatURL                     []string    `json:"chat_url"`
	AnnouncementURL             []string    `json:"announcement_url"`
	TwitterScreenName           string      `json:"twitter_screen_name"`
	FacebookUsername            string      `json:"facebook_username"`
	BitcointalkThreadIdentifier interface{} `json:"bitcointalk_thread_identifier"`
	TelegramChannelIdentifier   string      `json:"telegram_channel_identifier"`
	SubredditURL                string      `json:"subreddit_url"`
	ReposURL                    ReposURL    `json:"repos_url"`
}

type ReposURL struct {
	Github    []string      `json:"github"`
	Bitbucket []interface{} `json:"bitbucket"`
}

type MarketData struct {
	CurrentPrice                           map[string]float64 `json:"current_price"`
	TotalValueLocked                       interface{}        `json:"total_value_locked"`
	McapToTvlRatio                         interface{}        `json:"mcap_to_tvl_ratio"`
	FdvToTvlRatio                          interface{}        `json:"fdv_to_tvl_ratio"`
	Roi                                    interface{}        `json:"roi"`
	Ath                                    map[string]float64 `json:"ath"`
	AthChangePercentage                    map[string]float64 `json:"ath_change_percentage"`
	AthDate                                map[string]string  `json:"ath_date"`
	Atl                                    map[string]float64 `json:"atl"`
	AtlChangePercentage                    map[string]float64 `json:"atl_change_percentage"`
	AtlDate                                map[string]string  `json:"atl_date"`
	MarketCap                              map[string]int64   `json:"market_cap"`
	MarketCapRank                          int64              `json:"market_cap_rank"`
	FullyDilutedValuation                  map[string]float64 `json:"fully_diluted_valuation"`
	TotalVolume                            map[string]float64 `json:"total_volume"`
	High24H                                map[string]float64 `json:"high_24h"`
	Low24H                                 map[string]float64 `json:"low_24h"`
	PriceChange24H                         float64            `json:"price_change_24h"`
	PriceChangePercentage24H               float64            `json:"price_change_percentage_24h"`
	PriceChangePercentage7D                float64            `json:"price_change_percentage_7d"`
	PriceChangePercentage14D               float64            `json:"price_change_percentage_14d"`
	PriceChangePercentage30D               float64            `json:"price_change_percentage_30d"`
	PriceChangePercentage60D               float64            `json:"price_change_percentage_60d"`
	PriceChangePercentage200D              float64            `json:"price_change_percentage_200d"`
	PriceChangePercentage1Y                float64            `json:"price_change_percentage_1y"`
	MarketCapChange24H                     float64            `json:"market_cap_change_24h"`
	MarketCapChangePercentage24H           float64            `json:"market_cap_change_percentage_24h"`
	PriceChange24HInCurrency               map[string]float64 `json:"price_change_24h_in_currency"`
	PriceChangePercentage1HInCurrency      map[string]float64 `json:"price_change_percentage_1h_in_currency"`
	PriceChangePercentage24HInCurrency     map[string]float64 `json:"price_change_percentage_24h_in_currency"`
	PriceChangePercentage7DInCurrency      map[string]float64 `json:"price_change_percentage_7d_in_currency"`
	PriceChangePercentage14DInCurrency     map[string]float64 `json:"price_change_percentage_14d_in_currency"`
	PriceChangePercentage30DInCurrency     map[string]float64 `json:"price_change_percentage_30d_in_currency"`
	PriceChangePercentage60DInCurrency     map[string]float64 `json:"price_change_percentage_60d_in_currency"`
	PriceChangePercentage200DInCurrency    map[string]float64 `json:"price_change_percentage_200d_in_currency"`
	PriceChangePercentage1YInCurrency      map[string]float64 `json:"price_change_percentage_1y_in_currency"`
	MarketCapChange24HInCurrency           map[string]float64 `json:"market_cap_change_24h_in_currency"`
	MarketCapChangePercentage24HInCurrency map[string]float64 `json:"market_cap_change_percentage_24h_in_currency"`
	TotalSupply                            float64            `json:"total_supply"`
	MaxSupply                              float64            `json:"max_supply"`
	CirculatingSupply                      float64            `json:"circulating_supply"`
	LastUpdated                            string             `json:"last_updated"`
}

type Platforms struct {
	Empty string `json:""`
}

type PublicInterestStats struct {
	AlexaRank   int64       `json:"alexa_rank"`
	BingMatches interface{} `json:"bing_matches"`
}
