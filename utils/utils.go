/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	tableWriter "github.com/olekukonko/tablewriter"
)

const (
	MaxPerPage           = 250
	MaxMarketIDsPerBatch = 250
	HoldingDustThreshold = 1e-10
)

func GetCellColorFromPriceChange(change float64) tableWriter.Colors {
	if change < 0 {
		return tableWriter.Colors{tableWriter.Normal, tableWriter.FgRedColor}
	}
	return tableWriter.Colors{tableWriter.Normal, tableWriter.FgGreenColor}
}

func NormalizeCoinID(coinID string) string {
	return strings.ToLower(strings.TrimSpace(coinID))
}

func NormalizeCurrency(currency string) string {
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		return "usd"
	}
	return currency
}

func IsEffectivelyZero(value float64) bool {
	return math.Abs(value) < HoldingDustThreshold
}

func PriceFromCurrencyMap(prices map[string]float64, currency string) (float64, error) {
	currency = NormalizeCurrency(currency)
	price, ok := prices[currency]
	if !ok {
		return 0, fmt.Errorf("unsupported or invalid currency: %s", strings.ToUpper(currency))
	}
	return price, nil
}

func FloatFromCurrencyMap(values map[string]float64, currency string) (float64, error) {
	return PriceFromCurrencyMap(values, currency)
}

func Int64FromCurrencyMap(values map[string]int64, currency string) (int64, error) {
	currency = NormalizeCurrency(currency)
	value, ok := values[currency]
	if !ok {
		return 0, fmt.Errorf("unsupported or invalid currency: %s", strings.ToUpper(currency))
	}
	return value, nil
}

func ClearCoinName(coinName string) string {
	blankRegex := regexp.MustCompile(`[. ]`)
	specialCharRegex := regexp.MustCompile(`[(\[\])]`)
	return strings.ToLower(
		specialCharRegex.ReplaceAllString(
			blankRegex.ReplaceAllString(
				coinName, "-",
			), "",
		),
	)
}

func GetCaption() string {
	return fmt.Sprintf("Data source from coingecko.com at %s", GetCurrentTime())
}

func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func ParsePage(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil || i < 1 {
		return 0, fmt.Errorf("invalid page: %s (must be a positive integer)", s)
	}
	return i, nil
}

func ParsePerPage(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil || i < 1 || i > MaxPerPage {
		return 0, fmt.Errorf("invalid per-page: %s (must be between 1 and %d)", s, MaxPerPage)
	}
	return i, nil
}

func CurrencySymbol(currency string) string {
	if currency == "" || currency == "usd" {
		return "$"
	}
	return strings.ToUpper(currency)
}

func FormatISODate(iso string) string {
	if len(iso) >= 10 {
		return iso[:10]
	}
	return iso
}

func FormatCurrency(value float64) string {
	if value >= 1_000_000_000_000 { // Trillion
		return fmt.Sprintf("%.2fT", value/1_000_000_000_000)
	} else if value >= 1_000_000_000 { // Billion
		return fmt.Sprintf("%.2fB", value/1_000_000_000)
	} else if value >= 1_000_000 { // Million
		return fmt.Sprintf("%.2fM", value/1_000_000)
	} else if value >= 1_000 { // Thousand
		return fmt.Sprintf("%.2fK", value/1_000)
	}
	return fmt.Sprintf("%.2f", value)
}

func MinMax(array []float64) (float64, float64) {
	if len(array) == 0 {
		return 0, 0
	}
	min := array[0]
	max := array[0]
	for _, value := range array {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

